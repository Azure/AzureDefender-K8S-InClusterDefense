package utils

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ErrForTest1 struct {
	msg string
}

func (err ErrForTest1) Error() string {
	return err.msg
}

type ErrForTest2 struct{}

func (err ErrForTest2) Error() string {
	return "ErrForTest2 msg"
}

func (suite *TestSuite) Test_IsErrorIsTypeOf_ErrorIsOfInputType_ShouldReturnTrue() {
	// Setup
	err := &ErrForTest1{}
	typeErr := ErrForTest1{}
	ptrTypeErr := &typeErr

	// Act
	actual := IsErrorIsTypeOf(err, &ptrTypeErr)

	// Test
	suite.True(actual)
}

func (suite *TestSuite) Test_IsErrorIsTypeOf_ErrorIsOfInputTypeButWrapped_ShouldReturnTrue() {
	// Setup
	err := errors.Wrap(&ErrForTest1{}, "wrapping")
	typeErr := ErrForTest1{}
	ptrTypeErr := &typeErr

	// Act
	actual := IsErrorIsTypeOf(err, &ptrTypeErr)

	// Test
	suite.True(actual)
}

func (suite *TestSuite) Test_IsErrorIsTypeOf_ErrorIsOfInputTypeWithDiffMsg_ShouldReturnTrue() {
	// Setup
	err := &ErrForTest1{msg: "Lior"}
	typeErr := ErrForTest1{msg: "Tomer"}
	ptrTypeErr := &typeErr

	// Act
	actual := IsErrorIsTypeOf(err, &ptrTypeErr)

	// Test
	suite.True(actual)
}

func (suite *TestSuite) Test_IsErrorIsTypeOf_ErrorIsNotOfInputType_ShouldReturnFalse() {
	// Setup
	err := &ErrForTest1{}
	typeErr := ErrForTest2{}
	ptrTypeErr := &typeErr

	// Act
	actual := IsErrorIsTypeOf(err, &ptrTypeErr)

	// Test
	suite.False(actual)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
