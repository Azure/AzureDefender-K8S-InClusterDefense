package webhook

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// IManagerFactory Factory to create manager.Manager
type IManagerFactory interface {
	// CreateManager Initialize the manager object of the service - this object is manages the creation and registration
	// of the controllers of the server
	CreateManager() (mgr manager.Manager, err error)
}

// ManagerFactory implements IManagerFactory interface
var _ IManagerFactory = (*ManagerFactory)(nil)

// ManagerFactory Factory to create manager.Manager from configuration
type ManagerFactory struct {
	// Configuration is the manager configuration
	configuration *ManagerConfiguration
	// InstrumentationProvider is the instrumentation factory for manager.
	instrumentationProvider instrumentation.IInstrumentationProvider
}

// ManagerConfiguration Factory configuration to create a manager.Manager
type ManagerConfiguration struct {
	// Port is the port that the manager will register the server on.
	Port int
	// CertDir is the directory that the certificates are saved.
	CertDir string
}

// NewManagerFactory Constructor for ManagerFactory
func NewManagerFactory(configuration *ManagerConfiguration, instrumentationProvider instrumentation.IInstrumentationProvider) (factory *ManagerFactory) {
	return &ManagerFactory{
		configuration:           configuration,
		instrumentationProvider: instrumentationProvider}
}

// CreateManager Initialize the manager object of the service - this object is manages the creation and registration
// of the controllers of the server
func (factory *ManagerFactory) CreateManager() (mgr manager.Manager, err error) {
	// GetConfig creates a *rest.Config for talking to a Kubernetes API server (using --kubeconfig or cluster provided config)
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kube-config")
	}

	options, err := factory.createOptions()
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup manager")
	}

	mgr, err = manager.New(cfg, *options)
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup manager")
	}
	// Assign new manager to the server
	return mgr, nil
}

// createOptions Creates manager options
func (factory *ManagerFactory) createOptions() (options *manager.Options, err error) {
	// Add prefix init to cert controller mgr
	tracerProvider := factory.instrumentationProvider.GetTracerProvider("Manager")
	scheme := runtime.NewScheme()
	if err = corev1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to add schema in createOptions")
	}

	options = &manager.Options{
		Scheme:  scheme,
		Logger:  tracerProvider.GetTracer("New"),
		Port:    factory.configuration.Port,
		CertDir: factory.configuration.CertDir,
	}
	return options, nil
}
