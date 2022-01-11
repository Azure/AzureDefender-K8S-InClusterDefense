package acrauth

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
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
	// azureBearerAuthorizerTokenProvider is a bearer based token provider
	azureBearerAuthorizerTokenProvider azureauth.IBearerAuthorizerTokenProvider
	// tokenExchanger is exchanger to exchange the bearer token to a refresh token
	tokenExchanger IACRTokenExchanger
	// cacheClient is cache for mapping acr registry to token
	cacheClient cache.ICacheClient
	// acrTokenProviderConfiguration is configuration data for ACRTokenProvider
	acrTokenProviderConfiguration *ACRTokenProviderConfiguration
}

// ACRTokenProviderConfiguration is configuration data for ACRTokenProvider
type ACRTokenProviderConfiguration struct {
	// RegistryRefreshTokenCacheExpirationTime is the expiration time **IN MINUTES** for registryRefreshToken in the cache client
	RegistryRefreshTokenCacheExpirationTime int
}

// NewACRTokenProvider Ctor
func NewACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, tokenExchanger IACRTokenExchanger, azureBearerAuthorizerTokenProvider azureauth.IBearerAuthorizerTokenProvider, cacheClient cache.ICacheClient, acrTokenProviderConfiguration *ACRTokenProviderConfiguration) *ACRTokenProvider {
	return &ACRTokenProvider{
		tracerProvider:                     instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		metricSubmitter:                    instrumentationProvider.GetMetricSubmitter(),
		azureBearerAuthorizerTokenProvider: azureBearerAuthorizerTokenProvider,
		tokenExchanger:                     tokenExchanger,
		cacheClient:                        cacheClient,
		acrTokenProviderConfiguration:      acrTokenProviderConfiguration,
	}
}

// GetACRRefreshToken provides a refresh token (used for generating access-token to registry data plane)
//  for registry provided.
// Refersh and extract ARM token from azure authorizer, then exchange it to refersh token using token exchanger
func (tokenProvider *ACRTokenProvider) GetACRRefreshToken(registry string) (string, error) {
	tracer := tokenProvider.tracerProvider.GetTracer("GetACRRefreshToken")
	tracer.Info("Received", "registry", registry)

	registryRefreshToken, err := tokenProvider.cacheClient.Get(registry)
	// Error as a result of key doesn't exist and error from the cache are treated the same (skip cache)
	if err != nil { // Couldn't get token from cache - skip and get results from provider
		err = errors.Wrap(err, "Couldn't get registryRefreshToken from cache")
		tracer.Error(err, "")
	} else { // If key exist - return token
		tracer.Info("registryRefreshToken exist in cache", "registry", registry)
		return registryRefreshToken, nil
	}

	// Otherwise, get azure token
	armToken, err := tokenProvider.azureBearerAuthorizerTokenProvider.GetOAuthToken(context.Background())
	if err != nil {
		err = errors.Wrap(err, "Failed to get armToken")
		tracer.Error(err, "")
		tokenProvider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ACRTokenProvider.GetACRRefreshToken"))
		return "", err
	}

	// Exchange arm token to ACR refresh token
	registryRefreshToken, err = tokenProvider.tokenExchanger.ExchangeACRAccessToken(registry, armToken)
	if err != nil {
		err = errors.Wrap(err, "Failed to exchange ACR access token")
		tracer.Error(err, "")
		tokenProvider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ACRTokenProvider.GetACRRefreshToken"))
		return "", err
	}

	// Save registryRefreshToken in cache
	go func() {
		err = tokenProvider.cacheClient.Set(registry, registryRefreshToken, utils.GetMinutes(tokenProvider.acrTokenProviderConfiguration.RegistryRefreshTokenCacheExpirationTime))
		if err != nil {
			err = errors.Wrap(err, "Failed to set registryRefreshToken in cache")
			tracer.Error(err, "")
			tokenProvider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ACRTokenProvider.GetACRRefreshToken"))
		} else {
			tracer.Info("Set registryRefreshToken in cache successfully", "registry", registry)
		}
	}()

	return registryRefreshToken, nil
}
