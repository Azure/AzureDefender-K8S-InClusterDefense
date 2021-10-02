package cache

import (
	"context"
	cachemetrics "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/operations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"strings"
	"time"
)

const (
	// _missingKeyErrorFreeCacheString is that is returned from freecache in case that the key is not found.
	_missingKeyErrorFreeCacheString = "entry not found"
	// client type of free cache.
	_freeCacheClientType clientType = "FreeCacheInMemCacheClient"
)

// FreeCacheInMemCacheClient implements ICacheClient  interface
var _ ICacheClient = (*FreeCacheInMemCacheClient)(nil)

// FreeCacheInMemCacheClient  is in-mem cache. it wraps freecache.Cache struct.
// For more information regarding this cache - "https://github.com/coocood/freecache".
type FreeCacheInMemCacheClient struct {
	// freeCache is the cache of freecache.Cache
	freeCache wrappers.IFreeCacheInMemCacheWrapper
	//tracerProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter
	metricSubmitter metric.IMetricSubmitter
}

// NewFreeCacheInMemCacheClient is constructor for FreeCacheInMemCacheClient.
func NewFreeCacheInMemCacheClient(instrumentationProvider instrumentation.IInstrumentationProvider, cache wrappers.IFreeCacheInMemCacheWrapper) *FreeCacheInMemCacheClient {
	return &FreeCacheInMemCacheClient{
		tracerProvider:  instrumentationProvider.GetTracerProvider("FreeCacheInMemCacheClient"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		freeCache:       cache,
	}
}

func (client *FreeCacheInMemCacheClient) Get(ctx context.Context, key string) (string, error) {
	tracer := client.tracerProvider.GetTracer("Get")
	tracer.Info("Get key executed", "Key", key)

	operationStatus := operations.MISS
	defer client.metricSubmitter.SendMetric(1, cachemetrics.NewCacheOperationMetric(client, operationStatus))

	entry, err := client.freeCache.Get([]byte(key))
	// Check if key is missing
	if (err != nil && strings.ToLower(err.Error()) == _missingKeyErrorFreeCacheString) || entry == nil {
		err = NewMissingKeyCacheError(key)
		tracer.Error(err, "", "Key", key)
		client.metricSubmitter.SendMetric(1, cachemetrics.NewGetErrEncounteredMetric(err, _freeCacheClientType))
		return "", err
		// Unexpected error was returned from freecache client.
	} else if err != nil {
		tracer.Error(err, "Failed to get a key", "Key", key, "value", string(entry))
		client.metricSubmitter.SendMetric(1, cachemetrics.NewGetErrEncounteredMetric(err, _freeCacheClientType))
		return "", err
	}

	operationStatus = operations.HIT
	// Convert entry ([]byte) to string
	value := string(entry)

	tracer.Info("Key found", "Key", key, "value", value)
	return value, nil
}

func (client *FreeCacheInMemCacheClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	tracer := client.tracerProvider.GetTracer("Set")
	tracer.Info("Set new key", "Key", key, "Value", value, "Expiration", expiration)

	if expiration < 0 {
		err := NewNegativeExpirationCacheError(expiration)
		tracer.Error(err, "", "Key", key, "Value", value, "Expiration", expiration)
		client.metricSubmitter.SendMetric(1, cachemetrics.NewSetErrEncounteredMetric(err, _freeCacheClientType))

		return err
	}

	//TODO Check overflow of expiration - casting from float64 to int?
	expirationInt := int(expiration.Seconds())
	err := client.freeCache.Set([]byte(key), []byte(value), expirationInt)
	if err != nil {
		tracer.Error(err, "Failed to set a key", "Key", key, "Value", value, "Expiration", expiration)
		client.metricSubmitter.SendMetric(1, cachemetrics.NewSetErrEncounteredMetric(err, _freeCacheClientType))

		return err
	}

	tracer.Info("Key was added successfully", "Key", key, "value", value)
	return nil
}
