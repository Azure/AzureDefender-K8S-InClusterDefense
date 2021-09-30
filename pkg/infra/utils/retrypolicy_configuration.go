package utils

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
