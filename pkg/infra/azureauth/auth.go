package azureauth

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

type AzureAuthroizerFactory struct {
	configuration *AzureAuthroizerConfiguration
}

type AzureAuthroizerConfiguration struct {
	isLocalDevelopmentMode bool
	clientId               string
}

func (factory *AzureAuthroizerFactory) NewARMAuthorizer() (autorest.Authorizer, error) {

	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return nil, err
	}

	settings.Values[auth.Resource] = settings.Environment.ResourceManagerEndpoint

	authorizer, err := factory.newAuthorizer(settings)
	return authorizer, err

}

func (factory *AzureAuthroizerFactory) newAuthorizer(settings auth.EnvironmentSettings) (autorest.Authorizer, error) {

	var authorizer autorest.Authorizer
	var err error

	// Set client id for user managed identity (empty for system manged identity)
	settings.Values[auth.ClientID] = factory.configuration.clientId

	if !factory.configuration.isLocalDevelopmentMode {
		authorizer, err = settings.GetAuthorizer()
	} else {
		// Local development
		authorizer, err = auth.NewAuthorizerFromCLIWithResource(settings.Values[auth.Resource])
	}

	return authorizer, err
}
