package webhook

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/go-logr/logr"
	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Server this struct is responsible for setting up azdproxy server in the cluster.
type Server struct {
	// tracerProvider is the tracer provider of the server.
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter of the server
	metricSubmitter metric.IMetricSubmitter
	// manager is the manager.Manager of the server - it is registers the server.
	manager manager.Manager
	//certRotator is the cert rotator which manage the certificates of the server.
	certRotator *rotator.CertRotator
	// webhookHandler Server admission webhook handler
	webhookHandler admission.Handler
	// configuration  - Server configuration
	configuration *ServerConfiguration
}

// ServerConfiguration configuration
type ServerConfiguration struct {
	// Path matches the MutatingWebhookConfiguration clientConfig path
	Path string
	// EnableCertRotation is flag that indicates whether cert rotator should run
	EnableCertRotation bool
}

// NewServer Server constructor
func NewServer(manager manager.Manager, logger logr.Logger, certRotator *rotator.CertRotator, webhookHandler admission.Handler, configuration *ServerConfiguration) *Server {

	// Extract tracer and metricSubmitter from instrumentationProvider
	tracerProvider := instrumentationProvider.GetTracerProvider("Server")
	metricSubmitter := instrumentationProvider.GetMetricSubmitter()

	return &Server{
		manager:        manager,
		logger:         logger,
		certRotator:    certRotator,
		webhookHandler: webhookHandler,
		configuration:  configuration,
	}
}

// Run Starting server - this is function is called from the main (entrypoint of azdproxy)
// It initializes the server with all the instrumentation, initialize the controllers, and register them.
// There are 2 controllers - cert-controller (https://github.com/open-policy-agent/cert-controller) that manages
// the certificates of the server and the mutation webhook server that is registered with the AzDSecInfo Handler.
func (server *Server) Run() (err error) {
	// Init cert controller - gets a channel of setting up the controller.
	if err = server.initCertController(); err != nil {
		return errors.Wrap(err, "failed to initialize cert controller")
	}

	// Set up controllers.
	go server.setupControllers()

	// Start all registered controllers - webhook mutation as https server and cert controller.
	if err := server.manager.Start(signals.SetupSignalHandler()); err != nil {
		return errors.Wrap(err, "unable to start manager")
	}
	return nil
}

//initCertController initialize the cert-controller.
// If disableCertRotation is true, it adds new rotator using cert-controller library.
func (server *Server) initCertController() (err error) {
	tracer := server.tracerProvider.GetTracer("initCertController")

	if server.configuration.enableCertRotator {
		tracer.Info("setting up cert rotation")
		// Add rotator - using cert-controller API //TODO Expiration of certificate?
		if err := rotator.AddRotator(server.manager, server.certRotator); err != nil {
			return errors.Wrap(err, "unable to setup cert rotation")
		}
	} else {
		tracer.Info("Skipping certificate provisioning setup")
		close(server.certRotator.IsReady)
	}
	return nil
}

// setupControllers is setting up all controllers of the server - cert-controller and webhook.
func (server *Server) setupControllers() {
	tracer := server.tracerProvider.GetTracer("setupControllers")

	// Setup cert-controller - wait until the channel is finish.
	tracer.Info("waiting for cert rotation setup")
	<-server.certRotator.IsReady
	tracer.Info("done waiting for cert rotation setup")

	// Register mutation webhook.
	server.registerWebhook()
}

// registerWebhook - assigning Handler to the mutation webhook and register it.
func (server *Server) registerWebhook() {
	tracer := server.tracerProvider.GetTracer("registerWebhook")

	//Register webhook
	mutationWebhook := &admission.Webhook{Handler: server.webhookHandler}
	server.manager.GetWebhookServer().Register(server.configuration.Path, mutationWebhook)
	tracer.Info("Webhook registered successfully", "path", server.configuration.Path)
}
