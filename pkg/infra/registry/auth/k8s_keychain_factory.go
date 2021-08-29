package auth

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
)

type IK8SKeyChainFactory interface {
	Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error)
}

type K8SKeyChainFactory struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// Kubernetes client
	client kubernetes.Interface
}

func NewK8SKeyChainFactory(instrumentationProvider instrumentation.IInstrumentationProvider, client kubernetes.Interface) *K8SKeyChainFactory {
	return &K8SKeyChainFactory{
		tracerProvider:  instrumentationProvider.GetTracerProvider("K8SKeyChainFactory"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		client:          client,
	}
}

func (factory *K8SKeyChainFactory) Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "namespace", namespace, "imagePullSecrets", imagePullSecrets, "serviceAccountName", serviceAccountName)

	return k8schain.New(context.Background(), factory.client, k8schain.Options{Namespace: namespace, ServiceAccountName: serviceAccountName, ImagePullSecrets: imagePullSecrets})
}
