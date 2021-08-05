package instrumentation

import (
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// ITracer interface of tracer that extends the logr.Logger interface
type ITracer struct {
	logr.Logger // Extend the logr.Logger interface
}

// Tracer implementation of ITracer interface - holds a Entry object in order to delegate Tivan tracer methods.
type Tracer struct {
	Entry *logrus.Entry
}

// Enabled tests whether this Logger is enabled.
//delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Enabled() bool {
	return tracer.Entry != nil
}

// Info writes msg. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Info(msg string, keysAndValues ...interface{}) {
	tracer.Entry.Info(msg, keysAndValues)
}

// Error writes error. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Error(err error, msg string, keysAndValues ...interface{}) {
	tracer.Entry.Error(err, msg, keysAndValues)
}

// V implements this method only for logr.Logger interface. doesn't do anything - return the same tracer!
func (tracer Tracer) V(level int) logr.Logger {
	return tracer //TODO ??
}

// WithValues implements this method only for logr.Logger interface. doesn't do anything - return the same tracer!
func (tracer Tracer) WithValues(keysAndValues ...interface{}) logr.Logger {
	return tracer
}

// WithName implements this method only for logr.Logger interface. doesn't do anything - return the same tracer!
func (tracer Tracer) WithName(name string) logr.Logger {
	return tracer
}
