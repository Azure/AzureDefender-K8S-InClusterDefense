package acrauth

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
)

type IACRTokenProvider interface {
	GetACRTokenFromARMToken(registry string) (string, error)
}

type ACRTokenProvider struct {
	tracerProvider        trace.ITracerProvider
	metricSubmitter       metric.IMetricSubmitter
	azureBearerAuthorizer azureauth.IBearerAuthorizer
	tokenExchanger IACRTokenExchanger
}

func NewACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, tokenExchanger IACRTokenExchanger ,azureBearerAuthorizer azureauth.IBearerAuthorizer) *ACRTokenProvider {
	return &ACRTokenProvider{
		tracerProvider:        instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		metricSubmitter:       instrumentationProvider.GetMetricSubmitter(),
		azureBearerAuthorizer: azureBearerAuthorizer,
		tokenExchanger: tokenExchanger,
	}
}

func (tokenProvider *ACRTokenProvider) GetACRTokenFromARMToken(registry string) (string, error) {
	tracer := tokenProvider.tracerProvider.GetTracer("GetACRTokenFromARMToken")
	tracer.Info("Received", "registry", registry)

	// Refresh token if needed
	err := azureauth.RefreshBearerAuthorizer(tokenProvider.azureBearerAuthorizer, context.Background())
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.azureauth.RefreshBearerAuthorizer: failed to refresh")
		tracer.Error(err, "")
		return "", err
	}
	armToken := tokenProvider.azureBearerAuthorizer.TokenProvider().OAuthToken()
	registryRefreshToken, err := tokenProvider.tokenExchanger.ExchangeACRAccessToken(registry, armToken)
	if err != nil {
		err = errors.Wrap(err, "ACRTokenProvider.tokenExchanger.ExchangeACRAccessToken: failed")
		tracer.Error(err, "")
		return "", err
	}

	// add caching? experation?
	return registryRefreshToken, nil
}
