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

// GetSizeInBytes returns the size of value in bytes.
func GetSizeInBytes(value string) int {
	// The reason for using len is exlpained in the following link:  https://www.golangprograms.com/golang-get-number-of-bytes-and-runes-in-a-string.html
	// In GoLang Strings are UTF-8 encoded, this means each character called rune can be of 1 to 4 bytes long.
	// So len returns the num of bytes that string takes.
	return len(value)
}
