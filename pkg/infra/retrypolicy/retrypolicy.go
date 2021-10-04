package retrypolicy

import (
	"github.com/pkg/errors"
	"strconv"
	"time"
)

// Handle is function that gets an error and returns true or false if the retry should handle with this error
// Returns true in case that the retry policy shouldn't retry another try on errors.
// Returns false in case that retry policy doesn't know how to handle with some error, so it will retry another time (according to the retry attempts and retry count)
type Handle func(error) bool

// All actions functions should be written in this type:
type (
	// ActionString is action function that returns string and error
	ActionString func() (string, error)
)

// RetryPolicy is the retry policy configuration that holds the relevant fields for executing retry policy.
type RetryPolicy struct {
	// RetryAttempts  is the number of attempts that the request should be executed.
	RetryAttempts int
	// RetryDuration is the time duration between each retry - it is represented as string
	RetryDuration int
	// TimeUnit is the unit of time for the backoff duration
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	TimeUnit string
}

// GetBackOffDuration uses the RetryPolicy instance's RetryDuration (int) and TimeUnit(string)
// to a return a time.Duration object of the backoff duration
func (r *RetryPolicy) GetBackOffDuration() (duration time.Duration, err error) {
	return time.ParseDuration(strconv.Itoa(r.RetryDuration) + r.TimeUnit)
}

// RetryActionString retry to run the action with retryPolicy
func (r *RetryPolicy) RetryActionString(action ActionString, handle Handle) (string, error) {
	// TODO Create tests for this method
	retryCount := 0

	retryDuration, err := r.GetBackOffDuration()
	if err != nil {
		return "", errors.Wrapf(err, "cannot parse given retry duration <(%v)>", r.RetryDuration)
	}

	// Update retryCount to 1
	retryCount = 1
	for retryCount <= r.RetryAttempts {
		// Act
		value, err := action()

		if err != nil && handle(err) { // Check if handle knows how to handle with error.
			return "", err
		} else if err == nil { // Check if err is nil - returns the result.
			return value, nil
		} else { // in case that err != nil and handle(err) is false, should try another execution.
			retryCount += 1
			// wait (retryCount * craneWrapper.retryDuration) milliseconds between retries
			time.Sleep(time.Duration(retryCount) * retryDuration)
		}

	}
	return "", err
}
