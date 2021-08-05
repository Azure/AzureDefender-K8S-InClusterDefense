package azureauth

import (
	"fmt"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// IAzureAuthorizerFactory Factory to create an azure authorizer
type IAzureAuthorizerFactory interface {
	// NewARMAuthorizer Generates a new ARM azure client authorizer
	NewARMAuthorizer() (autorest.Authorizer, error)
}

// EnvAzureAuthorizerFactory Factory to create an azure authorizer using managed identity
// implements azureauth.IAzureAuthorizerFactory
type EnvAzureAuthorizerFactory struct {
	// configuration factory's configuration
	configuration *EnvAzureAuthorizerConfiguration

	// authWrapper wrapper to all auth package related calls
	authWrapper wrappers.IAzureAuthWrapper
}

// EnvAzureAuthorizerConfiguration Factory configuration to create an azure authorizer from environment (env variables or managed identity)
type EnvAzureAuthorizerConfiguration struct {
	// isLocalDevelopmentMode is factory set to local development
	isLocalDevelopmentMode bool

	// MSI client id
	MSIClientId string
}

// NewEnvAzureAuthorizerFactory Constructor for EnvAzureAuthorizerFactory
func NewEnvAzureAuthorizerFactory(configuration *EnvAzureAuthorizerConfiguration, authWrapper wrappers.IAzureAuthWrapper) *EnvAzureAuthorizerFactory {
	return &EnvAzureAuthorizerFactory{
		configuration: configuration,
		authWrapper:   authWrapper,
	}
}

// NewARMAuthorizer Generates a new ARM azure client authorizer using MSI configured
func (factory *EnvAzureAuthorizerFactory) NewARMAuthorizer() (autorest.Authorizer, error) {

	// Gets authorizer setting from environment
	settings, err := factory.authWrapper.GetSettingsFromEnvironment()
	if err != nil {
		// Error in fetching settings
		return nil, err
	}

	// Set ARM as the resource to authorize in settings
	settings.GetValues()[auth.Resource] = settings.GetEnvironment().ResourceManagerEndpoint

	// Generate the MSI authorizer by settings provided and factory's configurations
	authorizer, err := factory.newAuthorizer(settings)
	return authorizer, err
}

// newAuthorizer Creates a new authorizer from settings provided.
// The function adds the factory configured user assigned MSI client id to settings value map and generates
// an authorizer using setting' (auth.EnvironmentSettings) GetAuthorizer.
// If factory is configured in local development mode (configuration.isLocalDevelopmentMode == true):
// creates an authorizer from azure cli with no logged-in user as ID.
func (factory *EnvAzureAuthorizerFactory) newAuthorizer(settings wrappers.IEnvironmentSettingsWrapper) (autorest.Authorizer, error) {

	if settings == nil {
		return nil, fmt.Errorf("null argument settings")
	}
	// Set client id for user managed identity (empty for system manged identity)
	settings.GetValues()[auth.ClientID] = factory.configuration.MSIClientId

	// If not
	if !factory.configuration.isLocalDevelopmentMode {
		return settings.GetAuthorizer()
	}
	// Else - Local development
	return factory.authWrapper.NewAuthorizerFromCLIWithResource(settings.GetValues()[auth.Resource])
}
