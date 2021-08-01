package azureauth

import (
	"fmt"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// Factory to create an azure authorizer
type IAzureAuthroizerFactory interface {
	// Generates a new ARM azure client authorizer
	NewARMAuthorizer() (autorest.Authorizer, error)
}

// Factory to create an azure authorizer using managed identity
// Implements azureauth.IAzureAuthroizerFactory
type AzureAuthroizerFromEnvFactory struct {
	configuration *AzureAuthroizerFromEnvConfiguration
	authWrapper   wrappers.IAzureAuthWrapper
}

// Factory configuration to create an azure authorizer from environment (env varaibles or managed identity)
type AzureAuthroizerFromEnvConfiguration struct {
	// Is factory set to local development
	isLocalDevelopmentMode bool

	// MSI client id
	MSIClientId string
}

// Constructor
func NewAzureAuthroizerFromEnvFactory(configuration *AzureAuthroizerFromEnvConfiguration, authWrapper wrappers.IAzureAuthWrapper) *AzureAuthroizerFromEnvFactory {
	return &AzureAuthroizerFromEnvFactory{
		configuration: configuration,
		authWrapper:   authWrapper,
	}
}

// Generates a new ARM azure client authorizer using MSI configured
func (factory *AzureAuthroizerFromEnvFactory) NewARMAuthorizer() (autorest.Authorizer, error) {

	// Gets authorizer setting from environment
	settings, err := factory.authWrapper.GetSettingsFromEnvironment()
	if err != nil {
		// Error in fetching settings
		return nil, err
	}

	// Set ARM as the resource to authorize in settings
	settings.Values()[auth.Resource] = settings.Environment().ResourceManagerEndpoint

	// Generate the MSI authorizer by settings provided and factory's configurations
	authorizer, err := factory.newAuthorizer(settings)
	return authorizer, err
}

// Creates a new authorizer from settings provided.
// The function adds the factory configured user assigned MSI client id to settings value map and generates
// an authorizer using setting' (auth.EnvironmentSettings) GetAuthorizer.
// If factory is configured in local development mode (configuration.isLocalDevelopmentMode == true):
// creates an authorizer from azure cli with no logged in user as Id.
func (factory *AzureAuthroizerFromEnvFactory) newAuthorizer(settings wrappers.IEnvironmentSettingsWrapper) (autorest.Authorizer, error) {

	if settings == nil {
		return nil, fmt.Errorf("null argument settings")
	}
	// Set client id for user managed identity (empty for system manged identity)
	settings.Values()[auth.ClientID] = factory.configuration.MSIClientId

	// If not
	if !factory.configuration.isLocalDevelopmentMode {
		return settings.GetAuthorizer()
	} else {
		// Local development
		return factory.authWrapper.NewAuthorizerFromCLIWithResource(settings.Values()[auth.Resource])
	}
}

// Wrapper for azure/auth package used actions
type AzureAuthWrapper struct{}

func (wrapper *AzureAuthWrapper) GetSettingsFromEnvironment() (wrappers.IEnvironmentSettingsWrapper, error) {

	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return nil, err
	}
	return NewEnvironmentSettingsWrapper(&settings), nil
}

func (wrapper *AzureAuthWrapper) NewAuthorizerFromCLIWithResource(resource string) (autorest.Authorizer, error) {

	authorizer, err := auth.NewAuthorizerFromCLIWithResource(resource)
	if err != nil {
		return nil, err
	}
	return authorizer, nil
}

type EnvironmentSettingsWrapper struct{ settings *auth.EnvironmentSettings }

func NewEnvironmentSettingsWrapper(settings *auth.EnvironmentSettings) wrappers.IEnvironmentSettingsWrapper {
	return &EnvironmentSettingsWrapper{
		settings: settings,
	}
}

func (wrapper *EnvironmentSettingsWrapper) GetAuthorizer() (autorest.Authorizer, error) {
	authorizer, err := wrapper.settings.GetAuthorizer()
	if err != nil {
		return nil, err
	}

	return authorizer, nil
}

func (wrapper *EnvironmentSettingsWrapper) Values() map[string]string {
	return wrapper.settings.Values
}

func (wrapper *EnvironmentSettingsWrapper) Environment() *azure.Environment {
	return &wrapper.settings.Environment
}
