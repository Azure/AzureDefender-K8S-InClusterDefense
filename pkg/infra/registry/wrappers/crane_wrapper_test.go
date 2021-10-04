package wrappers

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"regexp"
	"strconv"
	"testing"
	"time"
)

const (
	_retryAttempts int = 3
	_retryDuration     = 10
	_timeUnit          = "ms"
)

var (
	retryPolicyConfiguration = &utils.RetryPolicyConfiguration{
		RetryAttempts: _retryAttempts,
		RetryDuration: _retryDuration,
		TimeUnit:      _timeUnit,
	}
)

type TestSuite struct {
	suite.Suite
	craneWrapper *CraneWrapper
}

// This will run before each test in the suit
func (suite *TestSuite) SetupTest() {
	suite.craneWrapper = NewCraneWrapper(retryPolicyConfiguration)
}

// Test the amount of actual retries is equal _retryAttempts (by a linear factor)
// TODO once Digest method does not return a static digest, test an image tag which will fail to verify number of attempts
func (suite *TestSuite) TestCraneWrapper_NumberOfAttempts() {
	re, err := regexp.Compile("[0-9]+") // error if regexp invalid
	// Verify regex hasnt failed to compile
	if err != nil {
		suite.Fail("failed to compile regex")
	}
	// TODO remove skip and update test according to new DigestWithRetry method
	suite.T().Skip()
	//_, err = suite.craneWrapper.DigestWithRetry("")
	// number of attempts is tested only if DigestWithRetry has failed
	suite.NotNil(err, "Digest hasn't failed")
	// Extract number of actual attempts
	numberOfRetries, err := strconv.Atoi(re.FindString(err.Error()))
	// Fail the test in case extracted number of attempts can't be converted to int
	if err != nil {
		suite.Fail("Failed to convert extracted number of Attempts to int")
	}
	assert.Equal(suite.T(), numberOfRetries, _retryAttempts)
}

// Test that the sleep duration between each retry is getting bigger (by a linear factor)
// TODO once Digest method does not return a static digest, test an image tag which will fail to verify Back off
func (suite *TestSuite) TestCraneWrapper_RetriesBackOff() {
	startTime := time.Now()
	for i := 0; i < _retryAttempts; i++ {
		// TODO change empty string to a failing image tag
		suite.craneWrapper.Digest("")
		time.Sleep(_retryDuration)
	}
	// Calculate running time for static delay
	constDurationTime := util.GetDurationMilliseconds(startTime)
	startTime = time.Now()
	// TODO change empty string to a failing image tag
	// TODO remove skip and update test according to new DigestWithRetry method
	suite.T().Skip()
	// suite.craneWrapper.DigestWithRetry("")
	// Calculate running time for increasing delay
	increasingDurationTime := util.GetDurationMilliseconds(startTime)
	// TODO from > to <
	assert.True(suite.T(), constDurationTime > increasingDurationTime, "retries back off delay is not increasing")
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
