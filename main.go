package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	argqueries "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	azureauthwrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registryauthazure "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth/azure"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth/cranekeychain"
	registrywrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
	argbase "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"k8s.io/client-go/kubernetes"
	"log"
	k8sclientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/config"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/tivan"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"os"
)

const (
	_configFileKey = "CONFIG_FILE"
)

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	configFile := os.Getenv(_configFileKey)
	if len(configFile) == 0 {
		log.Fatalf("%v env variable is not defined.", _configFileKey)
	}
	// Load configuration
	AppConfig, err := config.LoadConfig(configFile)
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
	envAzureAuthorizerConfiguration := new(azureauth.EnvAzureAuthorizerConfiguration)

	// Create a map between configuration object and key in main config file
	keyConfigMap := map[string]interface{}{
		"webhook.ManagerConfiguration":                            managerConfiguration,
		"webhook.CertRotatorConfiguration":                        certRotatorConfiguration,
		"webhook.ServerConfiguration":                             serverConfiguration,
		"webhook.HandlerConfiguration":                            handlerConfiguration,
		"instrumentation.tivan.TivanInstrumentationConfiguration": tivanInstrumentationConfiguration,
		"instrumentation.trace.TracerConfiguration":               tracerConfiguration,
		"azureauth.envAzureAuthorizerConfiguration":               envAzureAuthorizerConfiguration}

	for key, configObject := range keyConfigMap {
		// Unmarshal the relevant parts of appConfig's data to each of the configuration objects
		err = config.CreateSubConfiguration(AppConfig, key, configObject)
		if err != nil {
			log.Fatal("failed to load specific configuration data.", err)
		}
	}

	// Create Tivan's instrumentation
	tivanInstrumentationResult, err := tivan.NewTivanInstrumentationResult(tivanInstrumentationConfiguration)
	if err != nil {
		log.Fatal("main.NewTivanInstrumentationResult", err)
	}

	// Create factories
	tracerFactory := tivan.NewTracerFactory(tracerConfiguration, tivanInstrumentationResult.Tracer)
	metricSubmitterFactory := tivan.NewMetricSubmitterFactory(metricSubmitterConfiguration, &tivanInstrumentationResult.MetricSubmitter)
	instrumentationProviderFactory := instrumentation.NewInstrumentationProviderFactory(instrumentationConfiguration, tracerFactory, metricSubmitterFactory)
	instrumentationProvider, err := instrumentationProviderFactory.CreateInstrumentationProvider()
	if err != nil {
		log.Fatal("main.instrumentationProviderFactory.CreateInstrumentationProvider", err)
	}

	authorizerFactory := azureauth.NewEnvAzureAuthorizerFactory(envAzureAuthorizerConfiguration, new(azureauthwrappers.AzureAuthWrapper))
	authorizer, err := authorizerFactory.CreateARMAuthorizer()
	if err != nil {
		log.Fatal("main.NewEnvAzureAuthorizerFactory.CreateARMAuthorizer", err)
	}


	// Registry Client
	k8sclientconfig, err := k8sclientconfig.GetConfig()
	if err != nil {
		log.Fatal("main.k8sclientconfig.GetConfig", err)
	}
	clientK8s, err :=  kubernetes.NewForConfig(k8sclientconfig)
	if err != nil {
		log.Fatal("main.kubernetes.NewForConfig", err)
	}

	// TODO add kublet client Id msi
	bearerAuthorizer, ok := authorizer.(azureauth.IBearerAuthorizer)
	if !ok{
		log.Fatal("main.kubernetes.bearerAuthorizer type assertion", err)

	}
	acrTokenProvider := registryauthazure.NewACRTokenProvider(instrumentationProvider, bearerAuthorizer)

	k8sKeychainFactory := cranekeychain.NewK8SKeychainFactory(instrumentationProvider, clientK8s)
	acrKeychain := cranekeychain.NewACRKeychainFactory(instrumentationProvider, acrTokenProvider)

	multiKeychainFactory := cranekeychain.NewMultiKeychainFactory(instrumentationProvider, k8sKeychainFactory, acrKeychain)
	registryClient := registry.NewCraneRegistryClient(instrumentationProvider, new(registrywrappers.CraneWrapper), multiKeychainFactory)

	// ARG
	argBaseClient := argbase.New()
	argBaseClient.Authorizer = authorizer
	argClient := arg.NewARGClient(instrumentationProvider, argBaseClient)
	argQueryGenerator, err := argqueries.CreateARGQueryGenerator(instrumentationProvider)
	if err != nil {
		log.Fatal("main.CreateARGQueryGenerator", err)
	}

	argDataProvider := arg.NewARGDataProvider(instrumentationProvider, argClient, argQueryGenerator)

	// Handler and azdSecinfoProvider
	azdSecInfoProvider := azdsecinfo.NewAzdSecInfoProvider(instrumentationProvider, argDataProvider, registryClient)
	handler := webhook.NewHandler(azdSecInfoProvider, handlerConfiguration, instrumentationProvider)

	// Manager and server
	managerFactory := webhook.NewManagerFactory(managerConfiguration, instrumentationProvider)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfiguration)
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
