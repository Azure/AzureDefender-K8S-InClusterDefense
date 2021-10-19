package cache

import (
	"context"
	cachemetrics "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/operations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	_redisClientType = "RedisCacheClient"
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
	//retryPolicy retry policy for communication with redis cluster.
	retryPolicy retrypolicy.IRetryPolicy
}

// NewRedisCacheClient is factory for RedisCacheClient
func NewRedisCacheClient(instrumentationProvider instrumentation.IInstrumentationProvider, redisBaseClient wrappers.IRedisBaseClientWrapper, retryPolicy retrypolicy.IRetryPolicy) *RedisCacheClient {

	return &RedisCacheClient{
		tracerProvider:  instrumentationProvider.GetTracerProvider("RedisCacheClient"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		redisClient:     redisBaseClient,
		retryPolicy:     retryPolicy,
	}
}

// Get gets a value from the redis cache. It returns error when key does not exist.
func (client *RedisCacheClient) Get(ctx context.Context, key string) (string, error) {
	tracer := client.tracerProvider.GetTracer("Get")
	tracer.Info("Get key executed", "Key", key)

	value, err := client.retryPolicy.RetryActionString(
		/*action ActionString get key using client.redisClient */
		func() (string, error) { return client.redisClient.Get(ctx, key).Result() },
		/*handler ShouldRetryOnSpecificError - handle with key is missing error*/
		func(err error) bool {
			if err == redis.Nil { // In case that key is missing
				client.metricSubmitter.SendMetric(1, cachemetrics.NewCacheClientGetMetric(client, operations.MISS))
				tracer.Info("Missing Key", "Key", key)
				err = NewMissingKeyCacheError(key)
				return false
			}

			client.metricSubmitter.SendMetric(1, cachemetrics.NewGetErrEncounteredMetric(err, _redisClientType))
			tracer.Error(err, "", "key", key)
			return true
		},
	)
	if err != nil {
		return "", err
	}
	// Get succeed.
	client.metricSubmitter.SendMetric(1, cachemetrics.NewCacheClientGetMetric(client, operations.HIT))
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

	if expiration < 0 {
		err := NewNegativeExpirationCacheError(expiration)
		tracer.Error(err, "", "Key", key, "Value", value, "Expiration", expiration)
		client.metricSubmitter.SendMetric(1, cachemetrics.NewSetErrEncounteredMetric(err, _redisClientType))

		return err
	}

	err := client.retryPolicy.RetryAction(
		// Action - set the values redis client.
		func() error { return client.redisClient.Set(ctx, key, value, expiration).Err() },
		// HandleError - if the err is redis.Nil then it means that the get is not exist.
		func(err error) bool { return err != redis.Nil },
	)

	if err != nil && err != redis.Nil {
		client.metricSubmitter.SendMetric(1, cachemetrics.NewSetErrEncounteredMetric(err, _redisClientType))
		tracer.Error(err, "Failed to set a key", "Key", key, "Value", value, "Expiration", expiration)
		return err
	}

	client.metricSubmitter.SendMetric(utils.GetSizeInBytes(value), cachemetrics.NewAddItemToCacheMetric(_redisClientType))
	tracer.Info("Key was added successfully", "Key", key, "value", value)
	return nil
}
