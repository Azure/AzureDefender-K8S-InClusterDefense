package trace

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
	"strings"
)

// Tracer implementation of ITracer interface - holds an Entry object in order to delegate Tivan tracer methods.
type Tracer struct {
	// Entry is logrus.Entry that Tivan use in their logger so we wrap it.
	Entry *logrus.Entry
	//context indicates the context of the log.
	context string
	//level is the tracer level (e.g. DEBUG,INFO,WARN, etc.)
	level zapcore.Level
	// Encoder is the fomrat that the msg will be written (e.g. json,tuple, etc.)
	encoder Encoder
}

// Info writes msg. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Info(msg string, keysAndValues ...interface{}) {
	// Check that the level of the tracer is enabled, and the len of keysAndValues are even:
	if !tracer.level.Enabled(zapcore.InfoLevel) || len(keysAndValues)%2 == 1 {
		return
	}

	msgWithContext := tracer.concatenateContextToMsg(msg)
	if keysAndValues != nil && len(keysAndValues) > 0 {
		tracer.Entry.Error(msgWithContext, keysAndValues)
	} else {
		tracer.Entry.Error(msgWithContext)
	}
}

// Error writes error. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Error(err error, msg string, keysAndValues ...interface{}) {
	// Check that the level of the tracer is enabled, and the len of keysAndValues are even:
	if !tracer.level.Enabled(zapcore.ErrorLevel) || len(keysAndValues)%2 == 1 {
		return
	}

	msgWithContext := tracer.concatenateContextToMsg(msg)
	if keysAndValues != nil && len(keysAndValues) > 0 {
		tracer.Entry.Error(msgWithContext, err, keysAndValues)
	} else {
		tracer.Entry.Error(msgWithContext, err)
	}
}

// V implements this method only for logr.Logger interface. doesn't do anything - return the same tracer!
func (tracer Tracer) V(level int) logr.Logger {
	return &Tracer{
		level:   tracer.level - zapcore.Level(level),
		context: tracer.context,
		Entry:   tracer.Entry,
	}
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append suffixes to the logger's name.
func (tracer Tracer) WithName(name string) logr.Logger {
	newTracer := tracer.addNewContext(name)
	return newTracer
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (tracer *Tracer) addNewContext(suffix string) *Tracer {
	if suffix == "" {
		return tracer
	}
	t := tracer.clone()
	// In case that context is empty, don't add . as separator
	if tracer.context == "" {
		t.context = suffix
		// concatenate the suffix to the current context.
	} else {
		t.context = strings.Join([]string{t.context, suffix}, ".")
	}
	return t
}

// clone the tracer.
func (tracer *Tracer) clone() *Tracer {
	copiedTracer := *tracer
	return &copiedTracer
}

// concatenateContextToMsg is concatenating the trace to the msg with specific encoding.
func (tracer *Tracer) concatenateContextToMsg(msg string) (newMsg string) {
	switch tracer.encoder {
	case JSON:
		return tracer.concatTraceToMsgJson(msg)
	case NONE:
		fallthrough
	default:
		return tracer.concatTraceToMsgNone(msg)
	}
}

// concatTraceToMsgNone is concatenating the trace to the msg as tuple without encoding:
// e.g. tracer.trace = "server.tag2digest", msg = "digest was resolved" -->
// the return value will be : "trace: server.tag2digest, msg: digest was resolved"
func (tracer *Tracer) concatTraceToMsgNone(msg string) (newMsg string) {
	return fmt.Sprintf(`context: %s,msg: %s`, tracer.context, msg)
}

// concatTraceToMsgJson is concatenating the trace to the msg as json:
// e.g. tracer.trace = "server.tag2digest", msg = "digest was resolved" -->
// the return value will be : "{"trace": "server.tag2digest", msg: "digest was resolved"}
func (tracer *Tracer) concatTraceToMsgJson(msg string) (newMsg string) {
	return fmt.Sprintf(`{"context":"%s","msg":"%s"}`, tracer.context, msg)
}

// WithValues implements this method only for logr.Logger interface. doesn't do anything - return the same tracer!
func (tracer Tracer) WithValues(keysAndValues ...interface{}) logr.Logger {
	return tracer
}

// Enabled tests whether this Logger is enabled.
//delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer Tracer) Enabled() bool {
	return tracer.Entry != nil
}
