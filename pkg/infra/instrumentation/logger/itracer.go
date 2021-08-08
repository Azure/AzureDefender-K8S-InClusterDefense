package instrumenation

import "github.com/go-logr/logr"

// ITracer interface of tracer that extends the logr.Logger interface
type ITracer interface {
	logr.Logger // Extend the logr.Logger interface
}
