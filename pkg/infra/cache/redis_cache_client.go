package cache

import (
	"context"
	"crypto/tls"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/go-redis/redis/v8"
	"time"
)

// RedisCacheClient implements ICacheClient interface
var _ ICacheClient = (*RedisCacheClient)(nil)

// RedisCacheClient redis cache client implements ICacheClient.
// It wraps go-redis client - "github.com/go-redis/redis".
type RedisCacheClient struct {
	// redisClient is the redis client from "github.com/go-redis/redis".
	// This struct delegate redisClient.
	redisClient wrappers.IRedisBaseClientWrapper
}

// newRedisCacheClient Cto'r for RedisCacheClient
func newRedisCacheClient(redisClient wrappers.IRedisBaseClientWrapper) *RedisCacheClient {
	return &RedisCacheClient{
		redisClient: redisClient,
	}
}

// CreateRedisCacheClient is factory for RedisCacheClient
func CreateRedisCacheClient(configuration *RedisCacheClientConfiguration) *RedisCacheClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr:            configuration.Address,
		Password:        configuration.Password,
		DB:              configuration.Table,
		MaxRetries:      configuration.MaxRetries,
		MinRetryBackoff: configuration.MinRetryBackoff,
	})
	_, _ = redisClient.Ping(context.Background()).Result()

	return newRedisCacheClient(redisClient)
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

// Get gets a value from the redis cache. It returns error when key does not exist.
func (client *RedisCacheClient) Get(ctx context.Context, key string) (string, error) {
	return client.redisClient.Get(ctx, key).Result()
}

//Set sets new item in the redis cache.
//Zero expiration means the key has no expiration time.
// It returns error when there was a problem trying to set the key.
// expiration must be non-negative expiration.
func (client *RedisCacheClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if expiration < 0 {
		return NewNegativeExpirationError(expiration)
	}

	return client.redisClient.Set(ctx, key, value, expiration).Err()
}
