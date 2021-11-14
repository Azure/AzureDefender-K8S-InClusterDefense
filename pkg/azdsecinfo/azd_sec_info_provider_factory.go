package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	"github.com/pkg/errors"
	"time"
)

type AzdSecInfoProviderFactory struct {
}

func NewAzdSecInfoProviderFactory () *AzdSecInfoProviderFactory{
	return &AzdSecInfoProviderFactory{}
}

func (factory *AzdSecInfoProviderFactory) CreateTag2DigestResolver(instrumentationProvider instrumentation.IInstrumentationProvider,
	argDataProvider arg.IARGDataProvider,
	tag2digestResolver tag2digest.ITag2DigestResolver,
	GetContainersVulnerabilityScanInfoTimeoutDuration *utils.TimeoutConfiguration, azdSecInfoProviderConfiguration *AzdSecInfoProviderConfiguration, cacheClient cache.ICacheClient) (*AzdSecInfoProvider, error){
	cacheExpirationTimeTimeout, err := time.ParseDuration(azdSecInfoProviderConfiguration.CacheExpirationTimeTimeout)
	if err != nil{
		return nil, errors.Wrapf(err, "Given azdSecInfoProviderConfiguration.cacheExpirationTimeTimeout string is not a valid time duration")
	}
	azdSecInfoProviderCacheClient := &AzdSecInfoProviderCacheClient{
		cacheClient: cacheClient,
		CacheExpirationTimeTimeout: cacheExpirationTimeTimeout,
	}
	return NewAzdSecInfoProvider(instrumentationProvider, argDataProvider, tag2digestResolver, GetContainersVulnerabilityScanInfoTimeoutDuration, azdSecInfoProviderCacheClient), nil
}
