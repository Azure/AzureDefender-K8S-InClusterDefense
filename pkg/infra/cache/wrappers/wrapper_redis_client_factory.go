package wrappers

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"io/ioutil"
)

// WrapperRedisClientFactory is a factory for WrapperRedisClient.
type WrapperRedisClientFactory struct {
	// tracerProvider is the tracer provider for the WrapperRedisClientFactory
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the WrapperRedisClientFactory
	metricSubmitter metric.IMetricSubmitter
}

// NewWrapperRedisClientFactory constructor
func NewWrapperRedisClientFactory (instrumentationProvider instrumentation.IInstrumentationProvider) *WrapperRedisClientFactory{
	return &WrapperRedisClientFactory{
		tracerProvider:  instrumentationProvider.GetTracerProvider("WrapperRedisClientFactory"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}
}

// Create creates redis client by getting RedisCacheClientConfiguration and extracting the certificates and password.
func (factory *WrapperRedisClientFactory) Create(configuration *RedisCacheClientConfiguration) (*redis.Client, error){
	tracer := factory.tracerProvider.GetTracer("Create")

	// Get certificates
	cert, err := tls.LoadX509KeyPair(configuration.TlsCrtPath, configuration.TlsKeyPath)
	if err != nil {
		err = errors.Wrap(err, "Failed to LoadX509KeyPair")
		tracer.Error(err, "")
		return nil, err
	}
	caCert, err := ioutil.ReadFile(configuration.CaCertPath)
	if err != nil {
		err = errors.Wrap(err, "Failed to read ca.cert file ")
		tracer.Error(err, "")
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	// Create tlsConfig object with the cert files
	tlsConfig := &tls.Config{
		ServerName:   configuration.Host,
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}

	// Get password
	passwordInBytes, err := ioutil.ReadFile(configuration.PasswordPath)
	if err != nil {
		err = errors.Wrap(err, "Failed to read password file")
		tracer.Error(err, "")
		return nil, err
	}
	password := string(passwordInBytes)


	// Create redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:            configuration.Address,
		Password:        password,
		DB:              configuration.Table,
		MaxRetries:      configuration.MaxRetries,
		MinRetryBackoff: configuration.MinRetryBackoff,
		TLSConfig: tlsConfig,
	})

	return redisClient, nil
}
