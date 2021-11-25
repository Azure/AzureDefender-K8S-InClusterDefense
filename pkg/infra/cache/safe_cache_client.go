package cache

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"time"
)

// SafeCacheClient implements ICacheClient interface
var _ ICacheClient = (*SafeCacheClient)(nil)

// SafeCacheClient implements ICacheClient.
// It adds functionality to ICacheClient such that it forbids keys with no expiration in the cache.
type SafeCacheClient struct {
	// cacheClient a cache client from to forward calls to
	cacheClient ICacheClient
}

// NewSafeCacheClient Constructor
func NewSafeCacheClient(cacheClient ICacheClient) *SafeCacheClient{
	return &SafeCacheClient{
		cacheClient: cacheClient,
	}
}

// Get gets a value from the cache. It returns  error when key does not exist.
func (client *SafeCacheClient) Get(key string) (string, error){
	return client.cacheClient.Get(key)
}

//Set sets new item in the cache. If the expiration time is less or equal to 0 returns an error
// Only keys with expiration time are allowed (expiration time zero means the key doesn't expire)
func (client *SafeCacheClient) Set(key string, value string, expiration time.Duration) error{
	if expiration <= 0{
		return utils.SetKeyInCacheWithNoExpiration
	}
	return client.cacheClient.Set(key, value, expiration)
}
