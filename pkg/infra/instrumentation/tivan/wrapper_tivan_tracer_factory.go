package tivan

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	log "github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

// TracerFactory implements trace.ITracerFactory  interface
var _ trace.ITracerFactory = (*TracerFactory)(nil)

// TracerFactory implementation of ITracerFactory.
type TracerFactory struct {
	// configuration is the configuration of the tracer.
	configuration *trace.TracerConfiguration
	// Entry is needed for Tivan's instrumentation
	entry *log.Entry
}

// NewTracerFactory creates new TracerFactory instance.
func NewTracerFactory(configuration *trace.TracerConfiguration, entry *log.Entry) trace.ITracerFactory {
	return &TracerFactory{
		configuration: configuration,
		entry:         entry,
	}
}

// CreateTracer Creates tracer
func (factory *TracerFactory) CreateTracer() (tracer trace.ITracer) {
	tracer = NewWrapperTivanTracer(factory.entry, factory.configuration.DefaultContext, factory.configuration.TracerLevel, trace.NONE)

	// Register the tracer as the main trace - without this, loggers won't be initialized (e.g crLog at cert-controller)
	ctrl.SetLogger(tracer)
	return tracer
}
