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

// CraneRegistryClient crane based implementation of the registry client interface registry.IRegistryClient
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

// GetDigestUsingDefaultAuth receives image reference and get it's digest using the default docker config auth
func (client *CraneRegistryClient) GetDigestUsingDefaultAuth(imageReference registry.IImageReference) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigestUsingDefaultAuth")
	tracer.Info("Received image:", "imageReference", imageReference)

	// Argument validation
	if imageReference == nil{
		err := errors.Wrap(nilArgError, "CraneRegistryClient.GetDigestUsingDefaultAuth")
		tracer.Error(err,"")
		return "", err
	}

	// Get digest
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

// GetDigestUsingACRAttachAuth receives image reference and get it's digest using ACR attach authntication
// ACR attach auth is based MSI token used to access the registry
// Authenticate with multikeychain with acrkeychain and default keychain
func (client *CraneRegistryClient) GetDigestUsingACRAttachAuth(imageReference registry.IImageReference) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigestUsingACRAttachAuth")
	tracer.Info("Received image:", "imageReference", imageReference)

	// Argument validation
	if imageReference == nil{
		err := errors.Wrap(nilArgError, "CraneRegistryClient.GetDigestUsingACRAttachAuth")
		tracer.Error(err,"")
		return "", err
	}

	// Create ACR auth keychain to registry - keychain with refresh token
	acrKeyChain, err := client.acrKeychainFactory.Create(imageReference.Registry())
	if err != nil {
		err = errors.Wrap(err, "CraneRegistryClient.GetDigestUsingACRAttachAuth: could not create acrKeychain")
		tracer.Error(err, "")
		return "", err
	}

	// Get digest and passing the keychain
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

// GetDigestUsingK8SAuth receives image reference and get it's digest using K8S secerts and auth
// K8S auth is based image pull secrets used in deployment or attached to service account to pull the image
// Authenticate with multikeychain with k8skeychain and default keychain
func (client *CraneRegistryClient) GetDigestUsingK8SAuth(imageReference registry.IImageReference, namespace string , imagePullSecrets []string, serviceAccountName string) (string, error) {
	tracer := client.tracerProvider.GetTracer("GetDigestUsingK8SAuth")
	tracer.Info("Received image:", "imageReference", imageReference, "namespace", namespace, "imagePullSecrets", imagePullSecrets, "serviceAccountName", serviceAccountName)

	// Argument validaiton
	if imageReference == nil{
		err := errors.Wrap(nilArgError, "CraneRegistryClient.GetDigestUsingK8SAuth")
		tracer.Error(err,"")
		return "", err
	}

	// Create K8S keychain
	k8sKeychain, err := client.k8sKeychainFactory.Create(namespace, imagePullSecrets, serviceAccountName)
	if err != nil {
		err = errors.Wrap(err, "CraneRegistryClient.GetDigestUsingK8SAuth: could not create k8sKeychain")
		tracer.Error(err, "")
		return "", err
	}

	// Get digest and passing key chain
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


// getDigest private function that receives imageReference and a keychain, it wraps keychain received with multikeychain and appends the defaultkeychain as well
// Then calls crane Digest fucntion using the multikeychian created and the client userAgent
func (client *CraneRegistryClient) getDigest(imageReference registry.IImageReference, keychain authn.Keychain) (string, error) {
	tracer := client.tracerProvider.GetTracer("getDigest")
	receivedKeyChainType := fmt.Sprintf("%T", keychain)
	tracer.Info("Received image:", "imageReference", imageReference.Original(), "receivedKeyChainType", receivedKeyChainType)


	// TODO add retry policy
	// Resolve digest using Options:
	//  - multikeychain of received keychain and the default keychain,
	// - userAgent of the client
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

