package wrappers

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// IEnvironmentSettingsWrapper Interfaces for auth wrappers
type IEnvironmentSettingsWrapper interface {
	GetMSIAuthorizer() (autorest.Authorizer, error)
	GetValues() map[string]string
	GetEnvironment() *azure.Environment
}

// EnvironmentSettingsWrapper implements IEnvironmentSettingsWrapper interface
var _ IEnvironmentSettingsWrapper = (*EnvironmentSettingsWrapper)(nil)

// EnvironmentSettingsWrapper wraps auth.EnvironmentSettings implements wrapper.IEnvironmentSettingsWrapper interface
type EnvironmentSettingsWrapper struct {
	settings *auth.EnvironmentSettings
}

// NewEnvironmentSettingsWrapper Constructor EnvironmentSettingsWrapper
func NewEnvironmentSettingsWrapper(settings *auth.EnvironmentSettings) *EnvironmentSettingsWrapper {
	return &EnvironmentSettingsWrapper{
		settings: settings,
	}
}

// GetMSIAuthorizer get authorizer from MSI config from settings of wrapped settings
func (wrapper *EnvironmentSettingsWrapper) GetMSIAuthorizer() (autorest.Authorizer, error) {
	authorizer, err := wrapper.settings.GetMSI().Authorizer()
	if err != nil {
		return nil, err
	}

	return authorizer, nil
}

// GetValues get settings value map from wrapped settings
func (wrapper *EnvironmentSettingsWrapper) GetValues() map[string]string {
	return wrapper.settings.Values
}

// GetEnvironment get auth GetEnvironment from wrapped settings
func (wrapper *EnvironmentSettingsWrapper) GetEnvironment() *azure.Environment {
	return &wrapper.settings.Environment
}
