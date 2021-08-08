package instrumenation

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
	"strings"
)

// Tracer implementation of ITracer interface - holds a Entry object in order to delegate Tivan tracer methods.
type Tracer struct {
	Entry *logrus.Entry
	trace string
	level zapcore.Level
}

// Enabled tests whether this Logger is enabled.
//delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Enabled() bool {
	return tracer.Entry != nil
}

// Info writes msg. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Info(msg string, keysAndValues ...interface{}) {
	// Check that the level of the tracer is enabled, and the len of keysAndValues are even:
	if !tracer.level.Enabled(zapcore.InfoLevel) || len(keysAndValues)%2 == 1 {
		return
	}

	msgWithTrace := tracer.concatTraceToMsg(msg)
	if keysAndValues != nil && len(keysAndValues) > 0 {
		tracer.Entry.Error(msgWithTrace, keysAndValues)
	} else {
		tracer.Entry.Error(msgWithTrace)
	}
}

// Error writes error. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Error(err error, msg string, keysAndValues ...interface{}) {
	// Check that the level of the tracer is enabled, and the len of keysAndValues are even:
	if !tracer.level.Enabled(zapcore.ErrorLevel) || len(keysAndValues)%2 == 1 {
		return
	}

	msgWithTrace := tracer.concatTraceToMsg(msg)
	if keysAndValues != nil && len(keysAndValues) > 0 {
		tracer.Entry.Error(msgWithTrace, keysAndValues)
	} else {
		tracer.Entry.Error(msgWithTrace)
	}
}

// V implements this method only for logr.Logger interface. doesn't do anything - return the same tracer!
func (tracer Tracer) V(level int) logr.Logger {
	return &Tracer{
		level: tracer.level - zapcore.Level(level),
		trace: tracer.trace,
		Entry: tracer.Entry,
	}
}

// WithValues implements this method only for logr.Logger interface. doesn't do anything - return the same tracer!
func (tracer Tracer) WithValues(keysAndValues ...interface{}) logr.Logger {
	return tracer
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append suffixes to the logger's name.
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

// clone the tracer.
func (tracer *Tracer) clone() *Tracer {
	copiedTracer := *tracer
	return &copiedTracer
}

// concatTraceToMsg is concatenating the trace to the msg as json:
// e.g. tracer.trace = "server.tag2digest", msg = "digest was resolved" -->
// the return value will be : "{"trace": "server.tag2digest", msg: "digest was resolved"}
func (tracer *Tracer) concatTraceToMsg(msg string) (newMsg string) {
	return fmt.Sprintf(`{"trace":"%s","msg":"%s"}`, tracer.trace, msg)
}
