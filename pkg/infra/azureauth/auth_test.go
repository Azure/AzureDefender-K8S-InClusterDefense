package azureauth

import (
	"testing"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/mocks"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/stretchr/testify/suite"
)

const (
	clientId string = "fakeClientId"
)

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct
type TestSuite struct {
	suite.Suite
	authWrapperMock  *mocks.IAzureAuthWrapper
	authSettingsMock *mocks.IEnvironmentSettingsWrapper
	factory          *AzureAuthroizerFromEnvFactory
	values           map[string]string
	env              *azure.Environment
	authorizer       autorest.Authorizer
}

var configuration = &AzureAuthroizerFromEnvConfiguration{
	isLocalDevelopmentMode: false,
	MSIClientId:            clientId,
}

var expectedValues = map[string]string{
	auth.ClientID: clientId,
	auth.Resource: azure.PublicCloud.ResourceManagerEndpoint,
}

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	suite.values = map[string]string{}
	suite.env = &azure.PublicCloud
	suite.authSettingsMock = &mocks.IEnvironmentSettingsWrapper{}
	suite.authWrapperMock = &mocks.IAzureAuthWrapper{}
	suite.factory = NewAzureAuthroizerFromEnvFactory(configuration, suite.authWrapperMock)
	suite.authorizer = autorest.NullAuthorizer{}
}

// This is an example test that will always succeed
func (suite *TestSuite) TestAzureAuthroizerFromEnvFactory_NewArmAuthorizer_NonDevelopmentMode_ClientIdAndResourceAuthUsingEnv() {

	suite.authSettingsMock.On("Environment").Return(suite.env).Once()
	suite.authSettingsMock.On("Values").Return(suite.values).Twice()
	suite.authSettingsMock.On("GetAuthorizer").Return(suite.authorizer, nil).Once()
	suite.authWrapperMock.On("GetSettingsFromEnvironment").Return(suite.authSettingsMock, nil).Once()
	authorizer, err := suite.factory.NewARMAuthorizer()

	suite.Nil(err)
	suite.Equal(suite.authorizer, authorizer)
	suite.Equal(expectedValues, suite.values)
	assertExcpectations(suite)
}

func (suite *TestSuite) TestAzureAuthroizerFromEnvFactory_NewArmAuthorizer_DevelopmentMode_ResourceAuthUsingCLI() {

	configuration.isLocalDevelopmentMode = true
	suite.authSettingsMock.On("Environment").Return(suite.env).Once()
	suite.authSettingsMock.On("Values").Return(suite.values).Times(3)
	suite.authWrapperMock.On("GetSettingsFromEnvironment").Return(suite.authSettingsMock, nil).Once()
	suite.authWrapperMock.On("NewAuthorizerFromCLIWithResource", expectedValues[auth.Resource]).Return(suite.authorizer, nil).Once()

	authorizer, err := suite.factory.NewARMAuthorizer()

	suite.Nil(err)
	suite.Equal(suite.authorizer, authorizer)
	suite.Equal(expectedValues, suite.values)
	assertExcpectations(suite)
}

func assertExcpectations(suite *TestSuite) {
	suite.authSettingsMock.AssertExpectations(suite.T())
	suite.authWrapperMock.AssertExpectations(suite.T())
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestAzureAuthroizerFromEnvFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
