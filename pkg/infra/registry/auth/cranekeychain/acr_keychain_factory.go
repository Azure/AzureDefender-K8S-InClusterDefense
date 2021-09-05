package cranekeychain

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth/azure"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
)

type IACRKeychainFactory interface {
	Create(loginServer string) (authn.Keychain, error)
}

type ACRKeychainFactory struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter  metric.IMetricSubmitter
	acrTokenProvider azure.IACRTokenProvider
}

func NewACRKeychainFactory(instrumentationProvider instrumentation.IInstrumentationProvider, acrTokenProvider azure.IACRTokenProvider) *ACRKeychainFactory {
	return &ACRKeychainFactory{
		tracerProvider:   instrumentationProvider.GetTracerProvider("ACRKeychainFactory"),
		metricSubmitter:  instrumentationProvider.GetMetricSubmitter(),
		acrTokenProvider: acrTokenProvider,
	}
}

func (factory *ACRKeychainFactory) Create(loginServer string) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "loginServer", loginServer)

	accessToken, err := factory.acrTokenProvider.GetACRTokenFromARMToken(loginServer)
	if err != nil {
		err = errors.Wrap(err, "ACRKeychainFactory.Create: failed on GetACRTokenFromARMToken")
		tracer.Error(err, "")
		return nil, err
	}
	return &ACRKeyChain{
		Token: accessToken,
	}, nil

}

type ACRKeyChain struct {
	Token string `json:"token"`
}

func (b *ACRKeyChain) Resolve(resource authn.Resource) (authn.Authenticator, error) {
	if !registryutils.IsRegistryEndpointACR(resource.RegistryStr()) {
		return authn.Anonymous, nil
	}
	return authn.FromConfig(authn.AuthConfig{
		IdentityToken: b.Token,
	}), nil
}
