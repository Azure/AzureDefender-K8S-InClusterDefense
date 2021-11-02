package utils

import "github.com/pkg/errors"

// ChannelDataWrapper is a struct that hold 2 members: data and error.
// Its main purpose is to enable sending and receiving generic data throw channels.
type ChannelDataWrapper struct {
	// data is a generic data object to be sent and received
	data interface{}
	// err is an error occurred while getting the data
	err error
}

// NewChannelDataWrapper - NewChannelDataWrapper Ctor
func NewChannelDataWrapper(data interface{}, err error) *ChannelDataWrapper {
	return &ChannelDataWrapper{
		data: data,
		err:  err,
	}
}

// GetData gets the adta from the ChannelDataWrapper.
// returns error (that is not nil) in case that wrapper got non-nil err in the constructor or in case that the data is nil
func (wrapper *ChannelDataWrapper) GetData() (interface{}, error) {
	if wrapper.err != nil {
		return nil, wrapper.err
	} else if wrapper.data == nil {
		return nil, errors.Wrap(NilArgumentError, "data is nil")
	}
	return wrapper.data, nil
}
