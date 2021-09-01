package azure

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

type IACRTokenProvider interface {
}

type ACRTokenProvider struct {
	tracerProvider        trace.ITracerProvider
	metricSubmitter       metric.IMetricSubmitter
	azureBearerAuthorizer azureauth.IBearerAuthorizer
}

func NewACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, azureBearerAuthorizer azureauth.IBearerAuthorizer) *ACRTokenProvider {
	return &ACRTokenProvider{
		//tracerProvider:        instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		//metricSubmitter:       instrumentationProvider.GetMetricSubmitter(),
		azureBearerAuthorizer: azureBearerAuthorizer,
	}
}

func (tokenProvider *ACRTokenProvider) GetACRTokenFromARMToken(loginServer string) (string, error) {
	//tracer := tokenProvider.tracerProvider.GetTracer("GetACRTokenFromARMToken")
	//tracer.Info("Received", "loginServer", loginServer)

	// Refresh token if needed
	err := azureauth.RefreshBearerAuthorizer(tokenProvider.azureBearerAuthorizer, context.Background())
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.azureauth.RefreshBearerAuthorizer: failed to refresh")
		//tracer.Error(err, "")
		return "", err
	}
	armToken := tokenProvider.azureBearerAuthorizer.TokenProvider().OAuthToken()
	loginServer = "tomerwdevopsstage.azurecr.io"
	token, err := ExchangeACRAccessToken2(loginServer, armToken)
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.ExchangeACRAccessToken: failed")
		//tracer.Error(err, "")
		return "", err
	}
	kc := &BearerIdentityKeyChain{Token: token}

	registryRefreshToken, err := crane.Digest(loginServer+"/alpine:targettag", crane.WithAuthFromKeychain(kc))

	//registryRefreshToken, err := ExchangeACRAccessToken(loginServer, tokenProvider.azureBearerAuthorizer.TokenProvider())
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.ExchangeACRAccessToken: failed")
		//tracer.Error(err, "")
		return "", err
	}

	//tracer.Info("Succeeded to ger registry token", "loginServer", loginServer)
	return registryRefreshToken, nil
}

type BearerIdentityKeyChain struct {
	Token string `json:"token"`
}

func (b *BearerIdentityKeyChain) Resolve(authn.Resource) (authn.Authenticator, error) {
	return authn.FromConfig(authn.AuthConfig{
		IdentityToken: b.Token,
	}), nil
}
