package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	argqueries "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	azureauthwrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registrywrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
	argbase "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"log"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/tivan"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/util"
)

type mainConfiguration struct {
	isDebug bool
}

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	// Load configuration
	mainConfig := getMainConfiguration()
	tivanInstrumentationConfiguration := getTivanInstrumentationConfiguration()
	tracerConfiguration := getTracerConfiguration()
	metricSubmitterConfiguration := getMetricSubmitterConfiguration()
	instrumentationConfiguration := getInstrumentationConfiguration()
	managerConfiguration := getManagerConfiguration()
	certRotatorConfig := getCertRotatorConfiguration()
	handlerConfiguration := getHandlerConfiguration()
	serverConfiguration := getServerConfiguration()

	// Create Tivan's instrumentation
	tivanInstrumentationResult, err := tivan.NewTivanInstrumentationResult(tivanInstrumentationConfiguration)
	if err != nil {
		log.Fatal("main.NewTivanInstrumentationResult", err)
	}

	// Create factories
	tracerFactory := tivan.NewTracerFactory(tracerConfiguration, tivanInstrumentationResult.Tracer)
	if mainConfig.isDebug { // Use zapr logger when debugging
		tracerFactory = trace.NewZaprTracerFactory(tracerConfiguration)
	}
	metricSubmitterFactory := tivan.NewMetricSubmitterFactory(metricSubmitterConfiguration, &tivanInstrumentationResult.MetricSubmitter)
	instrumentationProviderFactory := instrumentation.NewInstrumentationProviderFactory(instrumentationConfiguration, tracerFactory, metricSubmitterFactory)
	instrumentationProvider, err := instrumentationProviderFactory.CreateInstrumentationProvider()
	if err != nil {
		log.Fatal("main.instrumentationProviderFactory.CreateInstrumentationProvider", err)
	}

	authrozierFactory := azureauth.NewEnvAzureAuthorizerFactory(getEnvAzureAuthorizerConfiguration(), new(azureauthwrappers.AzureAuthWrapper))
	authorizer, err :=  authrozierFactory.CreateARMAuthorizer()
	if err != nil{
		log.Fatal("main.NewEnvAzureAuthorizerFactory.CreateARMAuthorizer", err)
	}

	// Registry Client
	regitryClient := registry.NewRegistryClient(instrumentationProvider, new(registrywrappers.CraneWrapper))

	// ARG
	argBaseClient := argbase.New()
	argBaseClient.Authorizer = authorizer
	argClient := arg.NewARGClient(instrumentationProvider, argBaseClient)
	argQueryGenerator, err := argqueries.CreateARGQueryGenerator()
	if err != nil{
		log.Fatal("main.CreateARGQueryGenerator", err)
	}

	argDataProvider := arg.NewARGDataProvider(instrumentationProvider, argClient, argQueryGenerator)

	// Handler and azsecinfo
	azdSecInfoProvider := azdsecinfo.NewAzdSecInfoProvider(instrumentationProvider, argDataProvider, regitryClient)
	handler := webhook.NewHandler(azdSecInfoProvider, handlerConfiguration, instrumentationProvider)

	// Manager and server
	managerFactory := webhook.NewManagerFactory(managerConfiguration, instrumentationProvider)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfig)
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

func getTivanInstrumentationConfiguration() *tivan.TivanInstrumentationConfiguration {
	return &tivan.TivanInstrumentationConfiguration{
		ComponentName: "AzDProxy",
		MdmNamespace:  "Tivan.Collector.Pods",
	}

}

//TODO All three methods below will be deleted once Or finishes the configuration
func getInstrumentationConfiguration() *instrumentation.InstrumentationProviderConfiguration {
	return &instrumentation.InstrumentationProviderConfiguration{}
}

func getTracerConfiguration() *trace.TracerConfiguration {
	return &trace.TracerConfiguration{
		TracerLevel: 0,
	}
}

func getMetricSubmitterConfiguration() *tivan.MetricSubmitterConfiguration {
	return &tivan.MetricSubmitterConfiguration{}
}

func getServerConfiguration() *webhook.ServerConfiguration {
	return &webhook.ServerConfiguration{
		Path:               "/mutate",
		EnableCertRotation: true,
	}
}

func getCertRotatorConfiguration() *webhook.CertRotatorConfiguration {
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

func getManagerConfiguration() *webhook.ManagerConfiguration {
	return &webhook.ManagerConfiguration{
		Port:    8000,
		CertDir: "/certs",
	}
}

func getMainConfiguration() (configuration *mainConfiguration) {
	return &mainConfiguration{
		isDebug: false,
	}
}

func getHandlerConfiguration() (configuration *webhook.HandlerConfiguration) {
	return &webhook.HandlerConfiguration{
		DryRun: false,
	}
}

func getEnvAzureAuthorizerConfiguration() *azureauth.EnvAzureAuthorizerConfiguration{
	return &azureauth.EnvAzureAuthorizerConfiguration{
		IsLocalDevelopmentMode: true,
		//TODO add MSI

	}
}
