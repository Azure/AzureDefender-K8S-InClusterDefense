package utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ArraysTestSuite struct {
	suite.Suite
	list []string
}

func (suite * ArraysTestSuite) SetupTest() {
	suite.list = []string{"hello","world","!"}
}

func (suite * ArraysTestSuite) Test_StringInSlice_StringExistsInSlice() {
	suite.True(StringInSlice("hello",suite.list))
}

func (suite * ArraysTestSuite) Test_StringInSlice_StringNotExistsInSlice() {
	suite.False(StringInSlice("str",suite.list))
}

func Test_ArraysTestSuite(t *testing.T) {
	suite.Run(t, new(ArraysTestSuite))
}