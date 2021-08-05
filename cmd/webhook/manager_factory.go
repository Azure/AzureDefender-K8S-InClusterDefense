package webhook

import (
	"fmt"
	"github.com/go-logr/logr"
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
	Configuration *ManagerConfiguration
	Logger        logr.Logger
}

// ManagerConfiguration Factory configuration to create a manager.Manager
type ManagerConfiguration struct {
	Port    int
	CertDir string
}

// NewManagerFactory Constrcutor for ManagerFactory
func NewManagerFactory(configuration *ManagerConfiguration, logger logr.Logger) (factory *ManagerFactory) {
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
		return nil, fmt.Errorf("unable to get kube-config: " + err.Error())
	}

	options, err := factory.createOptions()
	if err != nil {
		return nil, fmt.Errorf("unable to setup manager: " + err.Error())
	}

	mgr, err = manager.New(cfg, *options)
	if err != nil {
		return nil, fmt.Errorf("unable to setup manager: " + err.Error())
	}
	// Assign new manager to the server
	return mgr, nil
}

// createOptions Creates manager options
func (factory *ManagerFactory) createOptions() (options *manager.Options, err error) {
	scheme := runtime.NewScheme()
	if err = corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("unable to add schema: " + err.Error())
	}
	options = &manager.Options{
		Scheme:  scheme,
		Logger:  factory.Logger,
		Port:    factory.Configuration.Port,
		CertDir: factory.Configuration.CertDir,
	}
	return options, nil
}
