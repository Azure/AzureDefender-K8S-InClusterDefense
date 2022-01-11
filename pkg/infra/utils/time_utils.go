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

// GetMilliseconds receive durationInMilliseconds as int and return it as time.Duration in milliseconds.
func GetMilliseconds(durationInMilliseconds int) time.Duration {
	return time.Duration(durationInMilliseconds) * time.Millisecond
}

// GetSeconds receive durationInSeconds as int and return it as time.Duration in seconds.
func GetSeconds(durationInSeconds int) time.Duration {
	return time.Duration(durationInSeconds) * time.Second
}

// GetMinutes receive durationInMinutes as int and return it as time.Duration in minutes.
func GetMinutes(durationInMinutes int) time.Duration {
	return time.Duration(durationInMinutes) * time.Minute
}

// GetHours receive durationInHours as int and return it as time.Duration in hours.
func GetHours(durationInHours int) time.Duration {
	return time.Duration(durationInHours) * time.Hour
}

// RepeatEveryTick calls a given function repeatedly every duration.
func RepeatEveryTick(duration time.Duration, f func() error){
	ticker := time.NewTicker(duration)
	go func() {
		for range ticker.C{
			f()
		}
	}()
}