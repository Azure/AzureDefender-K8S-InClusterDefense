package instrumenation

import (
	instrumenation "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/logger"
	ctrl "sigs.k8s.io/controller-runtime"
)

// TracerFactory implementation of ITracerFactory.
type TracerFactory struct {
	Configuration instrumenation.TracerConfiguration
}

// NewTracerFactory creates new TracerFactory instance.
func NewTracerFactory(configuration instrumenation.TracerConfiguration) instrumenation.ITracerFactory {
	return &TracerFactory{
		Configuration: configuration,
	}
}

// CreateTracer Creates tracer
func (tracerFactory TracerFactory) CreateTracer() (tracer instrumenation.ITracer, err error) {
	tracer = &Tracer{
		Entry: tracerFactory.Configuration.Entry,
		level: tracerFactory.Configuration.TracerLevel,
		trace: tracerFactory.Configuration.DefaultTrace,
	}
	// Register the tracer as the main trace - without this, loggers won't be initialized (e.g crLog at cert-controller)
	ctrl.SetLogger(tracer)
	return tracer, nil
}
