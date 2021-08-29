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
	CLIENT_ID string = "fakeClientId"
)

var _configuration = &EnvAzureAuthorizerConfiguration{
	IsLocalDevelopmentMode: false,
	MSIClientId:            CLIENT_ID,
}

var _expectedValues = map[string]string{
	auth.ClientID: CLIENT_ID,
	auth.Resource: azure.PublicCloud.ResourceManagerEndpoint,
}

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct
type TestSuite struct {
	suite.Suite
	authWrapperMock  *mocks.IAzureAuthWrapper
	authSettingsMock *mocks.IEnvironmentSettingsWrapper
	factory          *EnvAzureAuthorizerFactory
	values           map[string]string
	env              *azure.Environment
	authorizer       autorest.Authorizer
}

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	suite.values = map[string]string{}
	suite.env = &azure.PublicCloud
	suite.authSettingsMock = &mocks.IEnvironmentSettingsWrapper{}
	suite.authWrapperMock = &mocks.IAzureAuthWrapper{}
	suite.factory = NewEnvAzureAuthorizerFactory(_configuration, suite.authWrapperMock)
	suite.authorizer = autorest.NullAuthorizer{}
}

// This is an example test that will always succeed
func (suite *TestSuite) TestAzureAuthorizerFromEnvFactory_CreateArmAuthorizer_NonDevelopmentMode_ClientIdAndResourceAuthUsingEnv() {

	suite.authSettingsMock.On("GetEnvironment").Return(suite.env).Once()
	suite.authSettingsMock.On("GetValues").Return(suite.values).Twice()
	suite.authSettingsMock.On("GetAuthorizer").Return(suite.authorizer, nil).Once()
	suite.authWrapperMock.On("GetSettingsFromEnvironment").Return(suite.authSettingsMock, nil).Once()
	authorizer, err := suite.factory.CreateARMAuthorizer()

	suite.Nil(err)
	suite.Equal(suite.authorizer, authorizer)
	suite.Equal(_expectedValues, suite.values)
	assertExpectations(suite)
}

func (suite *TestSuite) TestEnvAzureAuthorizerFactory_CreateArmAuthorizer_DevelopmentMode_ResourceAuthUsingCLI() {

	_configuration.IsLocalDevelopmentMode = true
	suite.authSettingsMock.On("GetEnvironment").Return(suite.env).Once()
	suite.authSettingsMock.On("GetValues").Return(suite.values).Times(3)
	suite.authWrapperMock.On("GetSettingsFromEnvironment").Return(suite.authSettingsMock, nil).Once()
	suite.authWrapperMock.On("NewAuthorizerFromCLIWithResource", _expectedValues[auth.Resource]).Return(suite.authorizer, nil).Once()

	authorizer, err := suite.factory.CreateARMAuthorizer()

	suite.Nil(err)
	suite.Equal(suite.authorizer, authorizer)
	suite.Equal(_expectedValues, suite.values)
	assertExpectations(suite)
}

func assertExpectations(suite *TestSuite) {
	suite.authSettingsMock.AssertExpectations(suite.T())
	suite.authWrapperMock.AssertExpectations(suite.T())
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestAzureAuthorizerFromEnvFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
