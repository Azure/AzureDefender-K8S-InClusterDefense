// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/go-logr/logr"
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
	CertDir           string                    // CertDir the directory where certs are stored
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
	// Create manager
	mgr, err := factory.ManagerFactory.CreateManager()
	if err != nil {
		return nil, fmt.Errorf("unable to create server: " + err.Error())
	}
	// Create CertRotator using ICertRotatorFactory
	certRotatorFactory := NewCertRotatorFactory(factory.Configuration.CertRotatorConfig)
	certRotator := certRotatorFactory.CreateCertRotator(factory.Configuration.CertDir)
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
