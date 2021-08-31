package registry

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth/cranekeychain"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

const (
	userAgent = "azdproxy"
)

// CraneRegistryClient container registry based client
type CraneRegistryClient struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// craneWrapper is the  wrappers.ICraneWrapper that wraps cranekeychain mod
	craneWrapper wrappers.ICraneWrapper
	// multiKeychainFactoryfactory is a factory to create a key chain for registry auth
	multiKeychainFactoryfactory cranekeychain.IMultiKeychainFactory
}

// NewCraneRegistryClient Constructor for the registry client
func NewCraneRegistryClient(instrumentationProvider instrumentation.IInstrumentationProvider, craneWrapper wrappers.ICraneWrapper, multiKeychainFactoryfactory cranekeychain.IMultiKeychainFactory) *CraneRegistryClient {
	return &CraneRegistryClient{
		tracerProvider:              instrumentationProvider.GetTracerProvider("CraneRegistryClient"),
		metricSubmitter:             instrumentationProvider.GetMetricSubmitter(),
		craneWrapper:                craneWrapper,
		multiKeychainFactoryfactory: multiKeychainFactoryfactory,
	}
}

// GetDigest receives image reference string and calls cranekeychain digest api to get it's digest from registry
func (client *CraneRegistryClient) GetDigest(imageRef string, authContext *auth.AuthContext) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigest")
	tracer.Info("Received image:", "imageRef", imageRef)

	// First check if we can extract it from ref it self (digest based ref)
	isDigestBasedImageRef, digest, err := TryExtractDigestFromImageRef(imageRef)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "utils.TryExtractDigestFromImageRef:")
		tracer.Error(err, "")
		return "", err
	}

	tracer.Info("utils.TryExtractDigestFromImageRef return values", "isDigestBasedImageRef", isDigestBasedImageRef, "digest", digest)
	if isDigestBasedImageRef {
		// Return digest extracted from ref
		return digest, nil
	}

	// TODO add retry policy
	keychain, err := client.multiKeychainFactoryfactory.Create(authContext)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "multiKeychainFactoryfactory.Create:")
		tracer.Error(err, "")
		return "", err
	}

	// TODO add retry policy
	// Resolve digest
	digest, err = client.craneWrapper.Digest(imageRef, crane.WithAuthFromKeychain(keychain), crane.WithUserAgent(userAgent))
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
