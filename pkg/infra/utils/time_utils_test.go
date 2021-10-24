package utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	_configuration3Sec = TimeoutConfiguration{TimeDuration: 5, TimeUnit: "s"}
	_duration3Sec      = 3 * time.Second
)

type TimeUtilsTestSuite struct {
	suite.Suite
}

func (suite *TimeUtilsTestSuite) Test_ParseTimeoutConfigurationToDurationOrDefault_validConfiguration_shouldNotUseDefault() {
	// Setup
	expected := 5 * time.Second
	// Act
	actual := ParseTimeoutConfigurationToDurationOrDefault(&_configuration3Sec, _duration3Sec)

	// Test
	suite.Exactly(expected, actual)
}

func (suite *TimeUtilsTestSuite) Test_ParseTimeoutConfigurationToDurationOrDefault_nilConfiguration_shouldUseDefault() {
	// Act
	actual := ParseTimeoutConfigurationToDurationOrDefault(nil, _duration3Sec)

	// Test
	suite.Exactly(_duration3Sec, actual)
}

func (suite *TimeUtilsTestSuite) Test_ParseTimeoutConfigurationToDurationOrDefault_invalidConfiguration_shouldUseDefault() {
	invalidConfiguration := TimeoutConfiguration{TimeDuration: 5, TimeUnit: "lior"}
	// Act
	actual := ParseTimeoutConfigurationToDurationOrDefault(&invalidConfiguration, _duration3Sec)

	// Test
	suite.Exactly(_duration3Sec, actual)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestTimeUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(TimeUtilsTestSuite))
}
