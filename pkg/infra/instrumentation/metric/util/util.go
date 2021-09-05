package util

import "time"

// GetDurationMilliseconds returns the duration between startTime and currentTime in milliseconds as int.
func GetDurationMilliseconds(startTime time.Time) int {
	endTime := time.Now().UTC()
	return int(startTime.Sub(endTime).Milliseconds())
}
