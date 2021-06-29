package main

import (
	"encoding/json"
	"os"
	"time"

	azresourcegraph "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2019-04-01/resourcegraph"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	timeout = 80 * time.Second
)

const (
	unscannedImage string = "Unscanned"
)

var (
	period       time.Duration = 100 * time.Second
	resourceName string        = "https://management.azure.com/"
)

// Server for arg client.
type ARGProxy struct {
	ctx context.Context
	// subscriptionId to azure
	SubscriptionID string
	// resource-group
	ResourceGroup string
	// Identity ClientID
	IdentityClientID string
}

type QueryResponse struct {
	//ScanStatus
	ScanStatus *string `json:"scanStatus,omitempty"`
	// SeveritySummary
	SeveritySummary map[string]int `json:"severitySummary,omitempty"`
}

// NewServer creates a new server instance.
func NewARGProxy() (ARGProxy, error) {
	log.Debugf("Creating ARG proxy")
	var proxy ARGProxy
	proxy.ctx = context.Background()
	proxy.SubscriptionID = os.Getenv("SUBSCRIPTION_ID")
	proxy.ResourceGroup = os.Getenv("RESOURCE_GROUP")
	proxy.IdentityClientID = os.Getenv("CLIENT_ID")

	return proxy, nil
}

// Process the image
func (proxy ARGProxy) GetImageSecInfo(image Image) (imageSecInfo *ImageSecInfo, err error) {
	// Create ARG client
	argClient, err := proxy.createARGClient()
	if err != nil {
		return nil, err
	}
	// Generate Query
	query := proxy.generateQuery(image.Digest)
	// Execute Query
	results, err := argClient.Resources(proxy.ctx, query)
	if err != nil {
		return nil, err
	}
	//Parse query response:
	scanInfoList, err := proxy.parseQueryResponse(results)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	log.Debugf("Scan Info List: %s", scanInfoList)
	// Extract first item in the list.
	return &scanInfoList[0], nil
}

// Parse query response from arg to ScanInfo array.
func (proxy ARGProxy) parseQueryResponse(results azresourcegraph.QueryResponse) (scanInfoList []ImageSecInfo, err error) {
	var data []QueryResponse
	count := *results.Count
	log.Debugf("results.Data: %s", results.Data)
	scanInfoList = make([]ImageSecInfo, 0)
	// In case that scan info returned from ARG.
	if count > 0 {
		raw, err := json.Marshal(results.Data)
		if err != nil {
			return nil, err
		}
		log.Debugf("raw: %s", raw)
		err = json.Unmarshal(raw, &data)
		if err != nil {
			return nil, err
		}
		log.Debugf("data: %s", data)
		for _, v := range data {
			oneScanInfo := ImageSecInfo{
				ScanStatus:      v.ScanStatus,
				SeveritySummary: v.SeveritySummary,
			}
			scanInfoList = append(scanInfoList, oneScanInfo)
		}
		// In case that there are no results of scan.
	} else {
		unScannedImageForAssignment := unscannedImage // Can't assign pointer of constant.
		oneScanInfo := ImageSecInfo{
			ScanStatus:      &unScannedImageForAssignment,
			SeveritySummary: nil,
		}
		scanInfoList = append(scanInfoList, oneScanInfo)
	}

	log.Debugf("total unhealthy images: %d", count)
	log.Debugf("scanInfoList: %s", scanInfoList)

	return scanInfoList, nil
}

// Generate query for ARG with the relevant options and digest.
func (proxy ARGProxy) generateQuery(digest string) azresourcegraph.QueryRequest {
	// Prepare query:
	subs := []string{proxy.SubscriptionID}
	rawQuery := `
		securityresources
		| where type == 'microsoft.security/assessments/subassessments'
		| where id matches regex '(.+?)/providers/Microsoft.Security/assessments/dbd0cb49-b563-45e7-9724-889e799fa648/'
		//| parse id with registryResourceId '/providers/Microsoft.Security/assessments/' *
		//| parse registryResourceId with * "/providers/Microsoft.ContainerRegistry/registries/" registryName
		| extend imageDigest = tostring(properties.additionalData.imageDigest)
		| where imageDigest == '` + digest + `'
		| extend repository = tostring(properties.additionalData.repositoryName)
		| extend scanFindingSeverity = tostring(properties.status.severity), scanStatus = tostring(properties.status.code)
		| summarize scanFindingSeverityCount = count() by scanFindingSeverity, scanStatus, repository
		| summarize severitySummary = make_bag(pack(scanFindingSeverity, scanFindingSeverityCount)) by scanStatus`

	log.Debugf("Query: %s", rawQuery)
	options := azresourcegraph.QueryRequestOptions{ResultFormat: azresourcegraph.ResultFormatObjectArray}
	query := azresourcegraph.QueryRequest{
		Subscriptions: &subs,
		Query:         &rawQuery,
		Options:       &options,
	}
	return query
}

// Create ARG Client using userassigned identity.
func (proxy ARGProxy) createARGClient() (azresourcegraph.BaseClient, error) {
	// Connect to ARG:
	argClient := azresourcegraph.New()
	token, tokenErr := proxy.generateToken()
	if tokenErr != nil {
		return argClient, errors.Wrapf(tokenErr, "failed to get management token")
	}
	// [4. creates a BearerAuthorizer using the given token provider]
	argClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return argClient, nil
}

// Generate endpoint and token for Bearer Authorizer using UserAssignedIdentity
func (proxy ARGProxy) generateToken() (msg *adal.Token, err error) {
	// [1. EndPoint] Get the MSI endpoint on Virtual Machines
	imdsTokenEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		log.Errorf("failed to get IMDS token endpoint, error: %+v", err)
	}
	// Get token using userAssignedID
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		token := proxy.getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint)
		if token == nil {
			log.Error("Tokens acquired from IMDS with and without identity client ID do not match")
		} else {
			return token, nil
		}
	}
	return nil, nil
}

// Get token of user assigned identity
func (proxy ARGProxy) getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint string) *adal.Token {
	// [2. SPT] creates a ServicePrincipalToken via the MSI VM Extension by using the clientID of specified user assigned identity
	spt, err := adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(imdsTokenEndpoint, resourceName, proxy.IdentityClientID)
	if err != nil {
		log.Errorf("failed to acquire a token from IMDS using user-assigned identity, error: %+v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := spt.RefreshWithContext(ctx); err != nil {
		log.Errorf("failed to refresh the service principal token, error: %+v", err)
		return nil
	}
	// [3. Token] Extract token from Service principal token
	token := spt.Token()
	if token.IsZero() {
		log.Errorf("%+v is a zero token", token)
		return nil
	}

	log.Infof("successfully acquired a service principal token from %s using a user-assigned identity (%s)", imdsTokenEndpoint, proxy.IdentityClientID)
	return &token
}
