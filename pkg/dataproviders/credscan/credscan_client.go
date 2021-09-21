package credscan

import (
	"time"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// cred scan server url - the sidecar
	_credScanServerUrl = "http://localhost:500/scanString"

	_threshold = 75

	_scanStatusHealthy = "healthy"

	_scanStatusUnhealthy = "unhealthy"

)

//interface_____________________________________________________________________________________________________________

type ICredScanDataProvider interface {
	GetCredScanResults(resourceMetadata admission.Request) ([]*CredScanInfo, error)
}

//structs_______________________________________________________________________________________________________________

// CredScanDataProvider for ICredScanDataProvider implementation
type CredScanDataProvider struct {
	tracerProvider trace.ITracerProvider
}

// structs hierarchy for credScan results_______________________________________________________________________________
// The hierarchy is bottom up

// CredentialInfoStruct - a struct contain the weakness description
type CredentialInfoStruct struct {
	//the weakness description
	Description string `json:"name"`
}

// CredScanInfo represents cred scan information about a possible unhealthy property
type CredScanInfo struct {

	// a struct contain the weakness description
	CredentialInfo CredentialInfoStruct `json:"credentialInfo"`

	// a number represent the MatchingConfidence of the weakness (from 1 to 100)
	MatchingConfidence float64 `json:"MatchingConfidence"`

	// scanStatus == healthy -> no secret found. scanStatus == unhealthy -> secret found
	ScanStatus string
}


// CredScanInfoList a list of cred scan information
type CredScanInfoList struct {

	//GeneratedTimestamp represents the time the scan info list (this) was generated
	GeneratedTimestamp time.Time `json:"generatedTimestamp"`

	//List of CredScanInfo
	CredScanResults []*CredScanInfo `json:"CredScanInfo"`
}