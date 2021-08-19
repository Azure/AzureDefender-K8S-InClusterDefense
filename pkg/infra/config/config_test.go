package config

import (
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	_configurationName string = "TestConfig"
	_configurationType string = "yaml"
	_configurationPath string = "/testdata"
	_readFromEnv bool = true
)

type Clothing struct {
	Jacket string
	Shoes int
	Trousers string
}

// test suite struct
type TestSuite struct {
	suite.Suite
	config *ConfigurationProvider
	exampleConfig map[string]interface{}
}


func createExampleMapString() map[string]interface{}{
	yamlExample := make(map[string]interface{})
	yamlExample["name"] = "Steve"
	yamlExample["hobbies"] = []interface {}{"skateboarding", "snowboarding", "go"}
	yamlExample["clothing"] = map[string]interface {}{"jacket":"leather", "shoes": 45, "trousers":"denim"}
	yamlExample["age"] = 35
	yamlExample["eyes"] = "blue"
	yamlExample["beard"] = true
	return yamlExample
}

// This will run before each test in the suit
func (suite *TestSuite) SetupTest(){
	config, err := LoadConfig(_configurationName, _configurationType, _configurationPath, _readFromEnv)
	suite.config = config
	suite.exampleConfig = createExampleMapString()
	suite.Nil(err)
}

// Test loaded configuration file is not null or empty
func (suite *TestSuite) TestConfig_LoadConfiguration_NonEmptyConfig() {
	suite.NotEmpty(suite.config.viperConfig)
	allSettings := suite.config.AllSettings()
	assert.Equal(suite.T(), allSettings, suite.exampleConfig)
}

// Test Sub method work properly
func (suite *TestSuite) TestConfig_SubConfiguration() {
	assert.Equal(suite.T(), suite.exampleConfig["clothing"], suite.config.SubConfig("clothing").AllSettings())
	subConfig := suite.config.SubConfig("dateOfBirth")
	assert.True(suite.T(), reflect2.IsNil(subConfig.viperConfig))
	subConfig = suite.config.SubConfig("trousers")
	assert.True(suite.T(), reflect2.IsNil(subConfig.viperConfig))
}

// Test Unmarshal method work properly
func (suite *TestSuite) TestConfig_UnmarshalConfiguration() {
	clothes := new(Clothing)
	subConfig := suite.config.SubConfig("clothing")
	err := subConfig.Unmarshal(clothes)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "leather", clothes.Jacket)
	assert.Equal(suite.T(), 45, clothes.Shoes)
	assert.Equal(suite.T(), "denim", clothes.Trousers)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestConfigTestSuite(t *testing.T) {
	settings, _ := auth.GetSettingsFromEnvironment()
	print(settings.Environment.Name)
	suite.Run(t, new(TestSuite))
}
