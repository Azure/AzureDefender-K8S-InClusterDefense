package azdsecinfo

import (
	"time"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {

	// GetContainerVulnerabilityScanInfo receives containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainerVulnerabilityScanInfo() (*contracts.ContainerVulnerabilityScanInfo, error)
}

// AzdSecInfoProvider represents default implementation of IAzdSecInfoProvider interface
type AzdSecInfoProvider struct {
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider() *AzdSecInfoProvider {
	return &AzdSecInfoProvider{}
}

// GetContainerVulnerabilityScanInfo receives containers list, and returns their fetched ContainerVulnerabilityScanInfo
func (*AzdSecInfoProvider) GetContainerVulnerabilityScanInfo() (*contracts.ContainerVulnerabilityScanInfo, error) {

	info := &contracts.ContainerVulnerabilityScanInfo{
		GeneratedTimestamp: time.Now().UTC(),
		Name:               "tomer",
		Image: &contracts.Image{
			Registry:   "tomer.azurecr.io",
			Repository: "redis",
			Digest:     "sha256",
		},
		ScanStatus:   contracts.HealthyScan,
		ScanFindings: nil,
	}

	return info, nil
}
