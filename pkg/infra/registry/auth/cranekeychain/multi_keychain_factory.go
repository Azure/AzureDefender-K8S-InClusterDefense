package cranekeychain

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
)

var (
	nullArgError = errors.New("NilArgArgument")
	unsupportedAuthType = errors.New("unsupportedAuthType")
)

type IMultiKeychainFactory interface {
	Create(ctx *auth.AuthConfig) (authn.Keychain, error)
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

func (factory *MultiKeychainFactory) Create(authCfg *auth.AuthConfig) (authn.Keychain, error) {
	tracer := factory.tracerProvider.GetTracer("Create")
	tracer.Info("Received:", "ctx", authCfg)

	if authCfg == nil  || authCfg.Context == nil {
		err := errors.Wrapf(nullArgError, "MultiKeychainFactory.Create: arg-> %v", authCfg)
		tracer.Error(err, "")
		return nil, err
	}

	var kc authn.Keychain = nil
	var err error = nil
	switch authCfg.AuthType {
	case auth.ACRAuth:
		kc, err = factory.acrKeychainFactory.Create(authCfg.Context.RegistryEndpoint)
		if err != nil {
			err = errors.Wrap(err, "MultiKeychainFactory.Create: could not create acrKeychain")
			tracer.Error(err, "")
			return nil, err
		}
	case auth.K8SAuth:

		kc, err = factory.k8sKeychainFactory.Create(authCfg.Context.Namespace, authCfg.Context.ImagePullSecrets, authCfg.Context.ServiceAccountName)
		if err != nil {
			err = errors.Wrap(err, "MultiKeychainFactory.Create: could not create k8schain")
			tracer.Error(err, "")
			return nil, err
		}
	default:
		err = errors.Wrapf(err, "MultiKeychainFactory.Create: unsupportedAuthType: %v", authCfg.AuthType)
		tracer.Error(err, "")
		return nil, err
	}

	// Add default key chain
	return authn.NewMultiKeychain(kc,authn.DefaultKeychain), nil
}
