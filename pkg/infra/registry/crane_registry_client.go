package registry

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
	"github.com/pkg/errors"
)

// CraneRegistryClient container registry based client
type CraneRegistryClient struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// craneWrapper is the  wrappers.ICraneWrapper that wraps crane mod
	craneWrapper wrappers.ICraneWrapper
}

// NewCraneRegistryClient Constructor for the registry client
func NewCraneRegistryClient(instrumentationProvider instrumentation.IInstrumentationProvider, craneWrapper wrappers.ICraneWrapper) *CraneRegistryClient {
	return &CraneRegistryClient{
		tracerProvider:  instrumentationProvider.GetTracerProvider("CraneRegistryClient"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		craneWrapper:    craneWrapper,
	}
}

// GetDigest receives image reference string and calls crane digest api to get it's digest from registry
func (client *CraneRegistryClient) GetDigest(imageRef string) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigest")
	tracer.Info("Received image:", "imageRef", imageRef)

	// Resolve digest
	digest, err := client.craneWrapper.DigestWithRetry(imageRef, client.tracerProvider, client.metricSubmitter)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "CraneRegistryClient.GetDigest:")
		tracer.Error(err, "")

		// TODO return wrapped error type to handle on caller
		return "", err
	}

	// Log digest and return it
	tracer.Info("Image resolved successfully", "imageRef", imageRef, "digest", digest)
	return digest, nil
}
