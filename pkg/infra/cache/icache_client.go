package cache

import "time"

type ICacheClient interface {
	// Get gets a value from the cache. It returns  error when key does not exist.
	Get(key string) (string, error)

	//Set sets new item in the cache.
	//Zero expiration means the key has no expiration time.
	Set(key string, value string, expiration time.Duration)
}
