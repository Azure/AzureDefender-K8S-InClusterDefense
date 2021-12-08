package utils

// ValidatePositiveInt gets an unknown number of variables of type int and return an false in case at least one of them is non-positive
// Otherwise return true
func ValidatePositiveInt(variables... int) bool{
	for _, variable := range variables{
		if variable <= 0 {
			return false
		}
	}
	// All variables are positive
	return true
}
