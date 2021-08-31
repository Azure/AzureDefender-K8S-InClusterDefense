package cranekeychain

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/google/go-containerregistry/pkg/authn"
)

type IACRKeychainFactory interface {
	Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error)
}

type ACRKeychainFactory struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
}

func NewACRKeychainFactory(instrumentationProvider instrumentation.IInstrumentationProvider) *ACRKeychainFactory {
	return &ACRKeychainFactory{
		tracerProvider:  instrumentationProvider.GetTracerProvider("ACRKeychainFactory"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}
}

func (factory *ACRKeychainFactory) Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "namespace", namespace, "imagePullSecrets", imagePullSecrets, "serviceAccountName", serviceAccountName)

	return nil, nil
}
