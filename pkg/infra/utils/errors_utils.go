package utils

import (
	"github.com/pkg/errors"
)

var (
	NilArgumentError = errors.New("NilArgumentError")
	ReadFromClosedChannelError = errors.New("ReadFromClosedChannelError")
	CantConvertChannelDataWrapper = errors.New("Cant Convert Channel Data Wrapper to the designated type")
)
