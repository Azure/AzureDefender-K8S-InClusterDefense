package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
)

type IARGDataProvider interface {
	TryGetImageVulnerabilityScanResults(registryHost string, repository string, digest string) (contracts.ScanStatus, []*contracts.ScanFinding)
}

type ARGDataProvider struct {
}
