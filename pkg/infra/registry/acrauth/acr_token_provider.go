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

const(
	// _armTokenCacheKeyPrefix is the prefix for ARMToken cache keys
	_armTokenCacheKeyPrefix = "armToken"
	// _registryRefreshTokenCacheKeyPrefix is the prefix for RegistryRefreshToken cache keys
	_registryRefreshTokenCacheKeyPrefix = "registryRefreshToken"
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
	// acrTokenProviderCacheClient is the cache ACRTokenProvider uses and its time expiration
	acrTokenProviderCacheClient *ACRTokenProviderCacheClient
}

// ACRTokenProviderConfiguration is configuration data for ACRTokenProvider
type ACRTokenProviderConfiguration struct {
	// ArmTokenCacheExpirationTime is the expiration time **in ms** for armToken in the cache client
	ArmTokenCacheExpirationTime string
	// RegistryRefreshTokenCacheExpirationTime is the expiration time **in ms** for registryRefreshToken in the cache client
	RegistryRefreshTokenCacheExpirationTime string
}

// ACRTokenProviderCacheClient is a cache client for ACRTokenProvider that contains the needed cache client configurations
type ACRTokenProviderCacheClient struct {
	// cacheClient is cache for mapping acr registry to token
	cacheClient cache.ICacheClient
	// armTokenCacheExpirationTime is the expiration time **in ms** for armToken in the cache client
	armTokenCacheExpirationTime time.Duration
	// registryRefreshTokenCacheExpirationTime is the expiration time **in ms** for registryRefreshToken in the cache client
	registryRefreshTokenCacheExpirationTime time.Duration
}

// NewACRTokenProvider Ctor
func NewACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, tokenExchanger IACRTokenExchanger, azureBearerAuthorizerTokenProvider azureauth.IBearerAuthorizerTokenProvider, acrTokenProviderCacheClient *ACRTokenProviderCacheClient) *ACRTokenProvider {
	return &ACRTokenProvider{
		tracerProvider:                     instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		metricSubmitter:                    instrumentationProvider.GetMetricSubmitter(),
		azureBearerAuthorizerTokenProvider: azureBearerAuthorizerTokenProvider,
		tokenExchanger:                     tokenExchanger,
		acrTokenProviderCacheClient: acrTokenProviderCacheClient,
	}
}

// GetACRRefreshToken provides a refresh token (used for generating access-token to registry data plane)
//  for registry provided.
// Refersh and extract ARM token from azure authorizer, then exchange it to refersh token using token exchanger
func (tokenProvider *ACRTokenProvider) GetACRRefreshToken(registry string) (string, error) {
	tracer := tokenProvider.tracerProvider.GetTracer("GetACRRefreshToken")
	tracer.Info("Received", "registry", registry)

	// First check if we can get registryRefreshToken from cache
	registryRefreshTokenCacheKey := tokenProvider.getRegistryRefreshTokenCacheKey(registry)
	registryRefreshToken, err := tokenProvider.acrTokenProviderCacheClient.cacheClient.Get(registryRefreshTokenCacheKey)
	// Error as a result of key doesn't exist and error from the cache are treated the same (skip cache)
	if err == nil { // If key exist - return token
		tracer.Info("registryRefreshToken exist in cache", "registry", registry)
		return registryRefreshToken, nil
	}
	tracer.Info("registryRefreshToken don't exist in cache", "registry", registry)

	// Otherwise, get azure token
	armToken, err := tokenProvider.getARMToken(registry)
	if err != nil {
		err = errors.Wrap(err, "Failed to get armToken")
		tracer.Error(err, "")
		return "", err
	}

	// Exchange arm token to ACR refresh token
	registryRefreshToken, err = tokenProvider.tokenExchanger.ExchangeACRAccessToken(registry, armToken)
	if err != nil {
		err = errors.Wrap(err, "Failed to exchange ACR access token")
		tracer.Error(err, "")
		return "", err
	}

	// Save registryRefreshToken in cache
	err = tokenProvider.acrTokenProviderCacheClient.cacheClient.Set(registryRefreshTokenCacheKey, registryRefreshToken, tokenProvider.acrTokenProviderCacheClient.registryRefreshTokenCacheExpirationTime)
	if err != nil{
		err = errors.Wrap(err, "Failed to set registryRefreshToken in cache")
		tracer.Error(err, "")
	}
	tracer.Info("Set registryRefreshToken in cache", "registry", registry)

	return registryRefreshToken, nil
}

// getARMToken gets ARM token by first trying to get the token from the cache. If fails, gets from azure.
func (tokenProvider *ACRTokenProvider) getARMToken(registry string) (string, error) {
	tracer := tokenProvider.tracerProvider.GetTracer("getARMToken")

	armTokenCacheKey := tokenProvider.getARMTokenCacheKey(registry)
	// First check if we can get armToken from cache
	armToken, err := tokenProvider.acrTokenProviderCacheClient.cacheClient.Get(armTokenCacheKey)
	// Error as a result of key doesn't exist and error from the cache are treated the same (skip cache)
	if err == nil { // If key exist - return token
		tracer.Info("ARMToken exist in cache", "registry", registry)
		return armToken, nil
	}
	tracer.Info("ARMToken don't exist in cache", "registry", registry)

	// Get azure token
	armToken, err = tokenProvider.azureBearerAuthorizerTokenProvider.GetOAuthToken(context.Background())
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.azureauth.azureBearerAuthorizerTokenProvider: failed")
		tracer.Error(err, "")
		return "", err
	}

	// Save ARMToken in cache
	err = tokenProvider.acrTokenProviderCacheClient.cacheClient.Set(armTokenCacheKey, armToken, tokenProvider.acrTokenProviderCacheClient.armTokenCacheExpirationTime)
	if err != nil{
		err = errors.Wrap(err, "Failed to set armToken in cache")
		tracer.Error(err, "")
	}
	tracer.Info("Set armToken in cache", "registry", registry)

	return armToken, nil
}

// getARMTokenCacheKey returns the ARMToken cache key of a given registry
func (tokenProvider *ACRTokenProvider) getARMTokenCacheKey(registry string) string {
	return _armTokenCacheKeyPrefix + registry
}

// getARMTokenCacheKey returns the RegistryRefreshToken cache key of a given registry
func (tokenProvider *ACRTokenProvider) getRegistryRefreshTokenCacheKey (registry string) string {
	return _registryRefreshTokenCacheKeyPrefix + registry
}
