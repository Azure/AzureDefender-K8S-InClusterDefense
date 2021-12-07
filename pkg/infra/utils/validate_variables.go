package utils

import "github.com/pkg/errors"

// ValidatePositiveInt gets an unknown number of variables of type int and return an error in case at least one of them is non-positive
func ValidatePositiveInt(variables... int) error{
	for _, variable := range variables{
		if variable <= 0 {
			err := errors.Wrapf(InvalidConfiguration, "Got non-positive int %v", variable)
			return err
		}
	}
	// All variables are positive
	return nil
}
