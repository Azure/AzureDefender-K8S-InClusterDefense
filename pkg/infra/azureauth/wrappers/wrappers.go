package wrappers

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// IAzureAuthWrapper Interfaces for auth wrappers
type IAzureAuthWrapper interface {
	GetSettingsFromEnvironment() (IEnvironmentSettingsWrapper, error)
	NewAuthorizerFromCLIWithResource(string) (autorest.Authorizer, error)
}

// AzureAuthWrapper Wrapper for azure/auth package used actions implements IAzureAuthWrapper interface
type AzureAuthWrapper struct{}

// GetSettingsFromEnvironment get setting from auth.GetSettingsFromEnvironment and wraps it
func (wrapper *AzureAuthWrapper) GetSettingsFromEnvironment() (IEnvironmentSettingsWrapper, error) {

	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return nil, err
	}
	return NewEnvironmentSettingsWrapper(&settings), nil
}

// NewAuthorizerFromCLIWithResource get authorizer from auth.NewAuthorizerFromCLIWithResource
func (wrapper *AzureAuthWrapper) NewAuthorizerFromCLIWithResource(resource string) (autorest.Authorizer, error) {

	authorizer, err := auth.NewAuthorizerFromCLIWithResource(resource)
	if err != nil {
		return nil, err
	}
	return authorizer, nil
}

// IEnvironmentSettingsWrapper Interfaces for auth wrappers
type IEnvironmentSettingsWrapper interface {
	GetAuthorizer() (autorest.Authorizer, error)
	Values() map[string]string
	Environment() *azure.Environment
}

// EnvironmentSettingsWrapper wraps auth.EnvironmentSettings implements wrapper.IEnvironmentSettingsWrapper interface
type EnvironmentSettingsWrapper struct{ settings *auth.EnvironmentSettings }

// NewEnvironmentSettingsWrapper Constructor EnvironmentSettingsWrapper
func NewEnvironmentSettingsWrapper(settings *auth.EnvironmentSettings) *EnvironmentSettingsWrapper {
	return &EnvironmentSettingsWrapper{
		settings: settings,
	}
}

// GetAuthorizer get autorizer from wrapped settings
func (wrapper *EnvironmentSettingsWrapper) GetAuthorizer() (autorest.Authorizer, error) {
	authorizer, err := wrapper.settings.GetAuthorizer()
	if err != nil {
		return nil, err
	}

	return authorizer, nil
}

// Values get settings value map from wrapped settings
func (wrapper *EnvironmentSettingsWrapper) Values() map[string]string {
	return wrapper.settings.Values
}

// Environment get auth Environment from wrapped settings
func (wrapper *EnvironmentSettingsWrapper) Environment() *azure.Environment {
	return &wrapper.settings.Environment
}
