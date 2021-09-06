package crane

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/acrauth"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
)

// IACRKeychainFactory responsible to create an ACR auth based keychain to authenticate to registry
type IACRKeychainFactory interface {
	// Create is creating  an ACR auth based keychain to registry using provided registry
	Create(registry string) (authn.Keychain, error)
}

// ACRKeychainFactory basic implementation of
type ACRKeychainFactory struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter  metric.IMetricSubmitter
	// acrTokenProvider provides an ACR token
	acrTokenProvider acrauth.IACRTokenProvider
}

// NewACRKeychainFactory Ctor
func NewACRKeychainFactory(instrumentationProvider instrumentation.IInstrumentationProvider, acrTokenProvider acrauth.IACRTokenProvider) *ACRKeychainFactory {
	return &ACRKeychainFactory{
		tracerProvider:   instrumentationProvider.GetTracerProvider("ACRKeychainFactory"),
		metricSubmitter:  instrumentationProvider.GetMetricSubmitter(),
		acrTokenProvider: acrTokenProvider,
	}
}

// Create creating  an ACR auth based keychain to registry using provided registry
func (factory *ACRKeychainFactory) Create(registry string) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "registry", registry)

	// Get a refresh token for registry
	accessToken, err := factory.acrTokenProvider.GetACRRefreshToken(registry)
	if err != nil {
		err = errors.Wrap(err, "ACRKeychainFactory.Create: failed on GetACRRefreshToken")
		tracer.Error(err, "")
		return nil, err
	}

	//Create an ACR keychain
	return &ACRKeyChain{
		Token: accessToken,
	}, nil

}


// ACRKeyChain represents an ACR based keychain
type ACRKeyChain struct {
	Token string `json:"token"`
}
// Resolve Implements keychain required function, check if registry is ACR or not to decide it to return the auth with token
// or anonymous (non acr dns suffix -> anonymous, otherwise -> IdenitytToken(refreshtoken based) auth bosed)
func (b *ACRKeyChain) Resolve(resource authn.Resource) (authn.Authenticator, error) {
	if !registryutils.IsRegistryEndpointACR(resource.RegistryStr()) {
		return authn.Anonymous, nil
	}
	return authn.FromConfig(authn.AuthConfig{
		// Identity token assigment specify it's a refresh token based auth - OAuth2
		IdentityToken: b.Token,
	}), nil
}
