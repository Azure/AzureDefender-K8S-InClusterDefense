package wrappers

import (
	"context"
	"crypto/tls"
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

// RedisCacheClientConfiguration redis cache client configuration
type RedisCacheClientConfiguration struct {
	// Address host:port address.
	Address string
	// Password Optional password. Must match the password specified in the
	// requirement pass server configuration option.
	Password string
	// Table is Database to be selected after connecting to the server.
	Table int
	// MaxRetries Maximum number of retries before giving up.
	// Default is to not retry failed commands.
	MaxRetries int
	// MinRetryBackoff Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config
}

func NewRedisBaseClientWrapper(configuration *RedisCacheClientConfiguration) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:            configuration.Address,
		Password:        configuration.Password,
		DB:              configuration.Table,
		MaxRetries:      configuration.MaxRetries,
		MinRetryBackoff: configuration.MinRetryBackoff,
	})
	//_, _ = redisClient.Ping(context.Background()).Result()
	return redisClient
}
