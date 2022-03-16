package utils

// StringInSlice return true if list contain str,false otherwise.
func StringInSlice(str string, list []string) bool {
	for _, listValue := range list {
		if listValue == str {
			return true
		}
	}
	return false
}
