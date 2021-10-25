package acrauth

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
)

// IACRTokenProvider responsible to provide a token to ACR registry
type IACRTokenProvider interface {
	// GetACRRefreshToken provide a refresh token (used for generating access-token to registry data plane)
	// for registry provided
	GetACRRefreshToken(registry string) (string, error)
}

// ACRTokenProvider implements IACRTokenProvider interface
var _ IACRTokenProvider = (*ACRTokenProvider)(nil)

// ACRTokenProvider azure based implementation of IACRTokenProvider
type ACRTokenProvider struct {
	// tracerProvider providing tracers
	tracerProvider trace.ITracerProvider
	// metricSubmitter submits metrics for class
	metricSubmitter metric.IMetricSubmitter
	// azureBearerAuthorizer is a bearer based authorizer
	azureBearerAuthorizer azureauth.IBearerAuthorizer
	// tokenExchanger is exchanger to exchange the bearer token to a refresh token
	tokenExchanger IACRTokenExchanger
	// tokenCache is cache for mapping acr registry to token
	tokenCache cache.ICacheClient
}

// NewACRTokenProvider Ctor
func NewACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, tokenExchanger IACRTokenExchanger, azureBearerAuthorizer azureauth.IBearerAuthorizer, tokenCache cache.ICacheClient) *ACRTokenProvider {
	return &ACRTokenProvider{
		tracerProvider:        instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		metricSubmitter:       instrumentationProvider.GetMetricSubmitter(),
		azureBearerAuthorizer: azureBearerAuthorizer,
		tokenExchanger:        tokenExchanger,
		tokenCache:            tokenCache,
	}
}

// GetACRRefreshToken provides a refresh token (used for generating access-token to registry data plane)
//  for registry provided.
// Refersh and extract ARM token from azure authorizer, then exchange it to refersh token using token exchanger
func (tokenProvider *ACRTokenProvider) GetACRRefreshToken(registry string) (string, error) {
	tracer := tokenProvider.tracerProvider.GetTracer("GetACRRefreshToken")
	tracer.Info("Received", "registry", registry)

	// First check if we can get digest from cache
	token, keyDontExistErr := tokenProvider.tokenCache.Get(registry)
	if keyDontExistErr == nil { // If key exist - return digest
		tracer.Info("Token exist in cache", "registry", registry)
		return token, nil
	}else {
		tracer.Info("Token don't exist in cache", "registry", registry)
	}

	// Refresh token if needed
	err := azureauth.RefreshBearerAuthorizer(tokenProvider.azureBearerAuthorizer, context.Background())
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.azureauth.RefreshBearerAuthorizer: failed to refresh")
		tracer.Error(err, "")
		return "", err
	}
	armToken := tokenProvider.azureBearerAuthorizer.TokenProvider().OAuthToken()

	// Exchange arm token to ACR refresh token
	registryRefreshToken, err := tokenProvider.tokenExchanger.ExchangeACRAccessToken(registry, armToken)
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.tokenExchanger.ExchangeACRAccessToken: failed")
		tracer.Error(err, "")
		return "", err
	}

	// Save in cache
	err = tokenProvider.tokenCache.Set(registry, registryRefreshToken, 0)
	if err != nil{
		err = errors.Wrap(err, "GetACRRefreshToken: Failed to set token in cache")
		tracer.Error(err, "")
		//return digest, err
	}

	// TODO add caching + experation
	return registryRefreshToken, nil
}
