// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package server

import (
	"flag"
	"fmt"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/util"
	"github.com/go-logr/logr"
	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Webhook configuration constants //TODO Change it to config-map
const (
	_secretName  = "azure-defender-proxy-cert"                           // matches the Secret name
	_serviceName = "azure-defender-proxy-service"                        // matches the Service name
	_webhookName = "azure-defender-proxy-mutating-webhook-configuration" // matches the MutatingWebhookConfiguration name
	_webhookPath = "/mutate"                                             // matches the MutatingWebhookConfiguration clientConfig path
	_defaultPort = 8000
)

//Cert controller constants //TODO Change it to config-map
const (
	_defaultCertDir = "/certs"
	_caName         = "azure-defender-proxy-ca"
	_caOrganization = "azure-defender-proxy"
)

// Params of program that can be configured. //TODO Change it to config-map
var (
	port                = flag.Int("port", _defaultPort, "port for the server. defaulted to 8000 if unspecified ")
	certDir             = flag.String("cert-dir", _defaultCertDir, "The directory where certs are stored, defaults to /certs")
	disableCertRotation = flag.Bool("disable-cert-rotation", false, "disable automatic generation and rotation of webhook TLS certificates/keys")
	dryRun              = flag.Bool("dry-run", false, "if true, do not mutate any resources")
)

// Server this struct is responsible for setting up azdproxy server in the cluster.
type Server struct {
	Logger  logr.Logger
	Manager manager.Manager
}

// NewServer Constructor for server
func NewServer() (server *Server) {
	return &Server{}
}

// Run Starting server - this is function is called from the main (entrypoint of azdproxy)
// It initializes the server with all the instrumentation, initialize the controllers, and register them.
// There are 2 controllers - cert-controller (https://github.com/open-policy-agent/cert-controller) that manages
// the certificates of the server and the mutation webhook server that is registered with the AzDSecInfo Handler.
func (server *Server) Run() (err error) {
	server.Logger = ctrl.Log.WithName("webhook-setup")
	instrumentation.InitLogger(server.Logger)

	// Log all parameters:
	server.Logger.Info("Parameters:",
		"port", *port,
		"certDir", certDir,
		"disableCertRotation", disableCertRotation,
		"dryRun", dryRun)

	// Init the manager object of the server - manager manages the creation and registration of the controllers.
	if err := server.initManager(); err != nil {
		return err
	}

	// Init cert controller - gets a channel of setting up the controller.
	certSetupFinished, err := server.initCertController()
	if err != nil {
		server.Logger.Error(err, "Failed to initialize cert controller")
		return err
	}

	// Set up controllers.
	go server.setupControllers(certSetupFinished)

	// Start all registered controllers - webhook mutation as https server and cert controller.
	if err := server.Manager.Start(signals.SetupSignalHandler()); err != nil {
		server.Logger.Error(err, "problem running manager")
		return err
	}
	return nil
}

// initManager Initialize the manager object of the service - this object is manages the creation and registration
// of the controllers of the server
func (server *Server) initManager() (err error) {
	scheme := runtime.NewScheme()
	if err = corev1.AddToScheme(scheme); err != nil {
		server.Logger.Error(err, "Unable to add schema")
		return err
	}

	// GetConfig creates a *rest.Config for talking to a Kubernetes API server (using --kubeconfig or cluster provided config)
	cfg, err := config.GetConfig()
	if err != nil {
		server.Logger.Error(err, "Unable to get kube-config")
		return err
	}

	// Creates new manager of creating controllers
	newManager, err := manager.New(cfg, manager.Options{
		Scheme:  scheme,
		Logger:  server.Logger,
		Port:    *port,
		CertDir: *certDir,
	})
	if err != nil {
		server.Logger.Error(err, "unable to setup manager")
		return err
	}
	// Assign new manager to the server
	server.Manager = newManager
	return nil
}

//initCertController initialize the cert-controller.
// If disableCertRotation is true, it adds new rotator using cert-controller library.
func (server *Server) initCertController() (certSetupFinished chan struct{}, err error) {
	certSetupFinished = make(chan struct{})
	if !*disableCertRotation {
		server.Logger.Info("setting up cert rotation")
		dnsName := fmt.Sprintf("%s.%s.svc", _serviceName, util.GetNamespace()) // matches the MutatingWebhookConfiguration webhooks name
		// Add rotator - using cert-controller API //TODO Expiration of certificate?
		if err := rotator.AddRotator(server.Manager, &rotator.CertRotator{
			SecretKey: types.NamespacedName{
				Namespace: util.GetNamespace(),
				Name:      _secretName,
			},
			CertDir:        *certDir,
			CAName:         _caName,
			CAOrganization: _caOrganization,
			DNSName:        dnsName,
			IsReady:        certSetupFinished,
			Webhooks:       []rotator.WebhookInfo{{Name: _webhookName, Type: rotator.Mutating}},
		}); err != nil {
			server.Logger.Error(err, "Unable to set up cert rotation")
			return nil, err
		}
	} else {
		server.Logger.Info("Skipping certificate provisioning setup")
		close(certSetupFinished)
	}
	return certSetupFinished, nil
}

// setupControllers is setting up all controllers of the server - cert-controller and webhook.
func (server *Server) setupControllers(certSetupFinished chan struct{}) {
	// Setup cert-controller - wait until the channel is finish.
	server.Logger.Info("waiting for cert rotation setup")
	<-certSetupFinished
	server.Logger.Info("done waiting for cert rotation setup")

	// Register mutation webhook.
	server.registerWebhook()
}

// registerWebhook - assigning Handler to the mutation webhook and register it.
func (server *Server) registerWebhook() {
	// Assign webhook handler
	webhookHandler := &Handler{
		Logger: server.Logger.WithName("webhook-handler"),
		DryRun: *dryRun,
	}

	//Register webhook
	mutationWebhook := &admission.Webhook{Handler: webhookHandler}
	server.Manager.GetWebhookServer().Register(_webhookPath, mutationWebhook)
	server.Logger.Info("Webhook registered successfully", "path", _webhookPath, "port", port)
}
