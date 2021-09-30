package utils

import (
	"fmt"
	"strings"
)

// GetType returns the full type of the object. it contains the type of the object with the packages as prefixes.
// e.g.  x: = main.User{} -- GetTypeWithoutPackage(x) -> main.User
func GetType(object interface{}) string {
	// Get the type as string.
	return fmt.Sprintf("%T", object)
}

// GetTypeWithoutPackage returns the type of the object without the package that the object belongs to.
// e.g.  x: = main.User{} -- GetTypeWithoutPackage(x) -> User
func GetTypeWithoutPackage(object interface{}) string {
	fullObjectType := GetType(object)
	// In case that the type contains prefix
	arr := strings.SplitAfter(fullObjectType, ".")
	objectType := arr[len(arr)-1]
	return objectType
}
