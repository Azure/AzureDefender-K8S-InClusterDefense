package instrumentation

import (
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ITracerFactory interface of tracer factory
type ITracerFactory interface {
	CreateTracer(entry *logrus.Entry) *ITracer
}

// TracerFactory implementation of ITracerFactory.
type TracerFactory struct {
}

// NewTracerFactory creates new TracerFactory instance.
func NewTracerFactory() *TracerFactory {
	return &TracerFactory{}
}

// CreateTracer Creates tracer
func (tracerFactory TracerFactory) CreateTracer(entry *logrus.Entry) *Tracer {
	tracer := &Tracer{Entry: entry}
	// Register the tracer as the main trace - without this, loggers won't be initialized (e.g crLog at cert-controller)
	ctrl.SetLogger(tracer)
	return tracer
}
