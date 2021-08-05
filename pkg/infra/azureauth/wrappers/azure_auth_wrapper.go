package wrappers

import (
	"github.com/Azure/go-autorest/autorest"
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
