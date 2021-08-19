package registry

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
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
	craneWrapper    wrappers.ICraneWrapper
}

// Constructor
func NewRegistryClient(instrumentationProvider instrumentation.IInstrumentationProvider, craneWrapper wrappers.ICraneWrapper) *RegistryClient {
	return &RegistryClient{
		tracerProvider:  instrumentationProvider.GetTracerProvider("RegistryClient"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		craneWrapper:    craneWrapper,
	}
}

// GetDigest receives image reference string and calls crane digest api to get it's digest from registry
func (client *RegistryClient) GetDigest(imageRef string) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigest")
	tracer.Info("Received image:", "imageRef", imageRef)

	// Resolve digest
	digest, err := client.craneWrapper.Digest(imageRef)
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
