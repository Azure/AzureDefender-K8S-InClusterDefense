package cache

import (
	"context"
	"time"
)

// ICacheClient is a basic cache client interface.
type ICacheClient interface {
	// Get gets a value from the cache. It returns  error when key does not exist.
	Get(ctx context.Context, key string) (string, error)

	//Set sets new item in the cache.
	//Zero expiration means the key has no expiration time.
	// It returns error when there was a problem trying to set the key.
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}
