package utils


type PositiveIntValidationObject struct{
	VariableName string
	Variable int
}

// ValidatePositiveInt gets an unknown number of variables of type int and return an false in case at least one of them is non-positive
// Otherwise return true
func ValidatePositiveInt(positiveIntValidationObjects... *PositiveIntValidationObject) (bool, string){
	for _, positiveIntValidationObject := range positiveIntValidationObjects{
		if positiveIntValidationObject.Variable <= 0 {
			return false, positiveIntValidationObject.VariableName
		}
	}
	// All variables are positive
	return true, ""
}
