// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"flag"
	"fmt"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/pkg/infra/util"
	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Webhook configuration constants //TODO Change it to config-map
const (
	secretName  = "azure-defender-proxy-cert"                           // matches the Secret name
	serviceName = "azure-defender-proxy-service"                        // matches the Service name
	webhookName = "azure-defender-proxy-mutating-webhook-configuration" // matches the MutatingWebhookConfiguration name
	webhookPath = "/mutate"                                             // matches the MutatingWebhookConfiguration clientConfig path
	defaultPort = 8000
)

//Cert controller constants //TODO Change it to config-map
const (
	defaultCertDir = "/certs"
	caName         = "azure-defender-proxy-ca"
	caOrganization = "azure-defender-proxy"
)

//Webhooks of AzDProxy.
var (
	webhooks = []rotator.WebhookInfo{
		{
			Name: webhookName,
			Type: rotator.Mutating,
		},
	}
)

// Params of program that can be configured. //TODO Change it to config-map
var (
	port                = flag.Int("port", defaultPort, "port for the server. defaulted to 8000 if unspecified ")
	certDir             = flag.String("cert-dir", defaultCertDir, "The directory where certs are stored, defaults to /certs")
	disableCertRotation = flag.Bool("disable-cert-rotation", false, "disable automatic generation and rotation of webhook TLS certificates/keys")
	dryRun              = flag.Bool("dry-run", false, "if true, do not mutate any resources")
)

// Server this struct is responsible for setting up azdproxy server in the cluster.
type Server struct {
	Instrumentation *instrumentation.Instrumentation
	Manager         manager.Manager
}

// NewServer creates new server
func NewServer(serverInstrumentation *instrumentation.Instrumentation) (server *Server) {
	return &Server{Instrumentation: serverInstrumentation}
}

// Run Starting server - this is function is called from the main (entrypoint of azdproxy)
// It initialize the server with all the instrumentation, initialize the controllers, and register them.
// There are 2 controllers - cert-controller (https://github.com/open-policy-agent/cert-controller) that manages
// the certificates of the server and the mutation webhook server that is registered with the AzDSecInfo Handler.
func (server *Server) Run() {
	flag.Parse()
	// Log all parameters:
	server.Instrumentation.Tracer.Info("Parameters:",
		"port", *port,
		"certDir", *certDir,
		"disableCertRotation", *disableCertRotation,
		"dryRun", *dryRun,
		"webhooks", webhooks)

	// Init the manager object of the server - manager manages the creation and registration of the controllers.
	if err := server.initManager(); err != nil {
		return //TODO exit with panic - error flow?
	}

	// Init cert controller - gets a channel of setting up the controller.
	certSetupFinished, err := server.initCertController()
	if err != nil {
		server.Instrumentation.Tracer.Error(err, "Failed to initialize cert controller")
		return //TODO exit with panic - error flow?
	}

	// Set up controllers.
	go server.setupControllers(certSetupFinished)

	// Start all registered controllers - webhook mutation as https server and cert controller.
	if err := server.Manager.Start(signals.SetupSignalHandler()); err != nil {
		server.Instrumentation.Tracer.Error(err, "problem running manager")
		return //TODO exit with panic - error flow?
	}
}

// initManager Initialize the manager object of the service - this object is manages the creation and registration
// of the controllers of the server
func (server *Server) initManager() (err error) {
	scheme := runtime.NewScheme()
	if err = corev1.AddToScheme(scheme); err != nil {
		server.Instrumentation.Tracer.Error(err, "Unable to add schema")
		return err
	}

	// GetConfig creates a *rest.Config for talking to a Kubernetes API server (using --kubeconfig or cluster provided config)
	cfg, err := config.GetConfig()
	if err != nil {
		server.Instrumentation.Tracer.Error(err, "Unable to get kube-config")
		return err
	}
	// Creates new manager of creating controllers
	newManager, err := manager.New(cfg, manager.Options{
		Scheme:  scheme,
		Logger:  server.Instrumentation.Tracer,
		Port:    *port,
		CertDir: *certDir,
	})
	if err != nil {
		server.Instrumentation.Tracer.Error(err, "unable to setup manager")
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
		dnsName := fmt.Sprintf("%s.%s.svc", serviceName, util.GetNamespace()) // matches the MutatingWebhookConfiguration webhooks name
		server.Instrumentation.Tracer.Info("setting up cert rotation")
		// Add rotator - using cert-controller API
		certRotator := NewCertRotator(&dnsName, certSetupFinished)
		if err := rotator.AddRotator(server.Manager, certRotator); err != nil {
			server.Instrumentation.Tracer.Error(err, "Unable to set up cert rotation")
			return nil, err
		}
	} else {
		server.Instrumentation.Tracer.Info("Skipping certificate provisioning setup")
		close(certSetupFinished)
	}
	return certSetupFinished, nil
}

// NewCertRotator creates new cert rotator
func NewCertRotator(dnsName *string, certSetupFinished chan struct{}) (certRotator *rotator.CertRotator) {
	//TODO Add expiration?
	return &rotator.CertRotator{
		SecretKey: types.NamespacedName{
			Namespace: util.GetNamespace(),
			Name:      secretName,
		},
		CertDir:        *certDir,
		CAName:         caName,
		CAOrganization: caOrganization,
		DNSName:        *dnsName,
		IsReady:        certSetupFinished,
		Webhooks:       webhooks,
	}
}

// setupControllers is setting up all controllers of the server - cert-controller and webhook.
func (server *Server) setupControllers(certSetupFinished chan struct{}) {
	// Setup cert-controller - wait until the channel is finish.
	server.Instrumentation.Tracer.Info("waiting for cert rotation setup")
	<-certSetupFinished
	server.Instrumentation.Tracer.Info("done waiting for cert rotation setup")

	// Register mutation webhook.
	server.registerWebhook()
}

// registerWebhook - assigning Handler to the mutation webhook and register it.
func (server *Server) registerWebhook() {
	// Creates new webhook handler
	webhookHandler := NewHandler(server.Instrumentation, *dryRun)

	//Register webhook
	mutationWebhook := &admission.Webhook{Handler: webhookHandler}
	server.Manager.GetWebhookServer().Register(webhookPath, mutationWebhook)
	server.Instrumentation.Tracer.Info("Webhook registered successfully", "path", webhookPath, "port", port)
}
