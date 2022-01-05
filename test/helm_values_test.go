package test

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct
type TestSuite struct {
	suite.Suite
}

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
}

func (suite *TestSuite) Test_AreValuesFilesOfHelmHaveTheSameKeys() {
	// Setup
	pwd, _ := os.Getwd()
	parentPwd := filepath.Dir(pwd)
	valuesPathFile := filepath.Join(parentPwd, "charts", "azdproxy", "values.yaml")
	valuesDevPathFile := filepath.Join(parentPwd, "charts", "azdproxy", "values-dev.yaml")
	// Act
	areEqual, err := utils.CheckIfTwoYamlsHaveTheSameKeys(valuesPathFile, valuesDevPathFile)
	//Test
	suite.Nil(err)
	suite.True(areEqual)
}

func (suite *TestSuite) Test_AreConfigurationFilesHaveTheSameKeys() {
	//TODO implement this method. there is a problem how to read the configuration.yaml file because the value of data is string and not map.
	//// Setup
	//pwd, _ := os.Getwd()
	//parentPwd := filepath.Dir(pwd)
	//configPath := filepath.Join(parentPwd, "charts", "azdproxy", "templates", "configuration.yaml")
	//configDebugPath := filepath.Join(parentPwd, "config", "AppConfigDebug.yaml")
	//
	//configMap, err := utils.CreateMapFromPathOfYaml(configPath)
	//suite.Nil(err)
	//configMapData, ok := configMap["data"]
	//suite.False(ok)
	//
	//configMapDataAsMap, ok := configMapData.(map[string]interface{})
	//suite.False(ok)
	//
	//configDebugMap, err := utils.CreateMapFromPathOfYaml(configDebugPath)
	//suite.Nil(err)
	//
	//// Act
	//areEqual := utils.AreMapsHaveTheSameKeys(configMapDataAsMap, configDebugMap)
	//
	////Test
	//suite.True(areEqual)
}

// TestHelmValuesSuite We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestHelmValuesSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
