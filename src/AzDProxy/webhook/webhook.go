package webhook

import (
	"flag"
	"fmt"

	"github.com/open-policy-agent/cert-controller/pkg/rotator"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	dnsName  = fmt.Sprintf("%s.%s.svc", serviceName, "default")
	webhooks = []rotator.WebhookInfo{
		{
			Name: webhookName,
			Type: rotator.Mutating,
		},
	}
)

const (
	// defaultCertDir     = "/certs"
	// defaultMetricsAddr = ":8888"
	// defaultHealthAddr  = ":9090"
	// defaultPort        = 443

	secretName     = "azure-defender-cert"          // matches the Secret name
	serviceName    = "azure-defender-proxy-service" // matches the Service name
	caName         = "azure-defender-proxy-ca"
	caOrganization = "azure-defender-proxy"
	webhookName    = "azure-defender-proxy-mutating-webhook-configuration" // matches the MutatingWebhookConfiguration name
	webhookPath    = "/mutate"                                             // matches the MutatingWebhookConfiguration clientConfig path
)

var (
	logLevel            = flag.String("log-level", "INFO", "Minimum log level. For example, DEBUG, INFO, WARNING, ERROR. Defaulted to INFO if unspecified.")
	logLevelKey         = flag.String("log-level-key", "level", "JSON key for the log level field, defaults to `level`")
	logLevelEncoder     = flag.String("log-level-encoder", "lower", "Encoder for the value of the log level field. Valid values: [`lower`, `capital`, `color`, `capitalcolor`], default: `lower`")
	port                = flag.Int("port", 443, "port for the server. defaulted to 443 if unspecified ")
	certDir             = flag.String("cert-dir", "/certs", "The directory where certs are stored, defaults to /certs")
	disableCertRotation = flag.Bool("disable-cert-rotation", false, "disable automatic generation and rotation of webhook TLS certificates/keys")
	offline             = flag.Bool("offline", false, "do not connect to API server to retrieve imagePullSecrets")
	dryRun              = flag.Bool("dry-run", false, "if true, do not mutate any resources")
)

func StartServer() {
	InitLogger()
	SetupLog.Info("Constants:", "dns_name", dnsName, "port", *port, "certDir", certDir, "disableCertRotation", disableCertRotation, "offline", offline, "dryRun", dryRun, "webhooks", webhooks)
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Errorf("unable to get kubeconfig: %w", err)
		return
	}
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:         scheme,
		Logger:         SetupLog,
		LeaderElection: false,
		Port:           *port,
		CertDir:        *certDir,
		// MetricsBindAddress:     *metricsAddr,
		// HealthProbeBindAddress: *healthAddr,
	})
	if err != nil {
		fmt.Errorf("unable to set up manager: %w", err)
		return
	}
	// if err := mgr.AddReadyzCheck("default", healthz.Ping); err != nil {
	// 	fmt.Errorf("unable to create readyz check: %w", err)
	// 	return
	// }
	// if err := mgr.AddHealthzCheck("default", healthz.Ping); err != nil {
	// 	fmt.Errorf("unable to create healthz check: %w", err)
	// 	return
	// }
	certSetupFinished := make(chan struct{})
	if !*disableCertRotation {
		SetupLog.Info("setting up cert rotation")
		if err := rotator.AddRotator(mgr, &rotator.CertRotator{
			SecretKey: types.NamespacedName{
				Namespace: "default", //util.GetNamespace(),
				Name:      secretName,
			},
			CertDir:        *certDir,
			CAName:         caName,
			CAOrganization: caOrganization,
			DNSName:        dnsName,
			IsReady:        certSetupFinished,
			Webhooks:       webhooks,
		}); err != nil {
			fmt.Errorf("unable to set up cert rotation: %w", err)
			return
		}
	} else {
		SetupLog.Info("skipping certificate provisioning setup")
		close(certSetupFinished)
	}

	go setupControllers(mgr, certSetupFinished)

	SetupLog.Info("starting manager and backgorund")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		fmt.Errorf("problem running manager: %w", err)
		return
	}
	SetupLog.Info("", "Listening on port:", *port)
}

func setupControllers(mgr manager.Manager, certSetupFinished chan struct{}) {
	SetupLog.Info("waiting for cert rotation setup")
	<-certSetupFinished
	SetupLog.Info("done waiting for cert rotation setup")
	var config *rest.Config
	if !*offline {
		config = mgr.GetConfig()
	}

	whh := &Handler{
		Log:    SetupLog,
		DryRun: *dryRun, // Delete param from func arguments
		Config: config,
	}
	mwh := &admission.Webhook{Handler: whh}
	mgr.GetWebhookServer().Register(webhookPath, mwh)
	SetupLog.Info("Webhook registered successfully", "path", webhookPath)

}
