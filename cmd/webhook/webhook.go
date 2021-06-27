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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Webhook configuration constants
const (
	secretName  = "azure-defender-cert"                                 // matches the Secret name
	serviceName = "azure-defender-proxy-service"                        // matches the Service name
	webhookName = "azure-defender-proxy-mutating-webhook-configuration" // matches the MutatingWebhookConfiguration name
	webhookPath = "/mutate"                                             // matches the MutatingWebhookConfiguration clientConfig path
	defaultPort = 8000
)

//Cert controller constants
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

// Params of program that can be configured.
var (
	port                = flag.Int("port", defaultPort, "port for the server. defaulted to 8000 if unspecified ")
	certDir             = flag.String("cert-dir", defaultCertDir, "The directory where certs are stored, defaults to /certs")
	disableCertRotation = flag.Bool("disable-cert-rotation", false, "disable automatic generation and rotation of webhook TLS certificates/keys")
	dryRun              = flag.Bool("dry-run", false, "if true, do not mutate any resources")
)

// Logger
var (
	logger = ctrl.Log.WithName("webhook-setup")
)

// StartServer Starting server.
func StartServer() {
	instrumentation.InitLogger(logger)
	logger.Info("Parameters:",
		"port", *port,
		"certDir", certDir,
		"disableCertRotation", disableCertRotation,
		"dryRun", dryRun,
		"webhooks", webhooks)

	mgr, isManagerInitializedFailed := initManager()
	if isManagerInitializedFailed {
		return
	}

	certSetupFinished := make(chan struct{})
	if !*disableCertRotation {
		// dnsName matches the MutatingWebhookConfiguration webhooks name
		dnsName := fmt.Sprintf("%s.%s.svc", serviceName, util.GetNamespace())
		logger.Info("setting up cert rotation")
		if err := rotator.AddRotator(mgr, &rotator.CertRotator{
			SecretKey: types.NamespacedName{
				Namespace: util.GetNamespace(), //util.GetNamespace(),
				Name:      secretName,
			},
			CertDir:        *certDir,
			CAName:         caName,
			CAOrganization: caOrganization,
			DNSName:        dnsName,
			IsReady:        certSetupFinished,
			Webhooks:       webhooks,
		}); err != nil {
			logger.Error(err, "unable to set up cert rotation")
			return
		}
	} else {
		logger.Info("skipping certificate provisioning setup")
		close(certSetupFinished)
	}
	go setupControllers(mgr, certSetupFinished)
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err, "problem running manager")
		return
	}
}

// Initialize manager object.
func initManager() (mgr manager.Manager, isManagerInitializedFailed bool) {
	scheme := runtime.NewScheme()
	err := corev1.AddToScheme(scheme)
	if err != nil {
		logger.Error(err, "Unable with new schema")
		return nil, true
	}
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(err, "Unable to get kubeconfig")
		return nil, true
	}
	mgr, err = manager.New(cfg, manager.Options{
		Scheme:  scheme,
		Logger:  logger,
		Port:    *port,
		CertDir: *certDir,
	})
	if err != nil {
		logger.Error(err, "unable to set up manager")
		return nil, true
	}
	return mgr, false
}

// Setup controllers - cert controller and register webhook
func setupControllers(mgr manager.Manager, certSetupFinished chan struct{}) {
	// Setup cert controller
	logger.Info("waiting for cert rotation setup")
	<-certSetupFinished
	logger.Info("done waiting for cert rotation setup")

	// Assign webhook handler
	whh := &Handler{
		Log:    logger.WithName("webhook-handler"),
		DryRun: *dryRun,
		Config: mgr.GetConfig(),
	}
	//Register webhook
	mwh := &admission.Webhook{Handler: whh}
	mgr.GetWebhookServer().Register(webhookPath, mwh)
	logger.Info("Webhook registered successfully", "path", webhookPath, "port", port)
}
