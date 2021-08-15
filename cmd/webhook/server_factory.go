// Package webhook is setting up the webhook service, and its own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/go-logr/logr"
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
	// Configuration to provide to server
	configuration *ServerConfiguration
	// Logger is the logger of the server
	logger logr.Logger
	// ManagerFactory is the factory for manager
	managerFactory IManagerFactory
	// CertRotatorFactory is the factory for cert rotator
	certRotatorFactory ICertRotatorFactory
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
	// Initialize logger.
	// TODO will be replaced in the instrumentation PR.
	instrumentation.InitLogger(factory.logger)
	// Create CertRotator using ICertRotatorFactory
	certRotator := factory.certRotatorFactory.CreateCertRotator()
	// Create manager
	mgr, err := factory.managerFactory.CreateManager()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Manager for server")
	}

	// Create Server
	server = NewServer(mgr, factory.logger, certRotator, factory.webhookHandler, factory.configuration)

	return server, nil
}
