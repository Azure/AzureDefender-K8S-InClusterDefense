package trace

import "github.com/go-logr/logr"

// ITracer interface of tracer that extends the logr.Logger interface.
type ITracer interface {
	// Logger Extend the logr.Logger interface
	logr.Logger
}
