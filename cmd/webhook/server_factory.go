// Package webhook is setting up the webhook service, and its own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
)

// IServerFactory factory to create server
type IServerFactory interface {
	// CreateServer creates new server
	CreateServer() (server *Server, err error)
}

// ServerFactory Factory to create a Server using configuration and manager.
type ServerFactory struct {
	// configuration is the server configuration
	configuration *ServerConfiguration
	// instrumentationFactory
	instrumentationFactory instrumentation.IInstrumentationFactory
	// ManagerFactory is the factory for manager
	managerFactory IManagerFactory
	// CertRotatorFactory is the factory for cert rotator
	certRotatorFactory ICertRotatorFactory
}

// ServerConfiguration Factory configuration to create a server.
type ServerConfiguration struct {
	// Path matches the MutatingWebhookConfiguration clientConfig path
	Path string
	// RunOnDryRunMode is boolean that define if the server should be on dry-run mode
	RunOnDryRunMode bool
	// EnableCertRotation is flag that indicates whether cert rotator should run
	EnableCertRotation bool
}

// NewServerFactory constructor for ServerFactory
func NewServerFactory(configuration *ServerConfiguration, managerFactory IManagerFactory,
	certRotatorFactory ICertRotatorFactory, instrumentationFactory instrumentation.IInstrumentationFactory) (factory IServerFactory) {
	return &ServerFactory{
		configuration:          configuration,
		managerFactory:         managerFactory,
		certRotatorFactory:     certRotatorFactory,
		instrumentationFactory: instrumentationFactory,
	}
}

// CreateServer creates new server
func (factory *ServerFactory) CreateServer() (server *Server, err error) {
	//Create manager using IInstrumentationFactory
	serverInstrumentation, err := factory.instrumentationFactory.CreateInstrumentation()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the instrumentation of the server")
	}

	// Create CertRotator using ICertRotatorFactory
	certRotator := factory.certRotatorFactory.CreateCertRotator()

	// Create manager using IManagerFactory
	mgr, err := factory.managerFactory.CreateManager()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the manager of the server")
	}

	// Create Server
	server = &Server{
		Instrumentation:   serverInstrumentation,
		Manager:           mgr,
		certRotator:       certRotator,
		enableCertRotator: factory.configuration.EnableCertRotation,
		path:              factory.configuration.Path,
		runOnDryMode:      factory.configuration.RunOnDryRunMode,
	}
	return server, nil
}
