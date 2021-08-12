package tivan

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	log "github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

// TracerFactory implementation of ITracerFactory.
type TracerFactory struct {
	// configuration is the configuration of the tracer.
	configuration *trace.TracerConfiguration
	// Entry is needed for Tivan's instrumentation
	Entry *log.Entry
}

// NewTracerFactory creates new TracerFactory instance.
func NewTracerFactory(configuration *trace.TracerConfiguration, entry *log.Entry) trace.ITracerFactory {
	return &TracerFactory{
		configuration: configuration,
		Entry:         entry,
	}
}

// CreateTracer Creates tracer
func (tracerFactory TracerFactory) CreateTracer() (tracer trace.ITracer, err error) {
	tracer = NewWrapperTivanTracer(tracerFactory.Entry, tracerFactory.configuration.DefaultContext, tracerFactory.configuration.TracerLevel, trace.NONE)

	// Register the tracer as the main trace - without this, loggers won't be initialized (e.g crLog at cert-controller)
	ctrl.SetLogger(tracer)
	return tracer, nil
}
