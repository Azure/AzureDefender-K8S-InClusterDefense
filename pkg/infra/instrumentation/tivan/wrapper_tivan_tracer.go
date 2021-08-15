package tivan

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
	"strings"
)

// WrapperTivanTracer implementation of ITracer interface - holds an Entry object in order to delegate Tivan tracer methods.
type WrapperTivanTracer struct {
	// entry is logrus.Entry that Tivan use in their logger so we wrap it.
	entry *logrus.Entry
	//context indicates the context of the log.
	context string
	//level is the tracer level (e.g. DEBUG,INFO,WARN, etc.)
	level zapcore.Level
	// Encoder is the fomrat that the msg will be written (e.g. json,tuple, etc.)
	encoder trace.Encoder
}

// NewWrapperTivanTracer Ctor of tracer
func NewWrapperTivanTracer(entry *logrus.Entry, context string, level zapcore.Level, encoder trace.Encoder) (tracer *WrapperTivanTracer) {
	return &WrapperTivanTracer{
		entry:   entry,
		context: context,
		level:   level,
		encoder: encoder,
	}
}

// Info writes msg. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer *WrapperTivanTracer) Info(msg string, keysAndValues ...interface{}) {
	// Check that the level of the tracer is enabled, and the len of keysAndValues are even:
	if !tracer.level.Enabled(zapcore.InfoLevel) {
		return
	}
	if len(keysAndValues)%2 == 1 {
		tracer.entry.Error("Error: len of keysAndValues should be even")
		return
	}

	msgWithContext := tracer.concatenateContextToMsg(msg)
	if keysAndValues != nil && len(keysAndValues) > 0 {
		tracer.entry.Info(msgWithContext, keysAndValues)
	} else {
		tracer.entry.Info(msgWithContext)
	}
}

// Error writes error. delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
func (tracer *WrapperTivanTracer) Error(err error, msg string, keysAndValues ...interface{}) {
	// Check that the level of the tracer is enabled, and the len of keysAndValues are even:
	if !tracer.level.Enabled(zapcore.ErrorLevel) {
		return
	}
	if len(keysAndValues)%2 == 1 {
		tracer.entry.Error("Error: len of keysAndValues should be even")
		return
	}

	msgWithContext := tracer.concatenateContextToMsg(msg)
	if keysAndValues != nil && len(keysAndValues) > 0 {
		tracer.entry.Error(msgWithContext, err, keysAndValues)
	} else {
		tracer.entry.Error(msgWithContext, err)
	}
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append suffixes to the logger's name.
func (tracer *WrapperTivanTracer) WithName(name string) logr.Logger {
	newTracer := tracer.addNewContext(name)
	return newTracer
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (tracer *WrapperTivanTracer) addNewContext(suffix string) *WrapperTivanTracer {
	if suffix == "" {
		return tracer
	}
	clonedTracer := tracer.clone()
	// In case that context is empty, don't add . as separator
	if tracer.context == "" {
		clonedTracer.context = suffix
		// concatenate the suffix to the current context.
	} else {
		clonedTracer.context = strings.Join([]string{clonedTracer.context, suffix}, ".")
	}
	return clonedTracer
}

// clone the tracer.
func (tracer *WrapperTivanTracer) clone() *WrapperTivanTracer {
	copiedTracer := *tracer
	return &copiedTracer
}

// concatenateContextToMsg is concatenating the trace to the msg with specific encoding.
func (tracer *WrapperTivanTracer) concatenateContextToMsg(msg string) (newMsg string) {
	switch tracer.encoder {
	case trace.JSON:
		return tracer.concatenateTraceToMsgJson(msg)
	case trace.NONE:
		fallthrough
	default:
		return tracer.concatenateTraceToMsgNone(msg)
	}
}

// concatenateTraceToMsgNone is concatenating the trace to the msg as tuple without encoding:
// e.g. tracer.trace = "server.tag2digest", msg = "digest was resolved" -->
// the return value will be : "trace: server.tag2digest, msg: digest was resolved"
func (tracer *WrapperTivanTracer) concatenateTraceToMsgNone(msg string) (newMsg string) {
	return fmt.Sprintf(`context: %s,msg: %s`, tracer.context, msg)
}

// concatenateTraceToMsgJson is concatenating the trace to the msg as json:
// e.g. tracer.trace = "server.tag2digest", msg = "digest was resolved" -->
// the return value will be : "{"trace": "server.tag2digest", msg: "digest was resolved"}
func (tracer *WrapperTivanTracer) concatenateTraceToMsgJson(msg string) (newMsg string) {
	return fmt.Sprintf(`{"context":"%s","msg":"%s"}`, tracer.context, msg)
}

// V implements this method only for logr.Logger interface.
//				******* DOESN'T DO ANYTHING - RETURN THE SAME TRACER *******
func (tracer *WrapperTivanTracer) V(level int) logr.Logger {
	return tracer
}

// WithValues implements this method only for logr.Logger interface.
//				******* DOESN'T DO ANYTHING - RETURN THE SAME TRACER *******
func (tracer *WrapperTivanTracer) WithValues(keysAndValues ...interface{}) logr.Logger {
	return tracer
}

// Enabled tests whether this Logger is enabled.
//delegate this method using Tracer.Entry data member in order to implement logr.Logger interface
//				******* DOESN'T DO ANYTHING - RETURN THE SAME TRACER *******
func (tracer *WrapperTivanTracer) Enabled() bool {
	return tracer.entry != nil
}
