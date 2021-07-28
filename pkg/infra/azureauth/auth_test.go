package azureauth

import (
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	string clientId = "fakeClientId"
)

type MockAzureAuthWrapper struct {
	mock.Mock
}

func init() (IAzureAuthroizerFactory factory) {
	factory := &NewAzureMSIAuthroizerFactory(
		&AzureMSIAuthroizerConfiguration{
			isLocalDevelopmentMode: false,
			MSIClientId:            clientId,
		},
		nil,
	)
}
func TestNewARMAuthorizer(t *testing.T) {
	a := new(MockAzureAuthWrapper)
	a.On()
	assert.Equal(t, "tomer", "tomerw")
}

func (wrapper *MockAzureAuthWrapper) GetSettingsFromEnvironment() (*IEnvironmentSettingsWrapper, error) {
	return nil, nil
}

func (wrapper *MockAzureAuthWrapper) NewAuthorizerFromCLIWithResource(resource string) (*autorest.Authorizer, error) {
	return nil, nil
}
