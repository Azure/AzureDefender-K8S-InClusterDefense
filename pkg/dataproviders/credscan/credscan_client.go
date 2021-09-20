package credscan

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"time"
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

type CredentialInfoStruct struct {
	Description string `json:"name"`
}

// CredScanInfo represents cred scan information
type CredScanInfo struct {

	//  Name container name in resource spec
	CredentialInfo CredentialInfoStruct `json:"credentialInfo"`

	//  Name container name in resource spec
	MatchingConfidence float64 `json:"MatchingConfidence"`

	// scanStatus == healthy -> no secret found. scanStatus == unhealthy -> secret found
	ScanStatus string
}


// CredScanInfoList a list of cred scan information
type CredScanInfoList struct {

	//GeneratedTimestamp represents the time the scan info list (this) was generated
	GeneratedTimestamp time.Time `json:"generatedTimestamp"`

	//Containers List of CredScanInfo
	CredScanResults []*CredScanInfo `json:"CredScanInfo"`
}