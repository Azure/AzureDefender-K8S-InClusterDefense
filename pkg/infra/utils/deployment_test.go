package utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	// Reset singleton to nil each test! The test will be failed if you delete this
	_singleton = nil
}

func (suite *TestSuite) Test_NewDeployment_Create2Instances_ShouldReturnError() {
	// Setup
	d, err := NewDeployment(&DeploymentConfiguration{IsLocalDevelopment: true, Namespace: "kube-system"})
	suite.NotNil(d)
	suite.Nil(err)
	// Act
	d, err = NewDeployment(&DeploymentConfiguration{IsLocalDevelopment: true, Namespace: "kube-system"})
	// Test
	suite.Nil(d)
	suite.NotNil(err)
}

func (suite *TestSuite) Test_NewDeployment_NilConfiguration_ShouldReturnError() {
	// Act
	d, err := NewDeployment(nil)
	// Test
	suite.Nil(d)
	suite.NotNil(err)
}

func (suite *TestSuite) Test_GetDeploymentInstance_FailureOnNewDeployment__ShouldReturnSameInstance() {
	// Setup
	expected, _ := NewDeployment(&DeploymentConfiguration{})
	_, _ = NewDeployment(&DeploymentConfiguration{})
	// Act
	actual := GetDeploymentInstance()
	// Test
	suite.Equal(expected, actual)
}

func (suite *TestSuite) Test_GetDeploymentInstance_BeforeInitialized__ShouldReturnError() {
	// Act
	actual := GetDeploymentInstance()
	// Test
	suite.Nil(actual)
}

func (suite *TestSuite) Test_IsLocalDevelopment_IsTrue__ShouldReturnTrue() {
	// Setup
	d, _ := NewDeployment(&DeploymentConfiguration{IsLocalDevelopment: true})
	// Act
	actual := d.IsLocalDevelopment()
	// Test
	suite.True(actual)
}

func (suite *TestSuite) Test_IsLocalDevelopment_IsFalse__ShouldReturnFalse() {
	// Setup
	d, _ := NewDeployment(&DeploymentConfiguration{IsLocalDevelopment: false})
	// Act
	actual := d.IsLocalDevelopment()
	// Test
	suite.False(actual)
}

func (suite *TestSuite) Test_GetNamespace_SetKubeSystem__ShouldReturnKubeSystem() {
	// Setup
	d, _ := NewDeployment(&DeploymentConfiguration{Namespace: "kube-system"})
	// Act
	actual := d.GetNamespace()
	// Test
	suite.Equal("kube-system", actual)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestDeploymentUtils(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
