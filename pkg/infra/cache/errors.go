package cache

import (
	"fmt"
	"time"
)

var _ error = (*MissingKeyCacheError)(nil)

// MissingKeyCacheError is error that represent that is returned when ICacheClient is missing key on get method.
type MissingKeyCacheError struct {
	key string
}

func NewMissingKeyCacheError(key string) error {
	return &MissingKeyCacheError{key: key}
}
func (err *MissingKeyCacheError) Error() string {
	msg := fmt.Sprintf("Key <%v> is missing.", err.key)
	return msg
}

var _ error = (*NegativeExpirationCacheError)(nil)

// NegativeExpirationCacheError abstract class that implements negative expiration error
type NegativeExpirationCacheError struct {
	expiration time.Duration
}

func NewNegativeExpirationCacheError(expiration time.Duration) error {
	return &NegativeExpirationCacheError{expiration: expiration}
}

func (err *NegativeExpirationCacheError) Error() string {
	msg := fmt.Sprintf("Invalid expiration, expiration should be non-negative. got <%v>", err.expiration)
	return msg
}
