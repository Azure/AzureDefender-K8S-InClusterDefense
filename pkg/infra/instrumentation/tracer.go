package instrumentation

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"strings"
)

// ITracer interface of tracer that extends the logr.Logger interface
type ITracer struct {
	logr.Logger // Extend the logr.Logger interface
}

// Tracer implementation of ITracer interface - holds a Entry object in order to delegate Tivan tracer methods.
type Tracer struct {
	Entry *logrus.Entry
	trace string
}

// Enabled tests whether this Logger is enabled.
//delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Enabled() bool {
	return tracer.Entry != nil
}

type X struct {
	trace string
	msg   string
}

// Info writes msg. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Info(msg string, keysAndValues ...interface{}) {
	newMsg := tracer.constructMsg(msg)
	tracer.Entry.Info(newMsg)
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
	newTracer := tracer.Named(name)
	return newTracer
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (tracer *Tracer) Named(s string) *Tracer {
	if s == "" {
		return tracer
	}
	t := tracer.clone()
	if tracer.trace == "" {
		t.trace = s
	} else {
		t.trace = strings.Join([]string{t.trace, s}, ".")
	}
	return t
}

func (tracer *Tracer) clone() *Tracer {
	copiedTracer := *tracer
	return &copiedTracer
}

func (tracer *Tracer) constructMsg(msg string) (newMsg string) {
	return fmt.Sprintf(`{"trace":"%s","msg":"%s"}`, tracer.trace, msg)
}
