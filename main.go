package main

import (
	"context"
	"fmt"
	"github.com/Azure/ASC-go-libs/pkg/config"
	tivanInstrumentation "github.com/Azure/ASC-go-libs/pkg/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	argqueries "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	azureauthwrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachewrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/tivan"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/acrauth"
	registryauthazure "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/acrauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/crane"
	registrywrappers "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"os"
	k8sclientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	_cacheContext = context.Background()
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
	tivanInstrumentationConfiguration := new(tivanInstrumentation.InstrumentationConfiguration)
	metricSubmitterConfiguration := new(tivan.MetricSubmitterConfiguration)
	tracerConfiguration := new(trace.TracerConfiguration)
	instrumentationConfiguration := new(instrumentation.InstrumentationProviderConfiguration)
	azdIdentityEnvAzureAuthorizerConfiguration := new(azureauth.MSIAzureAuthorizerConfiguration)
	kubeletIdentityEnvAzureAuthorizerConfiguration := new(azureauth.MSIAzureAuthorizerConfiguration)
	argClientConfiguration := new(arg.ARGClientConfiguration)
	deploymentConfiguration := new(utils.DeploymentConfiguration)
	craneWrapperRetryPolicyConfiguration := new(retrypolicy.RetryPolicyConfiguration)
	argBaseClientRetryPolicyConfiguration := new(retrypolicy.RetryPolicyConfiguration)
	redisCacheClientRetryPolicyConfiguration := new(retrypolicy.RetryPolicyConfiguration)
	acrTokenExchangerClientRetryPolicyConfiguration := new(retrypolicy.RetryPolicyConfiguration)
	argDataProviderConfiguration := new(arg.ARGDataProviderConfiguration)
	tag2DigestResolverConfiguration := new(tag2digest.Tag2DigestResolverConfiguration)
	acrTokenProviderConfiguration := new(acrauth.ACRTokenProviderConfiguration)
	argDataProviderCacheConfiguration := new(cachewrappers.RedisCacheClientConfiguration)
	tokensCacheConfiguration := new(cachewrappers.FreeCacheInMemWrapperCacheConfiguration)
	azdSecInfoProviderConfiguration := new(azdsecinfo.AzdSecInfoProviderConfiguration)
	getContainersVulnerabilityScanInfoTimeoutDuration := new(utils.TimeoutConfiguration)

	// Create a map between configuration object and key in main config file
	keyConfigMap := map[string]interface{}{
		"webhook.managerConfiguration":                            managerConfiguration,
		"webhook.certRotatorConfiguration":                        certRotatorConfiguration,
		"webhook.serverConfiguration":                             serverConfiguration,
		"webhook.handlerConfiguration":                            handlerConfiguration,
		"instrumentation.tivan.tivanInstrumentationConfiguration": tivanInstrumentationConfiguration,
		"instrumentation.trace.tracerConfiguration":               tracerConfiguration,
		"azdIdentity.envAzureAuthorizerConfiguration":             azdIdentityEnvAzureAuthorizerConfiguration,
		"kubeletIdentity.envAzureAuthorizerConfiguration":         kubeletIdentityEnvAzureAuthorizerConfiguration,
		"arg.argBaseClient.retryPolicyConfiguration":              argBaseClientRetryPolicyConfiguration,
		"acr.craneWrappersConfiguration.retryPolicyConfiguration": craneWrapperRetryPolicyConfiguration,
		"acr.tokenExchanger.retryPolicyConfiguration":             acrTokenExchangerClientRetryPolicyConfiguration,
		"acr.acrTokenProviderConfiguration":                       acrTokenProviderConfiguration,
		"arg.argClientConfiguration":                              argClientConfiguration,
		"arg.argDataProviderConfiguration":                        argDataProviderConfiguration,
		"tag2digest.tag2DigestResolverConfiguration":              tag2DigestResolverConfiguration,
		"deployment": deploymentConfiguration,
		"cache.argDataProviderCacheConfiguration":                              argDataProviderCacheConfiguration,
		"cache.tokensCacheConfiguration":                                       tokensCacheConfiguration,
		"cache.redisClient.retryPolicyConfiguration":                           redisCacheClientRetryPolicyConfiguration,
		"azdSecInfoProvider.getContainersVulnerabilityScanInfoTimeoutDuration": getContainersVulnerabilityScanInfoTimeoutDuration,
		"azdSecInfoProvider.azdSecInfoProviderConfiguration":                   azdSecInfoProviderConfiguration,
	}

	for key, configObject := range keyConfigMap {
		// Unmarshal the relevant parts of appConfig's data to each of the configuration objects
		err = config.CreateSubConfiguration(AppConfig, key, configObject)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to load specifc configuration data. \nkey: <%s>\nobjectType: <%T>", key, configObject)
			log.Fatal(errMsg /*Once cache PR is merged, change this to GetType()*/)
		}
	}

	// Validate all TTL values for cache clients configurations are non-positive. Non-positive values are not allowed in order to make sure each value in cache has valid TTL.
	isValidConfiguration, configurationName := utils.ValidatePositiveInt(
		&utils.PositiveIntValidationObject{VariableName: "azdSecInfoProviderConfiguration.CacheExpirationContainerVulnerabilityScanInfo", Variable: azdSecInfoProviderConfiguration.CacheExpirationContainerVulnerabilityScanInfo},
		&utils.PositiveIntValidationObject{VariableName: "azdSecInfoProviderConfiguration.CacheExpirationTimeTimeout", Variable: azdSecInfoProviderConfiguration.CacheExpirationTimeTimeout},
		&utils.PositiveIntValidationObject{VariableName: "argDataProviderConfiguration.CacheExpirationTimeScannedResults", Variable: argDataProviderConfiguration.CacheExpirationTimeScannedResults},
		&utils.PositiveIntValidationObject{VariableName: "argDataProviderConfiguration.CacheExpirationTimeUnscannedResults", Variable: argDataProviderConfiguration.CacheExpirationTimeUnscannedResults},
		&utils.PositiveIntValidationObject{VariableName: "tag2DigestResolverConfiguration.CacheExpirationTimeForResults", Variable: tag2DigestResolverConfiguration.CacheExpirationTimeForResults},
		&utils.PositiveIntValidationObject{VariableName: "acrTokenProviderConfiguration.RegistryRefreshTokenCacheExpirationTime", Variable: acrTokenProviderConfiguration.RegistryRefreshTokenCacheExpirationTime},
	)
	if !isValidConfiguration {
		errMsg := fmt.Sprintf("Got non-positive cache TTL. Only positive values are allowed. Configuration name: <%s>", configurationName)
		log.Fatal(errMsg, utils.InvalidConfiguration)
	}
	// Create deployment singleton.
	deploymentInstance, err := utils.NewDeployment(deploymentConfiguration)
	if err != nil {
		log.Fatal("main.NewDeployment", err)
	}
	// Create Tivan's instrumentation
	// TODO we need a way get the pod name (probably using kubectl).
	tivanInstrumentationResult, err := tivan.NewTivanInstrumentationResult(tivanInstrumentationConfiguration)
	if err != nil {
		log.Fatal("main.NewTivanInstrumentationResult", err)
	}

	// Create factories
	tracerFactory := tivan.NewTracerFactory(tracerConfiguration, tivanInstrumentationResult.Tracer)
	metricSubmitterFactory := tivan.NewMetricSubmitterFactory(metricSubmitterConfiguration, tivanInstrumentationResult.MetricSubmitter)
	instrumentationProviderFactory := instrumentation.NewInstrumentationProviderFactory(instrumentationConfiguration, tracerFactory, metricSubmitterFactory)
	instrumentationProvider, err := instrumentationProviderFactory.CreateInstrumentationProvider()
	if err != nil {
		log.Fatal("main.instrumentationProviderFactory.CreateInstrumentationProvider", err)
	}

	kubeletIdentityAuthorizerFactory := azureauth.NewMSIEnvAzureAuthorizerFactory(instrumentationProvider, kubeletIdentityEnvAzureAuthorizerConfiguration, new(azureauthwrappers.AzureAuthWrapper))
	kubeletIdentityAuthorizer, err := kubeletIdentityAuthorizerFactory.CreateARMAuthorizer()
	if err != nil {
		log.Fatal("main.kubeletIdentityAuthorizerFactory.NewMSIEnvAzureAuthorizerFactory.CreateARMAuthorizer", err)
	}

	// Registry Client
	k8sClientConfig, err := k8sclientconfig.GetConfig()
	if err != nil {
		log.Fatal("main.k8sclientconfig.GetConfig", err)
	}
	clientK8s, err := kubernetes.NewForConfig(k8sClientConfig)
	if err != nil {
		log.Fatal("main.kubernetes.NewForConfig", err)
	}

	azureBearerAuthorizer, ok := kubeletIdentityAuthorizer.(azureauth.IBearerAuthorizer)
	if !ok {
		log.Fatal("main.kubeletIdentityAuthorizer.bearerAuthorizer type assertion", err)

	}
	//Cache clients
	//In mem
	freeCacheInMemCache := cachewrappers.NewFreeCacheInMem(tokensCacheConfiguration)
	freeCacheInMemCacheClient := cache.NewFreeCacheInMemCacheClient(instrumentationProvider, freeCacheInMemCache)
	//Redis
	var persistentCacheClient cache.ICacheClient
	// If this is local deployment - use in mem cache instead of redis
	if deploymentInstance.IsLocalDevelopment() {
		persistentCacheClient = freeCacheInMemCacheClient
	} else {
		// create Redis client
		redisCacheBaseClientFactory := cachewrappers.NewWrapperRedisClientFactory(instrumentationProvider)
		redisCacheBaseClient, err := redisCacheBaseClientFactory.Create(argDataProviderCacheConfiguration)
		if err != nil {
			log.Fatal("main.redisCacheBaseClientFactory.Create got invalid certificates or failed to load cert files or password file", err)
		}
		redisCacheRetryPolicy := retrypolicy.NewRetryPolicy(instrumentationProvider, redisCacheClientRetryPolicyConfiguration)
		redisCacheClient := cache.NewRedisCacheClient(instrumentationProvider, redisCacheBaseClient, redisCacheRetryPolicy, _cacheContext)

		// Check connection every argDataProviderCacheConfiguration.HeartbeatFrequency in minutes
		utils.RepeatEveryTick(utils.GetMinutes(argDataProviderCacheConfiguration.HeartbeatFrequency), redisCacheClient.Ping)

		// Export the client
		persistentCacheClient = redisCacheClient
	}

	azureBearerAuthorizerTokenProvider := azureauth.NewBearerAuthorizerTokenProvider(azureBearerAuthorizer)

	acrTokenExchangerClientRetryPolicy := retrypolicy.NewRetryPolicy(instrumentationProvider, acrTokenExchangerClientRetryPolicyConfiguration)
	acrTokenExchanger := registryauthazure.NewACRTokenExchanger(instrumentationProvider, &http.Client{}, acrTokenExchangerClientRetryPolicy)
	acrTokenProvider := registryauthazure.NewACRTokenProvider(instrumentationProvider, acrTokenExchanger, azureBearerAuthorizerTokenProvider, freeCacheInMemCacheClient, acrTokenProviderConfiguration)

	k8sKeychainFactory := crane.NewK8SKeychainFactory(instrumentationProvider, clientK8s)
	acrKeychainFactory := crane.NewACRKeychainFactory(instrumentationProvider, acrTokenProvider)

	craneWrapperRetryPolicy := retrypolicy.NewRetryPolicy(instrumentationProvider, craneWrapperRetryPolicyConfiguration)
	craneWrapper := registrywrappers.NewCraneWrapper(instrumentationProvider, craneWrapperRetryPolicy)
	// Registry Client
	registryClient := crane.NewCraneRegistryClient(instrumentationProvider, craneWrapper, acrKeychainFactory, k8sKeychainFactory)
	tag2digestResolver := tag2digest.NewTag2DigestResolver(instrumentationProvider, registryClient, persistentCacheClient, tag2DigestResolverConfiguration)

	// ARG

	azdIdentityAuthorizerFactory := azureauth.NewMSIEnvAzureAuthorizerFactory(instrumentationProvider, azdIdentityEnvAzureAuthorizerConfiguration, new(azureauthwrappers.AzureAuthWrapper))
	azdIdentityAuthorizer, err := azdIdentityAuthorizerFactory.CreateARMAuthorizer()
	if err != nil {
		log.Fatal("main.azdIdentityAuthorizerFactory.NewMSIEnvAzureAuthorizerFactory.CreateARMAuthorizer", err)
	}
	argBaseClient, err := wrappers.NewArgBaseClientWrapper(argBaseClientRetryPolicyConfiguration, azdIdentityAuthorizer)
	if err != nil {
		log.Fatal("main.NewArgBaseClientWrapper", err)
	}
	argClient := arg.NewARGClient(instrumentationProvider, argBaseClient, argClientConfiguration)
	argQueryGenerator, err := argqueries.CreateARGQueryGenerator(instrumentationProvider)
	if err != nil {
		log.Fatal("main.CreateARGQueryGenerator", err)
	}
	argDataProviderCacheClient := arg.NewARGDataProviderCacheClient(instrumentationProvider, persistentCacheClient, argDataProviderConfiguration)
	argDataProvider := arg.NewARGDataProvider(instrumentationProvider, argClient, argQueryGenerator, argDataProviderCacheClient, argDataProviderConfiguration)

	// Handler and azdSecinfoProvider
	azdSecInfoProviderCacheClient := azdsecinfo.NewAzdSecInfoProviderCacheClient(instrumentationProvider, persistentCacheClient, azdSecInfoProviderConfiguration)
	azdSecInfoProvider := azdsecinfo.NewAzdSecInfoProvider(instrumentationProvider, argDataProvider, tag2digestResolver, getContainersVulnerabilityScanInfoTimeoutDuration, azdSecInfoProviderCacheClient)
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
