// Package webhook is setting up the webhook service, and its own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// IServerFactory factory to create server
type IServerFactory interface {
	// CreateServer creates new server
	CreateServer() (server *Server)
}

// ServerFactory Factory to create a Server using configuration and manager.
type ServerFactory struct {
	Configuration  *ServerConfiguration // Configuration is the server configuration
	ManagerFactory IManagerFactory      // ManagerFactory is the factory for manager
	Logger         logr.Logger          // Logger is the logger of the server
}

// ServerConfiguration Factory configuration to create a server.
type ServerConfiguration struct {
	Path              string                    // Path matches the MutatingWebhookConfiguration clientConfig path
	CertRotatorConfig *CertRotatorConfiguration // CertRotatorConfig is the configuration of the rotator.CertRotator of the server
	RunOnDryRunMode   bool                      // RunOnDryRunMode is boolean that define if the server should be on dry-run mode
}

// NewServerFactory constructor for ServerFactory
func NewServerFactory(configuration *ServerConfiguration, managerFactory IManagerFactory, logger logr.Logger) (factory *ServerFactory) {
	return &ServerFactory{
		Configuration:  configuration,
		ManagerFactory: managerFactory,
		Logger:         logger,
	}
}

// CreateServer creates new server
func (factory *ServerFactory) CreateServer() (server *Server, err error) {
	// Initialize logger.
	// TODO will be replaced in the instrumentation PR.
	instrumentation.InitLogger(factory.Logger)
	// Create CertRotator using ICertRotatorFactory
	certRotatorFactory := NewCertRotatorFactory(factory.Configuration.CertRotatorConfig)
	certRotator := certRotatorFactory.CreateCertRotator()
	// Create manager
	mgr, err := factory.ManagerFactory.CreateManager(certRotator.CertDir)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create server")
	}

	// Create Server
	server = &Server{
		Manager:           mgr,
		Logger:            factory.Logger,
		path:              factory.Configuration.Path,
		runOnDryMode:      factory.Configuration.RunOnDryRunMode,
		certRotator:       certRotator,
		enableCertRotator: factory.Configuration.CertRotatorConfig.EnableCertRotation,
	}
	return server, nil
}
