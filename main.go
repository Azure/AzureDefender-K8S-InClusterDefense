package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	config "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/config"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/tivan"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"log"
	"os"
)

const (
	_isLocalDevelopmentKey = "IS_LOCAL_DEVELOPMENT"
)


// IsLocalDevelopment checks if program is running local or on a remote kubernetes cluster
func IsLocalDevelopment() bool {
	_, isFound := os.LookupEnv(_isLocalDevelopmentKey)
	if isFound{
		return false
	}
	return true
}

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	// Load configuration
	AppConfig, err := config.LoadConfig(os.Getenv("CONFIG_NAME"), os.Getenv("CONFIG_TYPE"),
		os.Getenv("CONFIG_PATH"), IsLocalDevelopment())
	if err != nil {
		log.Fatal(err)
	}

	// Create Configuration objects for each factory
	managerConfiguration := new(webhook.ManagerConfiguration)
	certRotatorConfiguration := new(webhook.CertRotatorConfiguration)
	serverConfiguration := new(webhook.ServerConfiguration)
	handlerConfiguration := new(webhook.HandlerConfiguration)
	tivanInstrumentationConfiguration := new(tivan.TivanInstrumentationConfiguration)
	metricSubmitterConfiguration := new(tivan.MetricSubmitterConfiguration)
	tracerConfiguration := new(trace.TracerConfiguration)
	instrumentationConfiguration := new(instrumentation.InstrumentationProviderConfiguration)

	// Unmarshal the relevant parts of appConfig's data to each of the configuration objects
	CreateSubConfiguration(AppConfig, "webhook.ManagerConfiguration", managerConfiguration)
	CreateSubConfiguration(AppConfig, "webhook.CertRotatorConfiguration", certRotatorConfiguration)
	CreateSubConfiguration(AppConfig, "webhook.ServerConfiguration", serverConfiguration)
	CreateSubConfiguration(AppConfig, "webhook.HandlerConfiguration", handlerConfiguration)
	CreateSubConfiguration(AppConfig, "tivan.TivanInstrumentationConfiguration", tivanInstrumentationConfiguration)
	CreateSubConfiguration(AppConfig, "trace.TracerConfiguration", tracerConfiguration)

	// Create Tivan's instrumentation
	tivanInstrumentationResult, err := tivan.NewTivanInstrumentationResult(tivanInstrumentationConfiguration)
	if err != nil {
		log.Fatal("main.NewTivanInstrumentationResult", err)
	}

	// Create factories
	tracerFactory := tivan.NewTracerFactory(tracerConfiguration, tivanInstrumentationResult.Tracer)
	if IsLocalDevelopment() { // Use zapr logger when debugging
		tracerFactory = trace.NewZaprTracerFactory(tracerConfiguration)
	}
	metricSubmitterFactory := tivan.NewMetricSubmitterFactory(metricSubmitterConfiguration, &tivanInstrumentationResult.MetricSubmitter)
	instrumentationProviderFactory := instrumentation.NewInstrumentationProviderFactory(instrumentationConfiguration, tracerFactory, metricSubmitterFactory)
	instrumentationProvider, err := instrumentationProviderFactory.CreateInstrumentationProvider()
	if err != nil {
		log.Fatal("main.instrumentationProviderFactory.CreateInstrumentationProvider", err)
	}

	managerFactory := webhook.NewManagerFactory(managerConfiguration, instrumentationProvider)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfiguration)
	azdSecInfoProvider := azdsecinfo.NewAzdSecInfoProvider()
	handler := webhook.NewHandler(azdSecInfoProvider, handlerConfiguration, instrumentationProvider)
	serverFactory := webhook.NewServerFactory(serverConfiguration, managerFactory, certRotatorFactory, handler, instrumentationProvider)

	// Create Server
	server, err := serverFactory.CreateServer()
	if err != nil {
		log.Fatal("main.serverFactory.CreateServer", err)
	}
	// Run server
	if err = server.Run(); err != nil {
		log.Fatal("main.server.Run", err)
	}
}

// CreateSubConfiguration Create new configuration object for each resource,
// based on it's values in the main configuration file
func CreateSubConfiguration(mainConfiguration *config.ConfigurationProvider, subConfigHierarchy string, configuration interface{}){
	ConfigValues := mainConfiguration.SubConfig(subConfigHierarchy)
	err := ConfigValues.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode the %v into struct, %v", subConfigHierarchy, err)
	}
}

