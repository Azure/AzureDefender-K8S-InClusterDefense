package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"	
)


type ISecInfoProvider interface {
	GetVulnerabilityScanInfo() *contracts.VulnerabilityScanInfo
}

type SecInfoProvider struct {
}

func NewSecInfoProvider() *SecInfoProvider {
	return &SecInfoProvider{}
}

func (provider *SecInfoProvider) GetVulnerabilityScanInfo() *contracts.ContainerScanSummary {
	return &contracts.VulnerabilityScanInfo{
		Timestamp: nil,
		Containers: nil
	}


}
