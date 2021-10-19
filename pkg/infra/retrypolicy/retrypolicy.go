package retrypolicy

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

// ShouldRetryOnSpecificError is function that gets an error and returns true or false if the retry should handle with this error
// Returns true in case that retry policy doesn't know how to handle with some error, so it will retry another time (according to the retry attempts and retry count)
// Returns false in case that the retry policy shouldn't retry another try on error specific error.
type ShouldRetryOnSpecificError func(error) bool

// All actions functions should be written in this type:
type (
	// ActionString is action function that returns string and error
	ActionString func() (string, error)
	// Action is action function that returns error
	Action func() error
)

// IRetryPolicy interface for retrypolicy
type IRetryPolicy interface {
	// RetryActionString try to execute action that returns string,error
	RetryActionString(action ActionString, handle ShouldRetryOnSpecificError) (value string, err error)
	// RetryAction try to execute action that returns only error
	RetryAction(action Action, handle ShouldRetryOnSpecificError) (err error)
}

// RetryPolicy implements IRetryPolicy interface
var _ IRetryPolicy = (*RetryPolicy)(nil)

// RetryPolicy manages the retry policy of functions.
type RetryPolicy struct {
	//tracerProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter
	metricSubmitter metric.IMetricSubmitter
	// duration of the retry policy.
	duration time.Duration
	// RetryAttempts  is the number of attempts that the request should be executed.
	retryAttempts int
}

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

// NewRetryPolicy Cto'r for retry policy object
func NewRetryPolicy(instrumentationProvider instrumentation.IInstrumentationProvider, configuration *RetryPolicyConfiguration) (*RetryPolicy, error) {
	duration, err := GetBackOffDuration(configuration)
	if err != nil {
		return nil, err
	}
	if configuration.RetryAttempts <= 0 {
		return nil, errors.New("RetryAttempts must be integer > 0")
	}

	retryPolicy := &RetryPolicy{
		duration:        duration,
		retryAttempts:   configuration.RetryAttempts,
		tracerProvider:  instrumentationProvider.GetTracerProvider("RetryPolicy"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}

	return retryPolicy, nil
}

// RetryActionString retry to run the action with retryPolicy
func (r *RetryPolicy) RetryActionString(action ActionString, shouldRetry ShouldRetryOnSpecificError) (value string, err error) {
	tracer := r.tracerProvider.GetTracer("RetryActionString")
	if action == nil || shouldRetry == nil {
		err = errors.Wrap(utils.NilArgumentError, "action and handle can't be nil")
		tracer.Error(err, "")
		return "", err
	}

	// Update retryCount to 1
	retryCount := 1
	for retryCount <= r.retryAttempts {
		// Act
		value, err = action()

		if err != nil && !shouldRetry(err) { // Check if shouldRetry knows how to handle with error.
			tracer.Info("failed but encountered with handled err", "err", err)
			return "", err

		} else if err == nil { // Check if err is nil - returns the result.
			tracer.Info("succeed", "retryCount", retryCount, "value", value)
			return value, nil
		}

		// in case that err != nil and handle(err) is false, should try another execution.
		retryCount += 1
		tracer.Info("waiting for another retry", "sleepTime", time.Duration(retryCount)*r.duration)
		time.Sleep(time.Duration(retryCount) * r.duration)

	}
	err = errors.Wrapf(err, "failed after %d tries", retryCount)
	tracer.Error(err, "")
	return "", err
}

// RetryAction retry to run the action with retryPolicy
func (r *RetryPolicy) RetryAction(action Action, shouldRetry ShouldRetryOnSpecificError) (err error) {
	tracer := r.tracerProvider.GetTracer("RetryActionString")
	if action == nil || shouldRetry == nil {
		err = errors.Wrap(utils.NilArgumentError, "action and handle can't be nil")
		tracer.Error(err, "")
		return err
	}

	// Update retryCount to 1
	retryCount := 1
	for retryCount <= r.retryAttempts {
		// Act
		err = action()

		if err != nil && !shouldRetry(err) { // Check if handle knows how to handle with error.
			tracer.Info("failed but encountered with handled err", "err", err)
			return err

		} else if err == nil { // Check if err is nil - returns the result.
			tracer.Info("succeed", "retryCount", retryCount)
			return nil
		}

		// in case that err != nil and handle(err) is false, should try another execution.
		retryCount += 1
		tracer.Info("waiting for another retry", "sleepTime", time.Duration(retryCount)*r.duration)
		time.Sleep(time.Duration(retryCount) * r.duration)

	}
	err = errors.Wrapf(err, "failed after %d tries", retryCount)
	tracer.Error(err, "")
	return err
}

// GetBackOffDuration uses the RetryPolicyConfiguration instance's RetryDuration (int) and TimeUnit(string)
// to a return a time.Duration object of the backoff duration
func GetBackOffDuration(configuration *RetryPolicyConfiguration) (duration time.Duration, err error) {
	if configuration == nil {
		return 0, errors.Wrap(utils.NilArgumentError, "_configuration can't be nil")
	} else if configuration.RetryDuration <= 0 {
		return 0, errors.New("RetryDuration must be > 0")
	}

	return time.ParseDuration(strconv.Itoa(configuration.RetryDuration) + configuration.TimeUnit)
}
