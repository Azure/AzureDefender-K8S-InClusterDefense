package azureauth

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// IAzureAuthorizerFactory Factory to create an azure authorizer
type IAzureAuthorizerFactory interface {
	// CreateARMAuthorizer Generates a new ARM azure client authorizer
	CreateARMAuthorizer() (autorest.Authorizer, error)
}

// MSIAzureAuthorizerFactory implements IAzureAuthorizerFactory interface
var _ IAzureAuthorizerFactory = (*MSIAzureAuthorizerFactory)(nil)

// MSIAzureAuthorizerFactory Factory to create an azure authorizer using managed identity
// implements azureauth.IAzureAuthorizerFactory
type MSIAzureAuthorizerFactory struct {

	// tracerProvider
	tracerProvider trace.ITracerProvider

	// configuration factory's configuration
	configuration *MSIAzureAuthorizerConfiguration

	// authWrapper wrapper to all auth package related calls
	authWrapper wrappers.IAzureAuthWrapper
}

// MSIAzureAuthorizerConfiguration Factory configuration to create an azure authorizer from environment (env variables or managed identity)
type MSIAzureAuthorizerConfiguration struct {
	// MSI client id
	MSIClientId string
}

// NewMSIEnvAzureAuthorizerFactory Constructor for MSIAzureAuthorizerFactory
func NewMSIEnvAzureAuthorizerFactory(instrumentationProvider instrumentation.IInstrumentationProvider, configuration *MSIAzureAuthorizerConfiguration, authWrapper wrappers.IAzureAuthWrapper) *MSIAzureAuthorizerFactory {
	return &MSIAzureAuthorizerFactory{
		tracerProvider: instrumentationProvider.GetTracerProvider("MSIAzureAuthorizerFactory"),
		configuration:  configuration,
		authWrapper:    authWrapper,
	}
}

// CreateARMAuthorizer Generates a new ARM azure client authorizer using MSI configured
func (factory *MSIAzureAuthorizerFactory) CreateARMAuthorizer() (autorest.Authorizer, error) {
	tracer := factory.tracerProvider.GetTracer("CreateARMAuthorizer")

	// Gets authorizer setting from environment
	settings, err := factory.authWrapper.GetSettingsFromEnvironment()
	if err != nil {
		err = errors.Wrap(err, "error in GetSettingsFromEnvironment")
		tracer.Error(err, "")
		// Error in fetching settings
		return nil, err
	}

	resourceManagerEndpoint := settings.GetEnvironment().ResourceManagerEndpoint

	// Set ARM as the resource to authorize in settings
	settings.GetValues()[auth.Resource] = resourceManagerEndpoint

	tracer.Info("Settings", "Resource", resourceManagerEndpoint)

	// Generate the MSI authorizer by settings provided and factory's configurations
	authorizer, err := factory.createAuthorizer(settings)
	if err != nil {
		err = errors.Wrap(err, "error in createAuthorizer")
		tracer.Error(err, "")
		return nil, err
	}
	return authorizer, err
}

// createAuthorizer Creates a new authorizer from settings provided.
// The function adds the factory configured user assigned MSI client id to settings value map and generates
// an authorizer using setting' (auth.EnvironmentSettings) GetAuthorizer.
// If factory is configured in local development mode (configuration.IsLocalDevelopmentMode == true):
// creates an authorizer from azure cli with no logged-in user as ID.
func (factory *MSIAzureAuthorizerFactory) createAuthorizer(settings wrappers.IEnvironmentSettingsWrapper) (autorest.Authorizer, error) {
	tracer := factory.tracerProvider.GetTracer("createAuthorizer")

	if settings == nil {
		err := errors.Wrap(utils.NilArgumentError, "null argument settings in createAuthorizer")
		tracer.Error(err, "")
		return nil, err
	}

	// Set client id for user managed identity (empty for system manged identity)
	settings.GetValues()[auth.ClientID] = factory.configuration.MSIClientId

	tracer.Info("Settings", "ClientID", factory.configuration.MSIClientId)

	// If Local development
	if utils.GetDeploymentInstance().IsLocalDevelopment() {
		// Local development - az cli auth
		return factory.authWrapper.NewAuthorizerFromCLIWithResource(settings.GetValues()[auth.Resource])
	}
	// Get MSI authorizer
	return settings.GetMSIAuthorizer()
}
