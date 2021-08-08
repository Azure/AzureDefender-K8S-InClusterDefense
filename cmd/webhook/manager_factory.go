package webhook

import (
	"github.com/go-logr/logr"
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

// ManagerFactory Factory to create manager.Manager from configuration
type ManagerFactory struct {
	Configuration *ManagerConfiguration // Configuration is the manager configuration
	Logger        logr.Logger           // Logger is the manager logger.
}

// ManagerConfiguration Factory configuration to create a manager.Manager
type ManagerConfiguration struct {
	Port    int    // Port is the port that the manager will register the server on.
	CertDir string // CertDir is the directory that the certificates are saved.
}

// NewManagerFactory Constructor for ManagerFactory
func NewManagerFactory(configuration *ManagerConfiguration, logger logr.Logger) (factory IManagerFactory) {
	return &ManagerFactory{
		Configuration: configuration,
		Logger:        logger}
}

// CreateManager Initialize the manager object of the service - this object is manages the creation and registration
// of the controllers of the server
func (factory *ManagerFactory) CreateManager() (mgr manager.Manager, err error) {
	// GetConfig creates a *rest.Config for talking to a Kubernetes API server (using --kubeconfig or cluster provided config)
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kube-config")
	}

	options, err := factory.createOptions(factory.Configuration.CertDir)
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
func (factory *ManagerFactory) createOptions(certDir string) (options *manager.Options, err error) {
	scheme := runtime.NewScheme()
	if err = corev1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to add schema")
	}
	options = &manager.Options{
		Scheme:  scheme,
		Logger:  factory.Logger,
		Port:    factory.Configuration.Port,
		CertDir: certDir,
	}
	return options, nil
}
