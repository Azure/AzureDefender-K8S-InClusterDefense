package cache

import (
	"fmt"
	"time"
)

var _ error = (*NegativeExpirationError)(nil)

// NegativeExpirationError abstarct class that implements negative expiration error
type NegativeExpirationError struct {
	expiration time.Duration
}

func NewNegativeExpirationError(expiration time.Duration) error {
	return &NegativeExpirationError{expiration: expiration}
}

func (err *NegativeExpirationError) Error() string {
	msg := fmt.Sprintf("Invalid expiration, expiration should be non-negative. got <%v>", err.expiration)
	return msg
}
