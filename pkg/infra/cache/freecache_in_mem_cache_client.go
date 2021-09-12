package cache

import (
	"context"
	"github.com/coocood/freecache"

	"time"
)

// FreeCacheInMemCacheClient implements ICacheClient  interface
var _ ICacheClient = (*FreeCacheInMemCacheClient)(nil)

// FreeCacheInMemCacheClient  is in-mem cache. it wraps freecache.Cache struct.
// For more information regarding this cache - "https://github.com/coocood/freecache".
type FreeCacheInMemCacheClient struct {
	// freeCache is the client of freecache.Cache
	freeCache *freecache.Cache
}

// FreeCacheInMemClientConfiguration is the configuration for FreeCacheInMemCacheClient.
type FreeCacheInMemClientConfiguration struct {
	// cacheSize in bytes.
	cacheSize int
}

// CreateFreeCacheInMemCacheClient is constructor for FreeCacheInMemCacheClient
func CreateFreeCacheInMemCacheClient(configuration *FreeCacheInMemClientConfiguration) *FreeCacheInMemCacheClient {
	bigCache := freecache.NewCache(configuration.cacheSize)
	return newFreeCacheInMemCacheClient(bigCache)
}

// newFreeCacheInMemCacheClient is private constructor of FreeCacheInMemCacheClient.
func newFreeCacheInMemCacheClient(cache *freecache.Cache) *FreeCacheInMemCacheClient {
	return &FreeCacheInMemCacheClient{
		freeCache: cache,
	}
}

func (client *FreeCacheInMemCacheClient) Get(ctx context.Context, key string) (string, error) {
	entry, err := client.freeCache.Get([]byte(key))
	if err != nil {
		return "", nil
	}
	// Convert entry to string
	value := string(entry)
	return value, nil
}

func (client *FreeCacheInMemCacheClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if expiration < 0 {
		return NewNegativeExpirationError(expiration)
	}
	//TODO Should the expiration be in seconds?
	//TODO Check overflow of expiration - casting from float64 to int?
	return client.freeCache.Set([]byte(key), []byte(value), int(expiration.Seconds()))
}
