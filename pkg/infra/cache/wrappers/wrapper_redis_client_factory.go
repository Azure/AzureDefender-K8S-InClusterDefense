package wrappers

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// WrapperRedisClientFactory is a factory for WrapperRedisClient.
type WrapperRedisClientFactory struct {
	// tracerProvider is the tracer provider for the WrapperRedisClientFactory
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the WrapperRedisClientFactory
	metricSubmitter metric.IMetricSubmitter
}

// NewWrapperRedisClientFactory constructor
func NewWrapperRedisClientFactory(instrumentationProvider instrumentation.IInstrumentationProvider) *WrapperRedisClientFactory {
	return &WrapperRedisClientFactory{
		tracerProvider:  instrumentationProvider.GetTracerProvider("WrapperRedisClientFactory"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}
}

// Create creates redis client by getting RedisCacheClientConfiguration and extracting the certificates and password.
// TODO  use only CA bundle and Server DNS + Password to Authenticate
func (factory *WrapperRedisClientFactory) Create(configuration *RedisCacheClientConfiguration) (*redis.Client, error) {
	tracer := factory.tracerProvider.GetTracer("Create")

	// Get tlsConfig
	tlsConfig, err := utils.CreateTlsConfig(configuration.TlsCrtPath, configuration.TlsKeyPath, configuration.CaCertPath, configuration.Host)
	if err != nil {
		err = errors.Wrap(err, "Failed to create tls config object")
		tracer.Error(err, "")
		factory.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "WrapperRedisClientFactory.Create"))
		return nil, err
	}

	// Get password
	password, err := utils.GetPasswordFromFile(configuration.PasswordPath)
	if err != nil {
		err = errors.Wrap(err, "Failed to get password from secret")
		tracer.Error(err, "")
		factory.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "WrapperRedisClientFactory.Create"))
		return nil, err
	}

	// Create redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:            configuration.Address,
		Password:        password,
		DB:              configuration.Table,
		MaxRetries:      configuration.MaxRetries,
		MinRetryBackoff: configuration.MinRetryBackoff,
		TLSConfig:       tlsConfig,
	})

	return redisClient, nil
}
