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

func (suite *ValidateVariablesTestSuite) Test_ValidatePositiveInt_Valid(){
	result := ValidatePositiveInt(1,2,3,4,5)
	suite.Equal(true, result)
}


func (suite *ValidateVariablesTestSuite) Test_ValidatePositiveInt_Zero(){
	result := ValidatePositiveInt(1,2,3,0,5)
	suite.Equal(false, result)
}


func (suite *ValidateVariablesTestSuite) Test_ValidatePositiveInt_Neg(){
	result := ValidatePositiveInt(1,-2,3,4,5)
	suite.Equal(false ,result)
}


func Test_ValidateVariablesTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateVariablesTestSuite))
}
