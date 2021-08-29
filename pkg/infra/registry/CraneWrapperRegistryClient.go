package registry

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
	"github.com/pkg/errors"
)

// CraneWrapperRegistryClient container registry based client
type CraneWrapperRegistryClient struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// craneWrapper is the  wrappers.ICraneWrapper that wraps crane mod
	craneWrapper wrappers.ICraneWrapper
}

// NewCraneWrapperRegistryClient Constructor for the registry client
func NewCraneWrapperRegistryClient(instrumentationProvider instrumentation.IInstrumentationProvider, craneWrapper wrappers.ICraneWrapper) *CraneWrapperRegistryClient {
	return &CraneWrapperRegistryClient{
		tracerProvider:  instrumentationProvider.GetTracerProvider("CraneWrapperRegistryClient"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		craneWrapper:    craneWrapper,
	}
}

// GetDigest receives image reference string and calls crane digest api to get it's digest from registry
func (client *CraneWrapperRegistryClient) GetDigest(imageRef string) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigest")
	tracer.Info("Received image:", "imageRef", imageRef)

	// TODO add retry policy
	// Resolve digest
	digest, err := client.craneWrapper.Digest(imageRef)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "CraneWrapperRegistryClient.GetDigest:")
		tracer.Error(err, "")

		// TODO return wrapped error type to handle on caller
		return "", err
	}

	// Log digest and return it
	tracer.Info("Image resolved successfully", "imageRef", imageRef, "digest", digest)
	return digest, nil
}
