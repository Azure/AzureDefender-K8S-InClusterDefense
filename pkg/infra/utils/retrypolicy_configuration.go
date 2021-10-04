package utils

import (
	"strconv"
	"time"
)

// RetryPolicyConfiguration is the retry policy configuration that holds the relevant fields for executing retry policy.
type RetryPolicyConfiguration struct {
	// RetryAttempts  is the number of attempts that the request should be executed.
	RetryAttempts int
	// RetryDuration is the time duration between each retry - it is represented as string
	RetryDuration int
	// TimeUnit is the unit of time for the backoff duration
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	TimeUnit string
}

// GetBackOffDuration uses the RetryPolicyConfiguration instance's RetryDuration (int) and TimeUnit(string)
// to a return a time.Duration object of the backoff duration
func (retryPolicyConfiguration *RetryPolicyConfiguration) GetBackOffDuration() (duration time.Duration, err error){
	return time.ParseDuration(strconv.Itoa(retryPolicyConfiguration.RetryDuration) + retryPolicyConfiguration.TimeUnit)
}
