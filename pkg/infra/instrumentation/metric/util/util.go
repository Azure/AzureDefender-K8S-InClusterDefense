package util

import "time"

// GetDurationMilliseconds returns the duration between startTime and endTime in milliseconds as int.
func GetDurationMilliseconds(startTime time.Time, endTime time.Time) int {
	return int(startTime.Sub(endTime).Milliseconds())
}
