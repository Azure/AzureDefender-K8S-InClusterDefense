package utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	_configuration3Sec = TimeoutConfiguration{TimeDurationInMS: 5}
)

type TimeUtilsTestSuite struct {
	suite.Suite
}

func (suite *TimeUtilsTestSuite) Test_ParseTimeoutConfigurationToDuration_validConfiguration_shouldNotUseDefault() {
	// Setup
	expected := 5 * time.Millisecond
	// Act
	actual := _configuration3Sec.ParseTimeoutConfigurationToDuration()

	// Test
	suite.Exactly(expected, actual)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestTimeUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(TimeUtilsTestSuite))
}
