package utils

import (
	"github.com/stretchr/testify/suite"
	"math"
	"testing"
	"time"
)

var (
	_configuration3Sec = TimeoutConfiguration{TimeDurationInMS: 5}
	_nanoToMS          = 1000000
	_nanoToSecond      = 1000000000
	_nanoToMinute      = 6 * math.Pow10(10)
	_nanoToHour        = 3.6 * math.Pow10(12)
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

func (suite *TimeUtilsTestSuite) Test_GetMilliseconds() {
	result := GetMilliseconds(1)
	expectedResult := time.Duration(_nanoToMS)
	suite.Equal(result, expectedResult)

}

func (suite *TimeUtilsTestSuite) Test_GetSeconds() {
	result := GetSeconds(1)
	expectedResult := time.Duration(_nanoToSecond)
	suite.Equal(result, expectedResult)

}

func (suite *TimeUtilsTestSuite) Test_GetMinutes() {
	result := GetMinutes(1)
	expectedResult := time.Duration(_nanoToMinute)
	suite.Equal(result, expectedResult)

}

func (suite *TimeUtilsTestSuite) Test_GetHours() {
	result := GetHours(1)
	expectedResult := time.Duration(_nanoToHour)
	suite.Equal(result, expectedResult)
}

func (suite *TimeUtilsTestSuite) Test_Repeat() {
	previousTime := time.Now()
	called := false
	calledInIf := false
	RepeatEveryTick(time.Millisecond*2, func() error {
		if called {
			diff := time.Now().Sub(previousTime)
			suite.GreaterOrEqual(diff, time.Millisecond)
			calledInIf = true
		}
		previousTime = time.Now()
		called = true
		return nil
	})
	time.Sleep(time.Millisecond * 50)
	suite.True(called)
	suite.True(calledInIf)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestTimeUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(TimeUtilsTestSuite))
}
