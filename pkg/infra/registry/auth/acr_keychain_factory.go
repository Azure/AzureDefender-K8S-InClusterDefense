package auth

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/google/go-containerregistry/pkg/authn"
)

type IACRKeyChainFactory interface {
	Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error)
}

type ACRKeyChainFactory struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// token exchanger
	tokenExchanger ACRTokenExchanger
}

func NewACRKeyChainFactory(instrumentationProvider instrumentation.IInstrumentationProvider) *ACRKeyChainFactory {
	return &ACRKeyChainFactory{
		tracerProvider:  instrumentationProvider.GetTracerProvider("ACRKeyChainFactory"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}
}

func (factory *ACRKeyChainFactory) Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "namespace", namespace, "imagePullSecrets", imagePullSecrets, "serviceAccountName", serviceAccountName)

	return nil, nil
}
