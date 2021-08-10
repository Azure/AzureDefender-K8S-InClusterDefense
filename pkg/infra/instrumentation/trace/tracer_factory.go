package trace

import (
	log "github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

// TracerFactory implementation of ITracerFactory.
type TracerFactory struct {
	// configuration is the configuration of the tracer.
	configuration *TracerConfiguration
	// Entry is needed for Tivan's instrumentation
	Entry *log.Entry
}

// NewTracerFactory creates new TracerFactory instance.
func NewTracerFactory(configuration *TracerConfiguration, entry *log.Entry) ITracerFactory {
	return &TracerFactory{
		configuration: configuration,
		Entry:         entry,
	}
}

// CreateTracer Creates tracer
func (tracerFactory TracerFactory) CreateTracer() (tracer ITracer, err error) {
	tracer = &Tracer{
		Entry:   tracerFactory.Entry,
		level:   tracerFactory.configuration.TracerLevel,
		context: tracerFactory.configuration.DefaultContext,
	}
	// Register the tracer as the main trace - without this, loggers won't be initialized (e.g crLog at cert-controller)
	ctrl.SetLogger(tracer)
	return tracer, nil
}
