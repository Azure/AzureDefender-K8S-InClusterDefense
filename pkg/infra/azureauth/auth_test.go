package azureauth

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	clientId string = "fakeClientId"
)

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct
type ExampleTestSuite struct {
	suite.Suite
}

var authWrapperMock IAzureAuthWrapper
var authSettingsMock IEnvironmentSettingsWrapper
var factory *AzureMSIAuthroizerFactory

var configuration = &AzureMSIAuthroizerConfiguration{
	isLocalDevelopmentMode: false,
	MSIClientId:            clientId,
}

// This will run before each test in the suite
func (suite *ExampleTestSuite) SetupTest() {

	//authSettingsMock := &mocks.IEnvironmentSettingsWrapper{}

	//factory = NewAzureMSIAuthroizerFactory(configuration, authWrapperMock)

}

// This is an example test that will always succeed
func (suite *ExampleTestSuite) TestExample() {
	suite.Equal(true, true)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}
