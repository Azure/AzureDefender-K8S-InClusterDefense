package utils

// ChannelDataWrapper is a struct that hold 2 members: data and error.
// Its main purpose is to enable sending and receiving generic data throw channels.
type ChannelDataWrapper struct {
	// DataWrapper is a generic data object to be sent and received
	DataWrapper interface{}
	// Err is an error occurred while getting the data
	Err error
}

// NewChannelDataWrapper - NewChannelDataWrapper Ctor
func NewChannelDataWrapper (data interface{}, err error) *ChannelDataWrapper{
	return &ChannelDataWrapper{
		DataWrapper: data,
		Err: err,
	}
}
