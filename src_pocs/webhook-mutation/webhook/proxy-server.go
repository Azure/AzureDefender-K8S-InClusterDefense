package main

import (
	"encoding/json"
	"os"

	// "io/ioutil"
	//"net/http"

	// "path"
	// "regexp"
	// "strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	// yaml "gopkg.in/yaml.v2"
	// azsecurity "github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"
	azresourcegraph "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2019-04-01/resourcegraph"

	//acrmgmt "github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/mgmt/2018-02-01/containerregistry"
	//acr "github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/runtime/2019-08-15-preview/containerregistry"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"

	"github.com/pkg/errors"

	"time"
)

// auth
const (
	// OAuthGrantTypeServicePrincipal for client credentials flow
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	cloudName                      string         = ""
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

// OAuthGrantType specifies which grant type to use.
type OAuthGrantType int

// AuthGrantType ...
func AuthGrantType() OAuthGrantType {
	return OAuthGrantTypeServicePrincipal
}

// Server
type Server struct {
	// subscriptionId to azure
	SubscriptionID string
	// resource-group
	ResourceGroup string
	// Identity ClientID
	IdentityClientID string
}

// ScanInfo
type ScanInfo struct {
	//ScanStatus
	ImageDigest *string `json:"imageDigest,omitempty"`
	//ScanStatus
	ScanStatus *string `json:"scanStatus,omitempty"`
	// SeveritySummary
	SeveritySummary map[string]int `json:"severitySummary,omitempty"`
}

// NewServer creates a new server instance.
func NewServer() (*Server, error) {
	log.Debugf("NewServer")
	var s Server
	s.SubscriptionID = os.Getenv("SUBSCRIPTION_ID")
	s.ResourceGroup = os.Getenv("RESOURCE_GROUP")
	s.IdentityClientID = os.Getenv("CLIENT_ID")

	return &s, nil
}

// Process the image
func (s *Server) Process(ctx context.Context, image Image) (resps []ScanInfo, err error) {
	// Create ARG client
	argClient, err := s.createARGClient()
	if err != nil {
		return nil, err
	}
	// Generate Query
	query := s.generateQuery(image.digest)
	// Execute Query
	results, err := argClient.Resources(ctx, query)
	if err != nil {
		return nil, err
	}
	//Parse query response:
	resps, err2 := s.parseQueryResponse(results)
	if err2 != nil {
		log.Debug(err2.Error())
		return nil, err2
	}

	return resps, nil
}

// Parse query response from arg to ScanInfo array.
func (s *Server) parseQueryResponse(results azresourcegraph.QueryResponse) (scanInfoList []ScanInfo, err error) {
	log.Debugf("results: %d", results)

	var data []ScanInfo
	count := *results.Count

	scanInfoList = make([]ScanInfo, 0)
	// In case that scan info returned from ARG.
	if count > 0 {
		raw, err := json.Marshal(results.Data)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(raw, &data)
		if err != nil {
			return nil, err
		}
		log.Debugf("Data: %d", data)
		for _, v := range data {
			oneScanInfo := ScanInfo{
				ScanStatus:      v.ScanStatus,
				SeveritySummary: v.SeveritySummary,
			}
			scanInfoList = append(scanInfoList, oneScanInfo)
		}
		// In case that there are no results of scan.
	} else {
		unScannedImageForAssignment := unscannedImage // Can't assign pointer of constant.
		oneScanInfo := ScanInfo{
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
func (s *Server) generateQuery(digest string) azresourcegraph.QueryRequest {
	// Prepare query:
	subs := []string{s.SubscriptionID}
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
		| summarize scanFindingSeverityCount = count() by scanFindingSeverity, scanStatus, repository, imageDigest
		| summarize severitySummary = make_bag(pack(scanFindingSeverity, scanFindingSeverityCount)) by  imageDigest, scanStatus`

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
func (s *Server) createARGClient() (azresourcegraph.BaseClient, error) {
	// Connect to ARG:
	argClient := azresourcegraph.New()
	token, tokenErr := s.generateToken()
	if tokenErr != nil {
		return argClient, errors.Wrapf(tokenErr, "failed to get management token")
	}
	// [4. creates a BearerAuthorizer using the given token provider]
	argClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return argClient, nil
}

// Generate endpoint and token for Bearer Authorizer using UserAssignedIdentity
func (s *Server) generateToken() (msg *adal.Token, err error) {
	// [1. EndPoint] Get the MSI endpoint on Virtual Machines
	imdsTokenEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		log.Errorf("failed to get IMDS token endpoint, error: %+v", err)
	}
	// Get token using userAssignedID
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		token := s.getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint)
		if token == nil {
			log.Error("Tokens acquired from IMDS with and without identity client ID do not match")
		} else {
			return token, nil
		}
	}
	return nil, nil
}

// Get token of user assigned identity
func (s *Server) getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint string) *adal.Token {
	// [2. SPT] creates a ServicePrincipalToken via the MSI VM Extension by using the clientID of specified user assigned identity
	spt, err := adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(imdsTokenEndpoint, resourceName, s.IdentityClientID)
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

	log.Infof("successfully acquired a service principal token from %s using a user-assigned identity (%s)", imdsTokenEndpoint, s.IdentityClientID)
	return &token
}
