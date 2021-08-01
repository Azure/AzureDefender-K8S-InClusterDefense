package wrappers

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

// Interfaces for auth wrappers
type IAzureAuthWrapper interface {
	GetSettingsFromEnvironment() (IEnvironmentSettingsWrapper, error)
	NewAuthorizerFromCLIWithResource(string) (autorest.Authorizer, error)
}

// Interfaces for auth wrappers
type IEnvironmentSettingsWrapper interface {
	GetAuthorizer() (autorest.Authorizer, error)
	Values() map[string]string
	Environment() *azure.Environment
}
