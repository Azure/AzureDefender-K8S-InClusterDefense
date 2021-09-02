package cranekeychain

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
)

type IMultiKeychainFactory interface {
	Create(ctx *auth.AuthContext) (authn.Keychain, error)
}

type MultiKeychainFactory struct {
	// tracerProvider is the tracer provider for the registry client
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the registry client.
	metricSubmitter metric.IMetricSubmitter
	// k8sKeychainFactory is the factory to create a K8S key chain
	k8sKeychainFactory IK8SKeychainFactory

	acrKeychainFactory IACRKeychainFactory
}

func NewMultiKeychainFactory(instrumentationProvider instrumentation.IInstrumentationProvider, k8sKeychainfactory IK8SKeychainFactory, acrKeychainFactory IACRKeychainFactory) *MultiKeychainFactory {
	return &MultiKeychainFactory{
		tracerProvider:     instrumentationProvider.GetTracerProvider("MultiKeychainFactory"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		k8sKeychainFactory: k8sKeychainfactory,
		acrKeychainFactory: acrKeychainFactory,
	}
}

func (factory *MultiKeychainFactory) Create(ctx *auth.AuthContext) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "ctx", ctx)

	kcList := make([]authn.Keychain,0,3)

	if len(ctx.ImagePullSecrets) != 0 && ctx.ServiceAccountName != "" {
		k8sKeychain, err := factory.k8sKeychainFactory.Create(ctx.Namespace, ctx.ImagePullSecrets, ctx.ServiceAccountName)
		if err != nil {
			// Add fallback on unable to create
			err = errors.Wrap(err, "MultiKeychainFactory.Create: could not create k8schain")
			tracer.Error(err, "")
		} else {
			kcList = append(kcList, k8sKeychain)

		}
	}

	acrKeychain, err := factory.acrKeychainFactory.Create(ctx.RegistryEndpoint)
	if err != nil {
		// Add fallback on unable to create
		err = errors.Wrap(err, "MultiKeychainFactory.Create: could not create acrKeychain")
		tracer.Error(err, "")
	}else{
		kcList = append(kcList, acrKeychain)
	}

	kcList = append(kcList, authn.DefaultKeychain)

	return authn.NewMultiKeychain(kcList...), nil
}