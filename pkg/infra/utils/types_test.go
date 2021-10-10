package utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuite struct {
	suite.Suite
}

type StructForTest struct{}

func (suite *TestSuite) Test_GetType_Struct() {
	// Setup
	instance := StructForTest{}
	expected := "utils.StructForTest"

	// Act
	actual := GetType(instance)

	// Test
	suite.Equal(expected, actual)
}

func (suite *TestSuite) Test_GetType_Nil() {
	// Setup
	expected := "<nil>"

	// Act
	actual := GetType(nil)

	// Test
	suite.Equal(expected, actual)
}

func (suite *TestSuite) Test_GetTypeWithoutPackage_Struct() {
	// Setup
	instance := StructForTest{}
	expected := "StructForTest"

	// Act
	actual := GetTypeWithoutPackage(instance)

	// Test
	suite.Equal(expected, actual)
}

func (suite *TestSuite) Test_GetTypeWithoutPackage_Nil() {
	// Setup
	expected := "<nil>"

	// Act
	actual := GetTypeWithoutPackage(nil)

	// Test
	suite.Equal(expected, actual)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
