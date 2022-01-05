package utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ValidateVariablesTestSuite struct {
	suite.Suite
}

func (suite *ValidateVariablesTestSuite) SetupTest() {

}

func (suite *ValidateVariablesTestSuite) Test_ValidatePositiveInt_Valid() {
	result, name := ValidatePositiveInt(&PositiveIntValidationObject{"1", 1}, &PositiveIntValidationObject{"2", 2}, &PositiveIntValidationObject{"3", 3})
	suite.Equal(true, result)
	suite.Equal("", name)
}

func (suite *ValidateVariablesTestSuite) Test_ValidatePositiveInt_Zero() {
	result, name := ValidatePositiveInt(&PositiveIntValidationObject{"1", 1}, &PositiveIntValidationObject{"2", 2}, &PositiveIntValidationObject{"0", 0})
	suite.Equal(false, result)
	suite.Equal("0", name)
}

func (suite *ValidateVariablesTestSuite) Test_ValidatePositiveInt_Neg() {
	result, name := ValidatePositiveInt(&PositiveIntValidationObject{"1", 1}, &PositiveIntValidationObject{"-2", -2}, &PositiveIntValidationObject{"0", 0})
	suite.Equal(false, result)
	suite.Equal("-2", name)
}

func Test_ValidateVariablesTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateVariablesTestSuite))
}
