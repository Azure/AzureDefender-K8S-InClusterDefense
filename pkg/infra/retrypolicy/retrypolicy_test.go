package retrypolicy

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	_configuration                 = RetryPolicyConfiguration{RetryAttempts: 4, RetryDuration: "3ms"}
	_expectedConfigurationDuration, _ = time.ParseDuration(_configuration.RetryDuration)
)

type TestSuite struct {
	suite.Suite
	retryPolicy  *RetryPolicy
	countActions int
}

var _ error = (*err1ForTests)(nil)
var _ error = (*err2ForTests)(nil)

type (
	// err1ForTests is struct for test error type
	err1ForTests struct{}
	// err2ForTests is struct for test error type
	err2ForTests struct{}
)

func (e err1ForTests) Error() string {
	return "err1ForTests"
}
func (e err2ForTests) Error() string {
	return "err2ForTests"
}

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	// Mock
	err := *new(error)
	suite.retryPolicy, err = NewRetryPolicy(instrumentation.NewNoOpInstrumentationProvider(), &_configuration)
	suite.Nil(err)
	suite.countActions = 0
}

func (suite *TestSuite) Test_NewRetryPolicy_NotNilConfiguration_ShouldReturnInstance() {

	r, err := NewRetryPolicy(instrumentation.NewNoOpInstrumentationProvider(), &_configuration)
	suite.Nil(err)
	suite.Equal(_configuration.RetryAttempts, r.retryAttempts)
	suite.Equal(_expectedConfigurationDuration, r.duration)
}

func (suite *TestSuite) Test_GetBackOffDuration_NotNilConfiguration_ShouldReturnInstance() {
	d, err := _configuration.GetBackOffDuration()
	suite.Nil(err)
	suite.Equal(_expectedConfigurationDuration, d)
}

func (suite *TestSuite) Test_RetryActionString_ActionNil_ShouldReturnError() {
	// Setup
	var action ActionString = nil
	var handle ShouldRetryOnSpecificError = func(error) bool { suite.countActions += 1; return false }
	// Act
	actual, err := suite.retryPolicy.RetryActionString(action, handle)
	// Test
	suite.Equal(utils.NilArgumentError, errors.Cause(err))
	suite.Equal("", actual)
	suite.Equal(0, suite.countActions)
}

func (suite *TestSuite) Test_RetryActionString_HandleNil_ShouldReturnError() {
	// Setup
	var action ActionString = func() (string, error) { suite.countActions += 1; return "lior", nil }
	var handle ShouldRetryOnSpecificError = nil
	// Act
	actual, err := suite.retryPolicy.RetryActionString(action, handle)
	// Test
	suite.Equal(utils.NilArgumentError, errors.Cause(err))
	suite.Equal("", actual)
	suite.Equal(0, suite.countActions)
}

func (suite *TestSuite) Test_RetryActionString_HandledError_ShouldBeExecutedOnce() {
	// Setup
	errForTest := &err1ForTests{}
	var action ActionString = func() (string, error) { suite.countActions += 1; return "", errForTest }
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err1ForTests)
		return !ok
	}

	// Act
	actual, err := suite.retryPolicy.RetryActionString(action, handle)
	// Test
	suite.Equal(errForTest, err)
	suite.Equal("", actual)
	suite.Equal(1, suite.countActions)
}

func (suite *TestSuite) Test_RetryActionString_UnHandledError_ShouldBeExecutedFewTimes() {
	// Setup
	errForTest := &err1ForTests{}
	var action ActionString = func() (string, error) {
		suite.countActions += 1
		return "", errForTest
	}
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err2ForTests)
		return !ok
	}

	// Act
	actual, err := suite.retryPolicy.RetryActionString(action, handle)
	// Test
	suite.Equal(errForTest, errors.Cause(err))
	suite.Equal("", actual)
	suite.Equal(_configuration.RetryAttempts, suite.countActions)
}

func (suite *TestSuite) Test_RetryActionString_HandledErrorSecondTime_ShouldBeExecutedTwice() {
	// Setup
	errForTest := &err1ForTests{}
	err2ForTest := &err2ForTests{}
	var action ActionString = func() (string, error) {
		suite.countActions += 1
		if suite.countActions > 1 {
			return "", err2ForTest
		}
		return "", errForTest
	}
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		errType := &err2ForTests{}
		return !utils.IsErrorIsTypeOf(err, &errType)
	}

	// Act
	actual, err := suite.retryPolicy.RetryActionString(action, handle)
	// Test
	suite.Equal(err2ForTest, errors.Cause(err))
	suite.Equal("", actual)
	suite.Equal(2, suite.countActions)
}

func (suite *TestSuite) Test_RetryActionString_NoError_ShouldBeExecutedOnce() {
	// Setup
	var action ActionString = func() (string, error) { suite.countActions += 1; return "lior", nil }

	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err2ForTests)
		return !ok
	}

	// Act
	actual, err := suite.retryPolicy.RetryActionString(action, handle)
	// Test
	suite.Nil(err)
	suite.Equal("lior", actual)
	suite.Equal(1, suite.countActions)
}

func (suite *TestSuite) Test_RetryAction_NoErrorSecondTime_ShouldBeExecutedTwice() {
	// Setup
	errForTest := &err1ForTests{}
	var action ActionString = func() (string, error) {
		suite.countActions += 1
		if suite.countActions > 1 {
			return "lior", nil
		}
		return "", errForTest
	}
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err2ForTests)
		return !ok
	}

	// Act
	actual, err := suite.retryPolicy.RetryActionString(action, handle)
	// Test
	suite.Nil(err)
	suite.Equal("lior", actual)
	suite.Equal(2, suite.countActions)
}

func (suite *TestSuite) Test_RetryAction_ActionNil_ShouldReturnError() {
	// Setup
	var action Action = nil
	var handle ShouldRetryOnSpecificError = func(error) bool { suite.countActions += 1; return false }
	// Act
	err := suite.retryPolicy.RetryAction(action, handle)
	// Test
	suite.Equal(utils.NilArgumentError, errors.Cause(err))
	suite.Equal(0, suite.countActions)
}

func (suite *TestSuite) Test_RetryAction_HandleNil_ShouldReturnError() {
	// Setup

	var action Action = func() error { suite.countActions += 1; return nil }
	var handle ShouldRetryOnSpecificError = nil
	// Act
	err := suite.retryPolicy.RetryAction(action, handle)
	// Test
	suite.Equal(utils.NilArgumentError, errors.Cause(err))
	suite.Equal(0, suite.countActions)
}

func (suite *TestSuite) Test_RetryAction_HandledError_ShouldBeExecutedOnce() {
	// Setup
	errForTest := &err1ForTests{}
	var action Action = func() error { suite.countActions += 1; return errForTest }
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err1ForTests)
		return !ok
	}

	// Act
	err := suite.retryPolicy.RetryAction(action, handle)
	// Test
	suite.Equal(errForTest, err)
	suite.Equal(1, suite.countActions)
}

func (suite *TestSuite) Test_RetryAction_UnHandledError_ShouldBeExecutedFewTimes() {
	// Setup
	errForTest := &err1ForTests{}
	var action Action = func() error {
		suite.countActions += 1
		return errForTest
	}
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err2ForTests)
		return !ok
	}

	// Act
	err := suite.retryPolicy.RetryAction(action, handle)
	// Test
	suite.Equal(errForTest, errors.Cause(err))
	suite.Equal(_configuration.RetryAttempts, suite.countActions)
}

func (suite *TestSuite) Test_RetryAction_HandledErrorSecondTime_ShouldBeExecutedTwice() {
	// Setup
	errForTest := &err1ForTests{}
	err2ForTest := &err2ForTests{}
	var action Action = func() error {
		suite.countActions += 1
		if suite.countActions > 1 {
			return err2ForTest
		}
		return errForTest
	}
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err2ForTests)
		return !ok
	}

	// Act
	err := suite.retryPolicy.RetryAction(action, handle)
	// Test
	suite.Equal(err2ForTest, errors.Cause(err))
	suite.Equal(2, suite.countActions)
}

func (suite *TestSuite) Test_RetryAction_NoError_ShouldBeExecutedOnce() {
	// Setup
	var action Action = func() error { suite.countActions += 1; return nil }

	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err2ForTests)
		return !ok
	}

	// Act
	err := suite.retryPolicy.RetryAction(action, handle)
	// Test
	suite.Nil(err)
	suite.Equal(1, suite.countActions)
}

func (suite *TestSuite) Test_RetryActionString_NoErrorSecondTime_ShouldBeExecutedTwice() {
	// Setup
	errForTest := &err1ForTests{}
	var action Action = func() error {
		suite.countActions += 1
		if suite.countActions > 1 {
			return nil
		}
		return errForTest
	}
	var handle ShouldRetryOnSpecificError = func(err error) bool {
		_, ok := err.(*err2ForTests)
		return !ok
	}

	// Act
	err := suite.retryPolicy.RetryAction(action, handle)
	// Test
	suite.Nil(err)
	suite.Equal(2, suite.countActions)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestRetryPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
