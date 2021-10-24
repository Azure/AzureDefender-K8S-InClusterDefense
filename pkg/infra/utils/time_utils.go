package utils

import (
	"strconv"
	"time"
)

type TimeoutConfiguration struct {
	TimeDuration int
	TimeUnit     string
}

// ParseTimeoutConfigurationToDurationOrDefault gets TimeoutConfiguration and default time duration. it tries to parse
// TimeoutConfiguration into Duration. if some error encountered while parsing the timeout, returns the default duration
func ParseTimeoutConfigurationToDurationOrDefault(configuration *TimeoutConfiguration, defaultDuration time.Duration) time.Duration {
	if configuration == nil {
		return defaultDuration
	}

	duration, err := time.ParseDuration(strconv.Itoa(configuration.TimeDuration) + configuration.TimeUnit)
	if err != nil {
		return defaultDuration
	}

	return duration
}
