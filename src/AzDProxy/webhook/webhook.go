package webhook

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/zapr"
	"github.com/google/k8s-digester/pkg/handler"

	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	crzap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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
	setupLog         = ctrl.Log.WithName("setup")
	logLevelEncoders = map[string]zapcore.LevelEncoder{
		"lower":        zapcore.LowercaseLevelEncoder,
		"capital":      zapcore.CapitalLevelEncoder,
		"color":        zapcore.LowercaseColorLevelEncoder,
		"capitalcolor": zapcore.CapitalColorLevelEncoder,
	}
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
	initLogger()
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Errorf("unable to get kubeconfig: %w", err)
		return
	}
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:         scheme,
		Logger:         setupLog,
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
	if err := mgr.AddReadyzCheck("default", healthz.Ping); err != nil {
		fmt.Errorf("unable to create readyz check: %w", err)
		return
	}
	if err := mgr.AddHealthzCheck("default", healthz.Ping); err != nil {
		fmt.Errorf("unable to create healthz check: %w", err)
		return
	}
	certSetupFinished := make(chan struct{})
	if !*disableCertRotation {
		setupLog.Info("setting up cert rotation")
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
		setupLog.Info("skipping certificate provisioning setup")
		close(certSetupFinished)
	}

	go setupControllers(mgr, certSetupFinished)

	setupLog.Info("starting manager")
	if err := mgr.Start(context.Background()); err != nil {
		fmt.Errorf("problem running manager: %w", err)
		return
	}
	return
}

func setupControllers(mgr manager.Manager, certSetupFinished chan struct{}) {
	setupLog.Info("waiting for cert rotation setup")
	<-certSetupFinished
	setupLog.Info("done waiting for cert rotation setup")
	var config *rest.Config
	if !*offline {
		config = mgr.GetConfig()
	}

	whh := &handler.Handler{
		Log:    setupLog,
		DryRun: *dryRun, // Delete param from func arguments
		Config: config,
	}
	mwh := &admission.Webhook{Handler: whh}
	setupLog.Info("starting webhook server", "path", webhookPath)
	mgr.GetWebhookServer().Register(webhookPath, mwh)
}

func initLogger() {
	encoder, ok := logLevelEncoders[*logLevelEncoder]
	if !ok {
		setupLog.Error(fmt.Errorf("invalid log level encoder: %v", *logLevelEncoder), "Invalid log level encoder")
		os.Exit(1)
	}

	switch *logLevel {
	case "DEBUG":
		eCfg := zap.NewDevelopmentEncoderConfig()
		eCfg.LevelKey = *logLevelKey
		eCfg.EncodeLevel = encoder
		logger := crzap.New(crzap.UseDevMode(true), crzap.Encoder(zapcore.NewConsoleEncoder(eCfg)))
		ctrl.SetLogger(logger)
		klog.SetLogger(logger)
	case "WARNING", "ERROR":
		setLoggerForProduction(encoder)
	case "INFO":
		fallthrough
	default:
		eCfg := zap.NewProductionEncoderConfig()
		eCfg.LevelKey = *logLevelKey
		eCfg.EncodeLevel = encoder
		logger := crzap.New(crzap.UseDevMode(false), crzap.Encoder(zapcore.NewJSONEncoder(eCfg)))
		ctrl.SetLogger(logger)
		klog.SetLogger(logger)
	}
}

func setLoggerForProduction(encoder zapcore.LevelEncoder) {

	sink := zapcore.AddSync(os.Stderr)
	var opts []zap.Option
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.LevelKey = *logLevelKey
	encCfg.EncodeLevel = encoder
	enc := zapcore.NewJSONEncoder(encCfg)
	lvl := zap.NewAtomicLevelAt(zap.WarnLevel)
	opts = append(opts, zap.AddStacktrace(zap.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}),
		zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	zlog := zap.New(zapcore.NewCore(&crzap.KubeAwareEncoder{Encoder: enc, Verbose: false}, sink, lvl))
	zlog = zlog.WithOptions(opts...)
	newlogger := zapr.NewLogger(zlog)
	ctrl.SetLogger(newlogger)
	klog.SetLogger(newlogger)
}
