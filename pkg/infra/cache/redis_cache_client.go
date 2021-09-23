package cache

import (
	"context"
	cachemetrics "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/operations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
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
	//tracerProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter
	metricSubmitter metric.IMetricSubmitter
}

// NewRedisCacheClient is factory for RedisCacheClient
func NewRedisCacheClient(instrumentationProvider instrumentation.IInstrumentationProvider, redisBaseClient wrappers.IRedisBaseClientWrapper) *RedisCacheClient {

	return &RedisCacheClient{
		tracerProvider:  instrumentationProvider.GetTracerProvider("RedisCacheClient"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		redisClient:     redisBaseClient,
	}
}

// Get gets a value from the redis cache. It returns error when key does not exist.
func (client *RedisCacheClient) Get(ctx context.Context, key string) (string, error) {
	tracer := client.tracerProvider.GetTracer("Get")
	tracer.Info("Get key executed", "Key", key)

	operationStatus := operations.MISS
	defer client.metricSubmitter.SendMetric(1, cachemetrics.NewCacheOperationMetric(client, operations.GET, operationStatus))

	value, err := client.redisClient.Get(ctx, key).Result()
	// Check if key is missing.
	if err == redis.Nil {
		err = NewMissingKeyCacheError(key)
		tracer.Error(err, "", "Key", key)
		return "", err
		// Unexpected error from redis client
	} else if err != nil {
		tracer.Error(err, "Failed to get a key", "Context", ctx, "Key", key)
		return "", err
	}

	operationStatus = operations.HIT
	tracer.Info("Key found", "Key", key, "value", value)
	return value, nil
}

//Set sets new item in the redis cache.
//Zero expiration means the key has no expiration time.
// It returns error when there was a problem trying to set the key.
// expiration must be non-negative expiration.
func (client *RedisCacheClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	tracer := client.tracerProvider.GetTracer("Set")
	tracer.Info("Set new key", "Key", key, "Value", value, "Expiration", expiration)

	operationStatus := operations.MISS
	defer client.metricSubmitter.SendMetric(1, cachemetrics.NewCacheOperationMetric(client, operations.SET, operationStatus))

	if expiration < 0 {
		err := NewNegativeExpirationCacheError(expiration)
		tracer.Error(err, "", "Key", key, "Value", value, "Expiration", expiration)
		return err
	}

	if err := client.redisClient.Set(ctx, key, value, expiration).Err(); err != nil {
		tracer.Error(err, "Failed to set a key", "Key", key, "Value", value, "Expiration", expiration)
		return err
	}

	operationStatus = operations.HIT
	tracer.Info("Key was added successfully", "Key", key, "value", value)
	return nil
}
