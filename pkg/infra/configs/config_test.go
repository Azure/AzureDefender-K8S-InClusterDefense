package configs

import (
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	CONFIGURATION_NAME string = "TestConfig"
	CONFIGURATION_TYPE string = "yaml"
	CONFIGURATION_PATH string = "."
	READ_FROM_ENV bool = true
)

type Clothing struct {
	Jacket string
	Shoes int
	Trousers string
}

// test suite struct
type TestSuite struct {
	suite.Suite
	config *Configuration
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
	config, err := NewConfiguration(CONFIGURATION_NAME, CONFIGURATION_TYPE, CONFIGURATION_PATH, READ_FROM_ENV)
	suite.config = config
	suite.exampleConfig = createExampleMapString()
	suite.Nil(err)
}

// Test loaded configuration file is not null or empty
func (suite *TestSuite) TestConfigurationHasLoadedSuccessfully() {
	suite.NotEmpty(suite.config.viperConfig)
	allSettings := suite.config.AllSettings()
	assert.Equal(suite.T(), allSettings, suite.exampleConfig)
}

// Test Sub method work properly
func (suite *TestSuite) TestConfigurationSub() {
	assert.Equal(suite.T(), suite.exampleConfig["clothing"], suite.config.SubConfig("clothing").AllSettings())
	subConfig := suite.config.SubConfig("dateOfBirth")
	assert.True(suite.T(), reflect2.IsNil(subConfig.viperConfig))
	subConfig = suite.config.SubConfig("trousers")
	assert.True(suite.T(), reflect2.IsNil(subConfig.viperConfig))
}

// Test Unmarshal method work properly
func (suite *TestSuite) TestConfigurationUnmarshal() {
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
	suite.Run(t, new(TestSuite))
}
