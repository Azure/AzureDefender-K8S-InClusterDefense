package wrappers

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// redis.Client implements IRedisBaseClientWrapper interface.
var _ IRedisBaseClientWrapper = (*redis.Client)(nil)

// IRedisBaseClientWrapper is a wrapper interface for base client of redis - "github.com/go-redis/redis"
type IRedisBaseClientWrapper interface {
	// Set Redis `SET key value [expiration]` command.
	// Use expiration for `SETEX`-like behavior.
	// Zero expiration means the key has no expiration time.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd

	// Get Redis `GET key` command. It returns redis.Nil error when key does not exist.
	Get(ctx context.Context, key string) *redis.StringCmd
}
