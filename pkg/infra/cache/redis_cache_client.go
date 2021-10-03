package cache

import (
	"context"
	cachemetrics "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/operations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
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
	retryPolicyConfiguration *utils.RetryPolicyConfiguration
}

// NewRedisCacheClient is factory for RedisCacheClient
func NewRedisCacheClient(instrumentationProvider instrumentation.IInstrumentationProvider, redisBaseClient wrappers.IRedisBaseClientWrapper, retryPolicy *utils.RetryPolicyConfiguration) *RedisCacheClient {

	return &RedisCacheClient{
		tracerProvider:           instrumentationProvider.GetTracerProvider("RedisCacheClient"),
		metricSubmitter:          instrumentationProvider.GetMetricSubmitter(),
		redisClient:              redisBaseClient,
		retryPolicyConfiguration: retryPolicy,
	}
}

// Get gets a value from the redis cache. It returns error when key does not exist.
func (client *RedisCacheClient) Get(ctx context.Context, key string) (string, error) {
	tracer := client.tracerProvider.GetTracer("Get")
	tracer.Info("Get key executed", "Key", key)
	// Default operation status is MISS, if we achive the end of the function then the operation changed to HIT.
	operationStatus := operations.MISS
	retryCount := 0
	defer client.metricSubmitter.SendMetric(retryCount-1, cachemetrics.NewCacheClientGetMetric(client, operationStatus))

	retryDuration, err := client.retryPolicyConfiguration.GetBackOffDuration()
	if err != nil {
		return "", errors.Wrapf(err, "cannot parse given retry duration <(%v)>", client.retryPolicyConfiguration.RetryDuration)
	}

	// Update retryCount to 1
	retryCount = 1
	var value string
	for retryCount < client.retryPolicyConfiguration.RetryAttempts+1 {
		value, err = client.redisClient.Get(ctx, key).Result()
		// Check if key is missing, we don't have to retry and return error.
		if err == redis.Nil {
			err = NewMissingKeyCacheError(key)
			break
			// Unexpected error from redis client - retry.
		} else if err != nil {
			retryCount += 1
			// wait (retryCount * craneWrapper.retryDuration) milliseconds between retries
			time.Sleep(time.Duration(retryCount) * retryDuration)
			// Get succeed.
		} else {
			break
		}
	}

	if err != nil {
		client.metricSubmitter.SendMetric(1, cachemetrics.NewGetErrEncounteredMetric(err, _redisClientType))
		tracer.Error(err, "", "key", key)
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

	if expiration < 0 {
		err := NewNegativeExpirationCacheError(expiration)
		tracer.Error(err, "", "Key", key, "Value", value, "Expiration", expiration)
		client.metricSubmitter.SendMetric(1, cachemetrics.NewSetErrEncounteredMetric(err, _redisClientType))

		return err
	}

	if err := client.redisClient.Set(ctx, key, value, expiration).Err(); err != nil {
		tracer.Error(err, "Failed to set a key", "Key", key, "Value", value, "Expiration", expiration)
		client.metricSubmitter.SendMetric(1, cachemetrics.NewSetErrEncounteredMetric(err, _redisClientType))

		return err
	}

	tracer.Info("Key was added successfully", "Key", key, "value", value)
	return nil
}
