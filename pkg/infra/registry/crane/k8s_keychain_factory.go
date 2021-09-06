package crane

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
)

type IK8SKeychainFactory interface {
	Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error)
}

type K8SKeychainFactory struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// Kubernetes client
	client kubernetes.Interface
}

func NewK8SKeychainFactory(instrumentationProvider instrumentation.IInstrumentationProvider, client kubernetes.Interface) *K8SKeychainFactory {
	return &K8SKeychainFactory{
		tracerProvider:  instrumentationProvider.GetTracerProvider("K8SKeychainFactory"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		client:          client,
	}
}

func (factory *K8SKeychainFactory) Create(namespace string, imagePullSecrets []string, serviceAccountName string) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "namespace", namespace, "imagePullSecrets", imagePullSecrets, "serviceAccountName", serviceAccountName)

	// TODO add support to not fail on non existant SA or Pull secret
	// TODO this will fail if pull secrets does not exists or SA is not accessibile - need to add a fallback to try to skip this if it fails
	return k8schain.New(context.Background(), factory.client, k8schain.Options{Namespace: namespace, ServiceAccountName: serviceAccountName, ImagePullSecrets: imagePullSecrets})
}
