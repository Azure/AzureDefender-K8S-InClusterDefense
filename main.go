package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	tivanInstrumentaiton "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/tivan"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/util"
	"log"
)

type mainConfiguration struct {
	isDebug bool
}

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	// Load configuration
	mainConfig := getMainConfiguration()
	tracerConfiguration := getTracerConfiguration()
	metricSubmitterConfiguration := getMetricSubmitterConfiguration()
	instrumentationConfiguration := getInstrumentationConfiguration()
	managerConfiguration := getManagerConfiguration()
	certRotatorConfig := getCertRotatorConfiguration()
	serverConfiguration := getServerConfiguration()

	// Create Tivan's instrumentation
	tivanInstrumentationResult, err := tivanInstrumentaiton.GetTivanInstrumentationResult()
	if err != nil {
		log.Fatal(err)
	}

	// Create factories
	tracerFactory := trace.NewTracerFactory(tracerConfiguration, tivanInstrumentationResult.Tracer)
	if mainConfig.isDebug { // Use zapr logger when debugging
		tracerFactory = trace.NewZaprTracerFactory(tracerConfiguration)
	}
	metricSubmitterFactory := metric.NewMetricSubmitterFactory(metricSubmitterConfiguration, &tivanInstrumentationResult.MetricSubmitter)
	instrumentationFactory := instrumentation.NewInstrumentationFactory(instrumentationConfiguration, tracerFactory, metricSubmitterFactory)
	managerFactory := webhook.NewManagerFactory(managerConfiguration, nil)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfig)
	serverFactory := webhook.NewServerFactory(serverConfiguration, managerFactory, certRotatorFactory, instrumentationFactory)

	// Create Server
	server, err := serverFactory.CreateServer()
	if err != nil {
		log.Fatal(err)
	}
	// Run server
	if err = server.Run(); err != nil {
		log.Fatal(err)
	}
}

//TODO All three methods below will be deleted once Or finishes the configuration
func getInstrumentationConfiguration() *instrumentation.InstrumentationConfiguration {
	return &instrumentation.InstrumentationConfiguration{}
}

func getTracerConfiguration() *trace.TracerConfiguration {
	return &trace.TracerConfiguration{
		TracerLevel:    0,
		DefaultContext: "AzD",
	}
}

func getMetricSubmitterConfiguration() *metric.MetricSubmitterConfiguration {
	return &metric.MetricSubmitterConfiguration{}
}

func getServerConfiguration() (configuration *webhook.ServerConfiguration) {
	return &webhook.ServerConfiguration{
		Path:               "/mutate",
		RunOnDryRunMode:    false,
		EnableCertRotation: true,
	}
}

func getCertRotatorConfiguration() (configuration *webhook.CertRotatorConfiguration) {
	return &webhook.CertRotatorConfiguration{
		Namespace:      util.GetNamespace(),
		SecretName:     "azure-defender-proxy-cert",                           // matches the Secret name
		ServiceName:    "azure-defender-proxy-service",                        // matches the Service name
		WebhookName:    "azure-defender-proxy-mutating-webhook-configuration", // matches the MutatingWebhookConfiguration name
		CaName:         "azure-defender-proxy-ca",
		CaOrganization: "azure-defender-proxy",
		CertDir:        "/certs",
	}
}

func getManagerConfiguration() (configuration *webhook.ManagerConfiguration) {
	return &webhook.ManagerConfiguration{
		Port:    8000,
		CertDir: "/certs",
	}
}

func getMainConfiguration() (configuration *mainConfiguration) {
	return &mainConfiguration{
		isDebug: true,
	}
}
