package azureauth

import (
	"github.com/pkg/errors"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// Errors
var (
	// _nullArgErr is error that is thrown when there is unexpected null argument.
	_nullArgErr = errors.New("null argument")
)

// IAzureAuthorizerFactory Factory to create an azure authorizer
type IAzureAuthorizerFactory interface {
	// CreateARMAuthorizer Generates a new ARM azure client authorizer
	CreateARMAuthorizer() (autorest.Authorizer, error)
}

// EnvAzureAuthorizerFactory implements IAzureAuthorizerFactory interface
var _ IAzureAuthorizerFactory = (*EnvAzureAuthorizerFactory)(nil)

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
	// IsLocalDevelopmentMode is factory set to local development
	IsLocalDevelopmentMode bool

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

// CreateARMAuthorizer Generates a new ARM azure client authorizer using MSI configured
func (factory *EnvAzureAuthorizerFactory) CreateARMAuthorizer() (autorest.Authorizer, error) {

	// Gets authorizer setting from environment
	settings, err := factory.authWrapper.GetSettingsFromEnvironment()
	if err != nil {
		// Error in fetching settings
		return nil, err
	}

	// Set ARM as the resource to authorize in settings
	settings.GetValues()[auth.Resource] = settings.GetEnvironment().ResourceManagerEndpoint

	// Generate the MSI authorizer by settings provided and factory's configurations
	authorizer, err := factory.createAuthorizer(settings)
	return authorizer, err
}

// createAuthorizer Creates a new authorizer from settings provided.
// The function adds the factory configured user assigned MSI client id to settings value map and generates
// an authorizer using setting' (auth.EnvironmentSettings) GetAuthorizer.
// If factory is configured in local development mode (configuration.IsLocalDevelopmentMode == true):
// creates an authorizer from azure cli with no logged-in user as ID.
func (factory *EnvAzureAuthorizerFactory) createAuthorizer(settings wrappers.IEnvironmentSettingsWrapper) (autorest.Authorizer, error) {

	if settings == nil {
		return nil, errors.Wrap(_nullArgErr, "null argument settings in createAuthorizer")
	}
	// Set client id for user managed identity (empty for system manged identity)
	settings.GetValues()[auth.ClientID] = factory.configuration.MSIClientId

	// If not
	if !factory.configuration.IsLocalDevelopmentMode {
		return settings.GetAuthorizer()
	}
	// Else - Local development
	return factory.authWrapper.NewAuthorizerFromCLIWithResource(settings.GetValues()[auth.Resource])
}
