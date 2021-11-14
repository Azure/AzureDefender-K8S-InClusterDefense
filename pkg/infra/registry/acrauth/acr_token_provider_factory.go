package acrauth

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
	"time"
)

type ACRTokenProviderFactory struct {
}

func NewACRTokenProviderFactory () *ACRTokenProviderFactory{
	return &ACRTokenProviderFactory{}
}

func (factory *ACRTokenProviderFactory) CreateACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, tokenExchanger IACRTokenExchanger, azureBearerAuthorizerTokenProvider azureauth.IBearerAuthorizerTokenProvider, cacheClient cache.ICacheClient, acrTokenProviderConfiguration *ACRTokenProviderConfiguration) (*ACRTokenProvider, error){
	armTokenCacheExpirationTime, err := time.ParseDuration(acrTokenProviderConfiguration.ArmTokenCacheExpirationTime)
	if err != nil{
		return nil, errors.Wrapf(err, "Given acrTokenProviderConfiguration.armTokenCacheExpirationTime string is not a valid time duration")
	}
	registryRefreshTokenCacheExpirationTime, err := time.ParseDuration(acrTokenProviderConfiguration.RegistryRefreshTokenCacheExpirationTime)
	if err != nil{
		return nil, errors.Wrapf(err, "Given acrTokenProviderConfiguration.registryRefreshTokenCacheExpirationTime string is not a valid time duration")
	}
	acrTokenProviderCacheClient := &ACRTokenProviderCacheClient{
		cacheClient:                             cacheClient,
		armTokenCacheExpirationTime:             armTokenCacheExpirationTime,
		registryRefreshTokenCacheExpirationTime: registryRefreshTokenCacheExpirationTime,
	}
	return NewACRTokenProvider(instrumentationProvider, tokenExchanger, azureBearerAuthorizerTokenProvider, acrTokenProviderCacheClient), nil
}
