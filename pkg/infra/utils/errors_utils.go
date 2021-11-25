package utils

import (
	"github.com/pkg/errors"
)

var (
	NilArgumentError              = errors.New("NilArgumentError")
	ReadFromClosedChannelError    = errors.New("ReadFromClosedChannelError")
	CantConvertChannelDataWrapper = errors.New("Cant Convert Channel Data Wrapper to the designated type")
	SetKeyInCacheWithNoExpiration = errors.New("SetKeyInCacheWithNoExpiration")
)

// IsErrorIsTypeOf gets an error and type of interface and returns true or false if the error caused by some error of type t.
// NOTICE - you should pass t as double pointer - e.g.
// err := &os.PathError{}
// err2 := &os.PathError{}
// IsErrorIsTypeOf(err, &err2)
func IsErrorIsTypeOf(err error, t interface{}) bool {
	return errors.As(err, t)
}
