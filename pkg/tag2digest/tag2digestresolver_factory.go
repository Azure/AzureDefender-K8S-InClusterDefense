package tag2digest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	"github.com/pkg/errors"
	"time"
)

// Tag2DigestResolverFactory Factory to create Tag2DigestResolver
type Tag2DigestResolverFactory struct {
}

// NewTag2DigestResolverFactory is a Constructor for Tag2DigestResolverFactory
func NewTag2DigestResolverFactory () *Tag2DigestResolverFactory{
	return &Tag2DigestResolverFactory{}
}

// CreateTag2DigestResolver creates a new instance of Tag2DigestResolver
func (factory *Tag2DigestResolverFactory) CreateTag2DigestResolver(instrumentationProvider instrumentation.IInstrumentationProvider, registryClient registry.IRegistryClient, cacheClient cache.ICacheClient, tag2DigestResolverConfiguration *Tag2DigestResolverConfiguration) (*Tag2DigestResolver, error){
	cacheExpirationTimeForResults, err := time.ParseDuration(tag2DigestResolverConfiguration.CacheExpirationTimeForResults)
	if err != nil{
		return nil, errors.Wrapf(err, "Given tag2DigestResolverConfiguration.cacheExpirationTimeForResults string is not a valid time duration")
	}
	cacheExpirationTimeForErrors, err := time.ParseDuration(tag2DigestResolverConfiguration.CacheExpirationTimeForErrors)
	if err != nil{
		return nil, errors.Wrapf(err, "Given tag2DigestResolverConfiguration.cacheExpirationTimeForErrors string is not a valid time duration")
	}
	tag2DigestResolverCacheClient := &Tag2DigestResolverCacheClient{
		cacheClient: cacheClient,
		CacheExpirationTimeForResults: cacheExpirationTimeForResults,
		CacheExpirationTimeForErrors: cacheExpirationTimeForErrors,
	}
	return NewTag2DigestResolver(instrumentationProvider, registryClient, tag2DigestResolverCacheClient), nil
}


