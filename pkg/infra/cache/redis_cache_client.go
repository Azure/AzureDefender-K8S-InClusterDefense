package cache

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/go-redis/redis"
	"time"
)

// RedisCacheClient redis cache client implements ICacheClient.
// It wraps go-redis client - "github.com/go-redis/redis".
type RedisCacheClient struct {
	// redisClient is the redis client from "github.com/go-redis/redis".
	// This struct delegate redisClient.
	redisClient wrappers.IRedisBaseClientWrapper
}

// NewRedisCacheClient Cto'r for RedisCacheClient
func NewRedisCacheClient(redisClient wrappers.IRedisBaseClientWrapper) ICacheClient {
	return &RedisCacheClient{
		redisClient: redisClient,
	}
}

// CreateRedisCacheClient is factory for RedisCacheClient
func CreateRedisCacheClient(configuration *wrappers.RedisBaseCacheClientConfiguration) ICacheClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr:            configuration.Addr,
		Password:        configuration.Password,
		DB:              configuration.Db,
		MaxRetries:      configuration.MaxRetries,
		MinRetryBackoff: configuration.MinRetryBackoff,
	})

	return NewRedisCacheClient(redisClient)
}

func (client *RedisCacheClient) Get(key string) (string, error) {
	return client.redisClient.Get(key).Result()
}

func (client *RedisCacheClient) Set(key string, value string, expiration time.Duration) {
	client.redisClient.Set(key, value, expiration)
}
