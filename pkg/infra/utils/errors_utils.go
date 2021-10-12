package utils

import (
	"github.com/pkg/errors"
)

var (
	NilArgumentError = errors.New("NilArgumentError")
	ReadFromClosedChannelError = errors.New("ReadFromClosedChannelError")
)
