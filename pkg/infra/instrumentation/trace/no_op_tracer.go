package trace

import "github.com/go-logr/logr"

// NoOpTracer is implementation that does nothing of ITracer
type NoOpTracer struct{}

// NewNoOpTracer Ctor for NoOpTracer
func NewNoOpTracer() *NoOpTracer {
	return &NoOpTracer{}
}

func (tracer *NoOpTracer) Enabled() bool {
	return false
}

func (tracer *NoOpTracer) Info(msg string, keysAndValues ...interface{}) {
}

func (tracer *NoOpTracer) Error(err error, msg string, keysAndValues ...interface{}) {
}

func (tracer *NoOpTracer) V(level int) logr.Logger {
	return tracer
}

func (tracer *NoOpTracer) WithValues(keysAndValues ...interface{}) logr.Logger {
	return tracer
}

func (tracer *NoOpTracer) WithName(name string) logr.Logger {
	return tracer
}
