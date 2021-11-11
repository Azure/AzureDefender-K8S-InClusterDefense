package acrauth

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
	"time"
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
	// azureBearerAuthorizerTokenProvider is a bearer based token provider
	azureBearerAuthorizerTokenProvider azureauth.IBearerAuthorizerTokenProvider
	// tokenExchanger is exchanger to exchange the bearer token to a refresh token
	tokenExchanger IACRTokenExchanger
	// tokenCache is cache for mapping acr registry to token
	cacheClient cache.ICacheClient
	// acrTokenProviderConfiguration is configuration data for ACRTokenProvider
	acrTokenProviderConfiguration *ACRTokenProviderConfiguration

}

// ACRTokenProviderConfiguration is configuration data for ACRTokenProvider
type ACRTokenProviderConfiguration struct {
	// CacheExpirationTime is the expiration time **in seconds** for tokens in the cache client
	CacheExpirationTime time.Duration
}

// NewACRTokenProvider Ctor
func NewACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, tokenExchanger IACRTokenExchanger, azureBearerAuthorizerTokenProvider azureauth.IBearerAuthorizerTokenProvider, cacheClient cache.ICacheClient, configuration *ACRTokenProviderConfiguration) *ACRTokenProvider {
	return &ACRTokenProvider{
		tracerProvider:                     instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		metricSubmitter:                    instrumentationProvider.GetMetricSubmitter(),
		azureBearerAuthorizerTokenProvider: azureBearerAuthorizerTokenProvider,
		tokenExchanger:                     tokenExchanger,
		cacheClient:                        cacheClient,
		acrTokenProviderConfiguration: configuration,
	}
}

// GetACRRefreshToken provides a refresh token (used for generating access-token to registry data plane)
//  for registry provided.
// Refersh and extract ARM token from azure authorizer, then exchange it to refersh token using token exchanger
func (tokenProvider *ACRTokenProvider) GetACRRefreshToken(registry string) (string, error) {
	tracer := tokenProvider.tracerProvider.GetTracer("GetACRRefreshToken")
	tracer.Info("Received", "registry", registry)

	// First check if we can get digest from cache
	token, keyDontExistErr := tokenProvider.cacheClient.Get(registry)
	if keyDontExistErr == nil { // If key exist - return digest
		tracer.Info("Token exist in cache", "registry", registry)
		return token, nil
	}
	tracer.Info("Token don't exist in cache", "registry", registry)


	// Get azure token
	armToken, err := tokenProvider.azureBearerAuthorizerTokenProvider.GetOAuthToken(context.Background())
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.azureauth.azureBearerAuthorizerTokenProvider: failed")
		tracer.Error(err, "")
		return "", err
	}

	// Exchange arm token to ACR refresh token
	registryRefreshToken, err := tokenProvider.tokenExchanger.ExchangeACRAccessToken(registry, armToken)
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.tokenExchanger.ExchangeACRAccessToken: failed")
		tracer.Error(err, "")
		return "", err
	}

	// Save in cache
	err = tokenProvider.cacheClient.Set(registry, registryRefreshToken, tokenProvider.acrTokenProviderConfiguration.CacheExpirationTime)
	if err != nil{
		err = errors.Wrap(err, "GetACRRefreshToken: Failed to set token in cache")
		tracer.Error(err, "")
	}
	tracer.Info("Set token in cache", "registry", registry)

	return registryRefreshToken, nil
}
