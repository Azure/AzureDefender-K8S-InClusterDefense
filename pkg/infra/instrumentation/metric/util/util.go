package util

import "time"

// GetDurationMilliseconds returns the duration between startTime and currentTime in milliseconds as int.
// Example calling:
// startTime := Time.Now().UTC()
// util.GetDurationMilliseconds(startTime)
func GetDurationMilliseconds(startTime time.Time) int {
	endTime := time.Now().UTC()
	return int(startTime.UTC().Sub(endTime).Milliseconds())
}
