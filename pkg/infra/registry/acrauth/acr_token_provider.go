package acrauth

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
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

// ACRTokenProvider azure based implementation of IACRTokenProvider
type ACRTokenProvider struct {
	// tracerProvider providing tracers
	tracerProvider        trace.ITracerProvider
	// metricSubmitter submits metrics for class
	metricSubmitter       metric.IMetricSubmitter
	// azureBearerAuthorizer is a bearer based authorizer
	azureBearerAuthorizer azureauth.IBearerAuthorizer
	// tokenExchanger is exchanger to exchange the bearer token to a refresh token
	tokenExchanger IACRTokenExchanger
}

// NewACRTokenProvider Ctor
func NewACRTokenProvider(instrumentationProvider instrumentation.IInstrumentationProvider, tokenExchanger IACRTokenExchanger ,azureBearerAuthorizer azureauth.IBearerAuthorizer) *ACRTokenProvider {
	return &ACRTokenProvider{
		tracerProvider:        instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		metricSubmitter:       instrumentationProvider.GetMetricSubmitter(),
		azureBearerAuthorizer: azureBearerAuthorizer,
		tokenExchanger: tokenExchanger,
	}
}

// GetACRRefreshToken provides a refresh token (used for generating access-token to registry data plane)
//  for registry provided.
// Refersh and extract ARM token from azure authorizer, then exchange it to refersh token using token exchanger
func (tokenProvider *ACRTokenProvider) GetACRRefreshToken(registry string) (string, error) {
	tracer := tokenProvider.tracerProvider.GetTracer("GetACRRefreshToken")
	tracer.Info("Received", "registry", registry)

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

	// TODO add caching + experation
	return registryRefreshToken, nil
}
