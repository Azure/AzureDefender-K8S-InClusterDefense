package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	corev1 "k8s.io/api/core/v1"
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {

	// GetContainerVulnerabilityScanInfo receives containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainerVulnerabilityScanInfo(*corev1.Container) (*contracts.ContainerVulnerabilityScanInfo, error)
}

// AzdSecInfoProvider represents default implementation of IAzdSecInfoProvider interface
type AzdSecInfoProvider struct {
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider() *AzdSecInfoProvider {
	return &AzdSecInfoProvider{}
}

// GetContainerVulnerabilityScanInfo receives containers list, and returns their fetched ContainerVulnerabilityScanInfo
func (*AzdSecInfoProvider) GetContainerVulnerabilityScanInfo(container *corev1.Container) (*contracts.ContainerVulnerabilityScanInfo, error) {
	// TODO
	info := &contracts.ContainerVulnerabilityScanInfo{
		Name: container.Name,
		Image: &contracts.Image{
			// TODO : change
			Registry: container.Image,
			// TODO : change
			Repository: container.Image,
			// TODO: change
			Digest: container.Image,
		},
		ScanStatus:   contracts.HealthyScan,
		ScanFindings: nil,
	}

	return info, nil
}
