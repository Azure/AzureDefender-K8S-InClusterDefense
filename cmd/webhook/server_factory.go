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
	Configuration *ServerConfiguration
	// Logger is the logger of the server
	Logger logr.Logger
	// ManagerFactory is the factory for manager
	ManagerFactory IManagerFactory
	// CertRotatorFactory is the factory for cert rotator
	CertRotatorFactory ICertRotatorFactory
	// Handler to provide to server
	WebhookHandler admission.Handler
}

// NewServerFactory constructor for ServerFactory
func NewServerFactory(configuration *ServerConfiguration, managerFactory IManagerFactory, certRotatorFactory ICertRotatorFactory, webhookHandler admission.Handler, logger logr.Logger) (factory IServerFactory) {
	return &ServerFactory{
		Configuration:      configuration,
		ManagerFactory:     managerFactory,
		CertRotatorFactory: certRotatorFactory,
		WebhookHandler:     webhookHandler,
		Logger:             logger,
	}
}

// CreateServer creates new server
func (factory *ServerFactory) CreateServer() (server *Server, err error) {
	// Initialize logger.
	// TODO will be replaced in the instrumentation PR.
	instrumentation.InitLogger(factory.Logger)
	// Create CertRotator using ICertRotatorFactory
	certRotator := factory.CertRotatorFactory.CreateCertRotator()
	// Create manager
	mgr, err := factory.ManagerFactory.CreateManager()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ,anager for server")
	}

	// Create Server
	server = NewServer(mgr, factory.Logger, certRotator, factory.WebhookHandler, factory.Configuration)

	return server, nil
}
