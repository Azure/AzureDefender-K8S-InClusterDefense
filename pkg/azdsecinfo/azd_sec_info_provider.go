package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	//name "github.com/google/go-containerregistry/pkg/name"
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
			Name: container.Image,
			// TODO: change
			Digest: container.Image,
		},
		// TODO : change
		ScanStatus: contracts.HealthyScan,
		// TODO : change
		ScanFindings: []*contracts.ScanFinding{
			{
				Patchable: true,
				Id:        "123",
				Severity:  "High",
			},
		},
	}

	return info, nil
}
