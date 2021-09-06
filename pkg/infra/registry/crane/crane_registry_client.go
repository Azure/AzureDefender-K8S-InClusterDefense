package crane

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

const (
	userAgent = "azdproxy"
)

var (
	nilArgError = errors.New("NilArgErrorReceived")
)

// CraneRegistryClient container registry based client
type CraneRegistryClient struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// craneWrapper is the  wrappers.ICraneWrapper that wraps crane mod
	craneWrapper wrappers.ICraneWrapper

	acrKeychainFactory IACRKeychainFactory
	k8sKeychainFactory IK8SKeychainFactory
}

// NewCraneRegistryClient Constructor for the registry client
func NewCraneRegistryClient(instrumentationProvider instrumentation.IInstrumentationProvider, craneWrapper wrappers.ICraneWrapper, acrKeychainFactory IACRKeychainFactory, k8sKeychainFactory IK8SKeychainFactory) *CraneRegistryClient {
	return &CraneRegistryClient{
		tracerProvider:              instrumentationProvider.GetTracerProvider("CraneRegistryClient"),
		metricSubmitter:             instrumentationProvider.GetMetricSubmitter(),
		craneWrapper:                craneWrapper,
		acrKeychainFactory: acrKeychainFactory,
		k8sKeychainFactory: k8sKeychainFactory,
	}
}

func (client *CraneRegistryClient) GetDigestUsingDefaultAuth(imageReference registry.IImageReference) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigestUsingDefaultAuth")
	tracer.Info("Received image:", "imageReference", imageReference)

	if imageReference == nil{
		err := errors.Wrap(nilArgError, "CraneRegistryClient.GetDigestUsingDefaultAuth")
		tracer.Error(err,"")
		return "", err
	}

	digest, err := client.getDigest(imageReference, authn.DefaultKeychain)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "Failed with client.getDigest:")
		tracer.Error(err, "")
		return "", err
	}

	tracer.Info("Image resolved successfully", "imageRef", imageReference, "digest", digest)
	return digest, nil
}

func (client *CraneRegistryClient) GetDigestUsingACRAttachAuth(imageReference registry.IImageReference) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigestUsingACRAttachAuth")
	tracer.Info("Received image:", "imageReference", imageReference)

	if imageReference == nil{
		err := errors.Wrap(nilArgError, "CraneRegistryClient.GetDigestUsingACRAttachAuth")
		tracer.Error(err,"")
		return "", err
	}

	acrKeyChain, err := client.acrKeychainFactory.Create(imageReference.Registry())
	if err != nil {
		err = errors.Wrap(err, "CraneRegistryClient.GetDigestUsingACRAttachAuth: could not create acrKeychain")
		tracer.Error(err, "")
		return "", err
	}

	digest, err := client.getDigest(imageReference, acrKeyChain)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "Failed with client.getDigest:")
		tracer.Error(err, "")
		return "", err
	}

	tracer.Info("Image resolved successfully", "imageRef", imageReference, "digest", digest)
	return digest, nil
}

func (client *CraneRegistryClient) GetDigestUsingK8SAuth(imageReference registry.IImageReference, namespace string , imagePullSecrets []string, serviceAccountName string) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigestUsingK8SAuth")
	tracer.Info("Received image:", "imageReference", imageReference, "namespace", namespace, "imagePullSecrets", imagePullSecrets, "serviceAccountName", serviceAccountName)

	if imageReference == nil{
		err := errors.Wrap(nilArgError, "CraneRegistryClient.GetDigestUsingK8SAuth")
		tracer.Error(err,"")
		return "", err
	}

	k8sKeychain, err := client.k8sKeychainFactory.Create(namespace, imagePullSecrets, serviceAccountName)
	if err != nil {
		err = errors.Wrap(err, "CraneRegistryClient.GetDigestUsingK8SAuth: could not create k8sKeychain")
		tracer.Error(err, "")
		return "", err
	}


	digest, err := client.getDigest(imageReference, k8sKeychain)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "Failed with client.getDigest:")
		tracer.Error(err, "")
		return "", err
	}

	tracer.Info("Image resolved successfully", "imageRef", imageReference, "digest", digest)
	return digest, nil
}

func (client *CraneRegistryClient) getDigest(imageReference registry.IImageReference, keychain authn.Keychain) (string, error) {
	tracer := client.tracerProvider.GetTracer("getDigest")
	receivedKeyChainType := fmt.Sprintf("%T", keychain)
	tracer.Info("Received image:", "imageReference", imageReference.Original(), "receivedKeyChainType", receivedKeyChainType)


	// TODO add retry policy
	// Resolve digest
	digest, err := client.craneWrapper.Digest(imageReference.Original(), crane.WithAuthFromKeychain(authn.NewMultiKeychain(keychain, authn.DefaultKeychain)), crane.WithUserAgent(userAgent))
	if err != nil {
		// Report error
		err = errors.Wrapf(err, "CraneRegistryClient.getDigest with receivedKeyChainType %v", receivedKeyChainType)
		tracer.Error(err, "")

		// TODO return wrapped error type to handle on caller
		return "", err
	}

	// Log digest and return it
	tracer.Info("Image resolved successfully", "imageRef", imageReference.Original(), "digest", digest)
	return digest, nil
}

