package registry

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

// IRegistryClient container registry based client
type IRegistryClient interface {
	GetDigest(imageRef string) (string, error)
}

// RegistryClient container registry based client
type RegistryClient struct {
	tracerProvider  trace.ITracerProvider
	metricSubmitter metric.IMetricSubmitter
}

// Constructor
func NewRegistryClient(instrumentationProvider instrumentation.IInstrumentationProvider) *RegistryClient {
	return &RegistryClient{
		tracerProvider:  instrumentationProvider.GetTracerProvider("RegistryClient"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}
}

// GetDigest receives image reference string and calls crane digest api to get it's digest from registry
func (client *RegistryClient) GetDigest(imageRef string) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigest")
	tracer.Info("Received image:", "imageRef", imageRef)

	// Resolve digest
	digest, err := crane.Digest(imageRef)
	if err != nil {
		// Report error
		wrappedError := errors.Wrap(err, "RegistryClient.GetDigest:")
		tracer.Error(wrappedError, "")
		return "", wrappedError
	}

	// Log digest and return it
	tracer.Info("Resolved to:", "digest", digest)
	return digest, nil
}
