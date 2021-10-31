package utils

import (
	"time"
)

type TimeoutConfiguration struct {
	TimeDurationInMS int
}

// ParseTimeoutConfigurationToDuration gets TimeoutConfiguration parse the timeDurationInMS of TimeoutConfiguration and returns it as duration.
func (configuration *TimeoutConfiguration) ParseTimeoutConfigurationToDuration() time.Duration {
	return time.Duration(configuration.TimeDurationInMS) * time.Millisecond
}
