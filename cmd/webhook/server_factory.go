// Package webhook is setting up the webhook service, and its own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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
	// instrumentationProviderFactory
	instrumentationProviderFactory instrumentation.IInstrumentationProviderFactory
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
func NewServerFactory(
	configuration *ServerConfiguration,
	managerFactory IManagerFactory,
	certRotatorFactory ICertRotatorFactory,
	instrumentationFactory instrumentation.IInstrumentationProviderFactory,
	handlerFactory IHandlerFactory) (factory IServerFactory) {
	return &ServerFactory{
		configuration:                  configuration,
		managerFactory:                 managerFactory,
		certRotatorFactory:             certRotatorFactory,
		instrumentationProviderFactory: instrumentationFactory,
		handlerFactory:                 handlerFactory,
	// Handler to provide to server
	webhookHandler admission.Handler
}

// NewServerFactory constructor for ServerFactory
func NewServerFactory(configuration *ServerConfiguration, managerFactory IManagerFactory, certRotatorFactory ICertRotatorFactory, webhookHandler admission.Handler, logger logr.Logger) (factory IServerFactory) {
	return &ServerFactory{
		configuration:      configuration,
		managerFactory:     managerFactory,
		certRotatorFactory: certRotatorFactory,
		webhookHandler:     webhookHandler,
		logger:             logger,
	}
}

// CreateServer creates new server
func (factory *ServerFactory) CreateServer() (server *Server, err error) {
	//Create manager using IInstrumentationFactory
	instrumentationProvider, err := factory.instrumentationProviderFactory.CreateInstrumentationProvider()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the instrumentation of the server")
	}
	// Create CertRotator using ICertRotatorFactory
	certRotator := factory.certRotatorFactory.CreateCertRotator()

	// Create manager using IManagerFactory
	mgr, err := factory.managerFactory.CreateManager()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Manager for server")
	}

	// Create Server
	server = NewServer(mgr, factory.logger, certRotator, factory.webhookHandler, factory.configuration)

	return server, nil
}
