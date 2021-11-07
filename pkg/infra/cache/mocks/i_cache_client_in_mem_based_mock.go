package mocks

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachewrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
)

var (
	// _configuration is generic FreeCacheInMemWrapperCacheConfiguration for tests purpose
	_configuration = &cachewrappers.FreeCacheInMemWrapperCacheConfiguration{CacheSize: 10000000}
)

// NewICacheInMemBasedMock returns FreeCacheInMemCacheClient that can be used as running time ICacheClient mock for other cache types (i.e redis cache)
func NewICacheInMemBasedMock() *cache.FreeCacheInMemCacheClient{
	freeCacheInMemCache := cachewrappers.NewFreeCacheInMem(_configuration)
	return cache.NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), freeCacheInMemCache)
}
