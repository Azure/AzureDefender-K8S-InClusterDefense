package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
)

type IARGDataProvider interface {
	TryGetImageVulnerabilityScanResults(registry string, repository string, digest string) ([]*contracts.ScanFinding, error)
}

type ARGDataProvider struct {
}
