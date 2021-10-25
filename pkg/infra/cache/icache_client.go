package cache

import (
	"time"
)

// ClientType is type that is used for all cache types - it is used for metrics.
type clientType = string

// ICacheClient is a basic cache client interface.
type ICacheClient interface {
	// Get gets a value from the cache. It returns  error when key does not exist.
	Get(key string) (string, error)

	//Set sets new item in the cache.
	//Zero expiration means the key has no expiration time.
	// It returns error when there was a problem trying to set the key.
	Set(key string, value string, expiration time.Duration) error
}
