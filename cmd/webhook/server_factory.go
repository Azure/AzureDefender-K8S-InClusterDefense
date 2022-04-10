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

// ServerFactory implements IServerFactory interface
var _ IServerFactory = (*ServerFactory)(nil)

// ServerFactory Factory to create a Server using configuration and manager.
type ServerFactory struct {
	// configuration is the server configuration
	configuration *ServerConfiguration
	// instrumentationProvider
	instrumentationProvider instrumentation.IInstrumentationProvider
	// ManagerFactory is the factory for manager
	managerFactory IManagerFactory
	// CertRotatorFactory is the factory for cert rotator
	certRotatorFactory ICertRotatorFactory
	// webhookHandler
	webhookHandler admission.Handler
}

// NewServerFactory constructor for ServerFactory
func NewServerFactory(configuration *ServerConfiguration,
	managerFactory IManagerFactory,
	certRotatorFactory ICertRotatorFactory,
	webhookHandler admission.Handler,
	instrumentationProvider instrumentation.IInstrumentationProvider) (factory IServerFactory) {
	return &ServerFactory{
		configuration:           configuration,
		managerFactory:          managerFactory,
		certRotatorFactory:      certRotatorFactory,
		webhookHandler:          webhookHandler,
		instrumentationProvider: instrumentationProvider,
	}
}

// CreateServer creates new server
func (factory *ServerFactory) CreateServer() (server *Server, err error) {

	// Create CertRotator using ICertRotatorFactory
	certRotator := factory.certRotatorFactory.CreateCertRotator()

	// Create manager using IManagerFactory
	mgr, err := factory.managerFactory.CreateManager()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Manager for server")
	}
	// Create Server
	server = NewServer(factory.instrumentationProvider, mgr, certRotator, factory.webhookHandler, factory.configuration)

	return server, nil
}
