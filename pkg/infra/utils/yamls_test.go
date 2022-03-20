package utils

import (
	"github.com/pkg/errors"
	"os"
	yaml1 "sigs.k8s.io/kustomize/kyaml/yaml"
)

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type YamlsTestSuite struct {
	suite.Suite
	File *yaml1.RNode
}

func (suite *YamlsTestSuite) SetupTest() {
	jsonFile, err := os.ReadFile("yaml_test.json")
	if err != nil {
		panic(suite)
	}
	yamlFile, err := yaml1.ConvertJSONToYamlNode(string(jsonFile))
	if err != nil {
		panic(suite)
	}
	suite.File = yamlFile

}

func (suite *YamlsTestSuite) Test_GoToDestNode_DestNodeExist() {
	dest, err := GoToDestNode(suite.File, "metadata")
	suite.Nil(err)
	str, err := dest.GetString("name")
	suite.Nil(err)
	suite.Equal(str, "name")
}

func (suite *YamlsTestSuite) Test_GoToDestNode_DestNodeNotExist() {
	dest, err := GoToDestNode(suite.File, "metadata", "field")
	suite.Nil(dest)
	suite.True(errors.Is(err, _errInvalidPath))
}

func Test_YamlsTestSuite(t *testing.T) {
	suite.Run(t, new(YamlsTestSuite))
}
