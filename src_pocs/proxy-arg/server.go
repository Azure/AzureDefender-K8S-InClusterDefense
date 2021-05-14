package main

import (
	"encoding/json"
	"fmt"

	// "io/ioutil"
	//"net/http"

	// "path"
	// "regexp"
	// "strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	// yaml "gopkg.in/yaml.v2"
	azsecurity "github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"
	azresourcegraph "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2019-04-01/resourcegraph"
	"github.com/Azure/go-autorest/autorest/date"

	//acrmgmt "github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/mgmt/2018-02-01/containerregistry"
	//acr "github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/runtime/2019-08-15-preview/containerregistry"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/pkg/errors"

	"flag"
	"net/http"
	"strings"
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

var (
	period           time.Duration
	resourceName     string
	subscriptionID   string
	resourceGroup    string
	identityClientID string
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
	// tenantID in AAD
	TenantID string
	// AAD app client secret (if not using POD AAD Identity)
	AADClientSecret string
	// AAD app client secret id (if not using POD AAD Identity)
	AADClientID string
	// Location of security center
	Location string
	// Scope of assessment
	Scope string
}

// Response
type Response struct {
	// ID - Vulnerability ID
	ID *string `json:"id,omitempty"`
	// DisplayName - User friendly display name of the sub-assessment
	DisplayName *string                         `json:"displayName,omitempty"`
	Status      *azsecurity.SubAssessmentStatus `json:"status,omitempty"`
	// Remediation - Information on how to remediate this sub-assessment
	Remediation *string `json:"remediation,omitempty"`
	// Impact - Description of the impact of this sub-assessment
	Impact *string `json:"impact,omitempty"`
	// Category - Category of the sub-assessment
	Category *string `json:"category,omitempty"`
	// Description - Human readable description of the assessment status
	Description *string `json:"description,omitempty"`
	// TimeGenerated - The date and time the sub-assessment was generated
	TimeGenerated   *date.Time                                       `json:"timeGenerated,omitempty"`
	ResourceDetails azsecurity.AzureResourceDetails                  `json:"resourceDetails,omitempty"`
	AdditionalData  ResponseContainerRegistryVulnerabilityProperties `json:"additionalData,omitempty"`
}

// ResponseContainerRegistryVulnerabilityProperties additional context fields for container registry Vulnerability
// assessment
type ResponseContainerRegistryVulnerabilityProperties struct {
	// Type - READ-ONLY; Vulnerability Type. e.g: Vulnerability, Potential Vulnerability, Information Gathered, Vulnerability
	Type *string `json:"type,omitempty"`
	// Cvss - READ-ONLY; Dictionary from cvss version to cvss details object
	Cvss map[string]*azsecurity.CVSS `json:"cvss"`
	// Patchable - READ-ONLY; Indicates whether a patch is available or not
	Patchable *bool `json:"patchable,omitempty"`
	// Cve - READ-ONLY; List of CVEs
	Cve *[]azsecurity.CVE `json:"cve,omitempty"`
	// PublishedTime - READ-ONLY; Published time
	PublishedTime *date.Time `json:"publishedTime,omitempty"`
	// VendorReferences - READ-ONLY
	VendorReferences *[]azsecurity.VendorReference `json:"vendorReferences,omitempty"`
	// RepositoryName - READ-ONLY; Name of the repository which the vulnerable image belongs to
	RepositoryName *string `json:"repositoryName,omitempty"`
	// ImageDigest - READ-ONLY; Digest of the vulnerable image
	ImageDigest *string `json:"imageDigest,omitempty"`
	// AssessedResourceType - Possible values include: 'AssessedResourceTypeAdditionalData', 'AssessedResourceTypeSQLServerVulnerability', 'AssessedResourceTypeContainerRegistryVulnerability', 'AssessedResourceTypeServerVulnerabilityAssessment'
	AssessedResourceType azsecurity.AssessedResourceType `json:"assessedResourceType,omitempty"`
}

// NewServer creates a new server instance.
func NewServer() (*Server, error) {
	log.Debugf("NewServer")
	var s Server
	s.SubscriptionID = "409111bf-3097-421c-ad68-a44e716edf58" // os.Getenv("SUBSCRIPTION_ID")
	// s.AADClientID = os.Getenv("CLIENT_ID")
	// s.AADClientSecret = os.Getenv("CLIENT_SECRET")
	// s.TenantID = os.Getenv("TENANT_ID")

	// if s.SubscriptionID == "" {
	// 	return nil, fmt.Errorf("could not find SUBSCRIPTION_ID")
	// }
	// if s.AADClientID == "" {
	// 	return nil, fmt.Errorf("could not find CLIENT_ID")
	// }
	// if s.AADClientSecret == "" {
	// 	return nil, fmt.Errorf("could not find CLIENT_SECRET")
	// }
	// if s.TenantID == "" {
	// 	return nil, fmt.Errorf("could not find TENANT_ID")
	// }

	return &s, nil
}

// ParseAzureEnvironment returns azure environment by name
func ParseAzureEnvironment(cloudName string) (*azure.Environment, error) {
	var env azure.Environment
	var err error
	if cloudName == "" {
		env = azure.PublicCloud
	} else {
		env, err = azure.EnvironmentFromName(cloudName)
	}
	return &env, err
}

func (s *Server) Process(ctx context.Context, digest string) (resps []Response, err error) {
	if digest == "" {
		return nil, fmt.Errorf("Failed to provide digest to query")
	}
	//TODO fix getimage.sh script
	digest = "sha256:f68b2ce26292e02ef3dbc6cfae17cec3d54b9afee9bf5298bf95d005c0783143"
	myClient := azresourcegraph.New()
	// token, tokenErr := s.GetManagementToken(AuthGrantType(), cloudName)
	token, tokenErr := getTokenMSI()
	if tokenErr != nil {
		return nil, errors.Wrapf(tokenErr, "failed to get management token")
	}

	myClient.Authorizer = autorest.NewBearerAuthorizer(token)
	subs := []string{s.SubscriptionID}
	rawQuery := `
	securityresources | where type == "microsoft.security/assessments/subassessments" 
	| extend resourceType = tostring(properties["additionalData"].assessedResourceType) 
	| extend status = tostring(properties["status"].code)
	| where resourceType == "ContainerRegistryVulnerability" 
	| extend repoName = tostring(properties["additionalData"].repositoryName) 
	| extend imageSha = tostring(properties["additionalData"].imageDigest)
	| where status == "Unhealthy"
	| where imageSha == "` + digest + `"`

	options := azresourcegraph.QueryRequestOptions{
		ResultFormat: azresourcegraph.ResultFormatObjectArray,
	}
	query := azresourcegraph.QueryRequest{
		Subscriptions: &subs,
		Query:         &rawQuery,
		Options:       &options,
	}
	results, err := myClient.Resources(ctx, query)
	if err != nil {
		return nil, err
	}
	var data []azsecurity.SubAssessment
	count := *results.Count
	resps = make([]Response, 0)
	if count > 0 {
		raw, err := json.Marshal(results.Data)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(raw, &data)
		if err != nil {
			return nil, err
		}

		for _, v := range data {
			rd, _ := v.SubAssessmentProperties.ResourceDetails.AsAzureResourceDetails()
			ad, _ := v.SubAssessmentProperties.AdditionalData.AsContainerRegistryVulnerabilityProperties()
			resp := Response{
				ID:              v.ID,
				DisplayName:     v.DisplayName,
				Status:          v.Status,
				Remediation:     v.Remediation,
				Impact:          v.Impact,
				Category:        v.Category,
				Description:     v.Description,
				TimeGenerated:   v.TimeGenerated,
				ResourceDetails: *rd,
				AdditionalData: ResponseContainerRegistryVulnerabilityProperties{
					AssessedResourceType: ad.AssessedResourceType,
					RepositoryName:       ad.RepositoryName,
					Type:                 ad.Type,
					Cvss:                 ad.Cvss,
					Patchable:            ad.Patchable,
					Cve:                  ad.Cve,
					PublishedTime:        ad.PublishedTime,
					VendorReferences:     ad.VendorReferences,
					ImageDigest:          ad.ImageDigest,
				},
			}
			resps = append(resps, resp)
		}
	}

	log.Debugf("total unhealthy images: %d", count)

	return resps, nil
}

func getTokenMSI() (msg *adal.Token, err error) {
	flag.DurationVar(&period, "period", 100*time.Second, "The period that the demo is being executed")
	flag.StringVar(&resourceName, "resource-name", "https://management.azure.com/", "The resource name to grant the access token")
	flag.StringVar(&subscriptionID, "subscription-id", "", "The Azure subscription ID")
	flag.StringVar(&resourceGroup, "resource-group", "", "The resource group name which the user-assigned identity read access to")
	flag.StringVar(&identityClientID, "identity-client-id", "", "The user-assigned identity client ID")
	flag.Parse()

	imdsTokenEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		log.Errorf("failed to get IMDS token endpoint, error: %+v", err)
	}

	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		curlIMDSMetadataInstanceEndpoint()
		t1 := getTokenFromIMDS(imdsTokenEndpoint)
		t2 := getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint)
		if t1 == nil || t2 == nil || !strings.EqualFold(t1.AccessToken, t2.AccessToken) {
			log.Error("Tokens acquired from IMDS with and without identity client ID do not match")
		} else {
			log.Infof("Try decoding your token %s at https://jwt.io", t1.AccessToken)
			return t1, nil
		}
	}
	return nil, nil
}

func getTokenFromIMDS(imdsTokenEndpoint string) *adal.Token {
	spt, err := adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(imdsTokenEndpoint, resourceName, identityClientID)
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

	token := spt.Token()
	if token.IsZero() {
		log.Errorf("%+v is a zero token", token)
		return nil
	}

	log.Infof("successfully acquired a service principal token from %s", imdsTokenEndpoint)
	return &token
}

func getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint string) *adal.Token {
	spt, err := adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(imdsTokenEndpoint, resourceName, identityClientID)
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

	token := spt.Token()
	if token.IsZero() {
		log.Errorf("%+v is a zero token", token)
		return nil
	}

	log.Info("successfully acquired a service principal token from %s using a user-assigned identity (%s)", imdsTokenEndpoint, identityClientID)
	return &token
}

func curlIMDSMetadataInstanceEndpoint() {
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance?api-version=2017-08-01", nil)
	if err != nil {
		log.Errorf("failed to create a new HTTP request, error: %+v", err)
		return
	}
	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	defer resp.Body.Close()
}
