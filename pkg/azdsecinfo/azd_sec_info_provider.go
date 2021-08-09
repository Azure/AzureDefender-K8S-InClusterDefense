package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {

	// GetContainersVulnerabilityScanInfo receives containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainersVulnerabilityScanInfo() (*contracts.ContainersVulnerabilityScanInfo, error)
}

// AzdSecInfoProvider represents default implementation of IAzdSecInfoProvider interface
type AzdSecInfoProvider struct {
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider() *AzdSecInfoProvider {
	return &AzdSecInfoProvider{}
}

// GetContainersVulnerabilityScanInfo receives containers list, and returns their fetched ContainersVulnerabilityScanInfo
func (*AzdSecInfoProvider) GetContainersVulnerabilityScanInfo() (*contracts.ContainersVulnerabilityScanInfo, error) {

	// info := &{

	// }

	return nil, nil
}
