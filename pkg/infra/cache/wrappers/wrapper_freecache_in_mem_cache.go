package wrappers

import (
	"github.com/coocood/freecache"
)

// freecache.Cache implements IFreeCacheInMemCacheWrapper interface.
var _ IFreeCacheInMemCacheWrapper = (*freecache.Cache)(nil)

// IFreeCacheInMemCacheWrapper is a wrapper interface for base client of redis - "github.com/coocood/freecache"
type IFreeCacheInMemCacheWrapper interface {

	// Set sets a key, value and expiration for a cache entry and stores it in the cache.
	// If the key is larger than 65535 or value is larger than 1/1024 of the cache size,
	// the entry will not be written to the cache. expireSeconds <= 0 means no expire,
	// but it can be evicted when cache is full.
	Set(key, value []byte, expireSeconds int) (err error)
	// Get returns the value or not found error.
	Get(key []byte) (value []byte, err error)
}

// FreeCacheInMemWrapperCacheConfiguration is the configuration for FreeCacheInMemCache.
type FreeCacheInMemWrapperCacheConfiguration struct {
	// CacheSize in bytes.
	CacheSize int
}

// NewFreeCacheInMem is Ctor for FreeCacheInMemWrapperCache
func NewFreeCacheInMem(configuration *FreeCacheInMemWrapperCacheConfiguration) *freecache.Cache {
	return freecache.NewCache(configuration.CacheSize)
}
