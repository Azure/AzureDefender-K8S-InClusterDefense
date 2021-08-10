package instrumentation

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
)

// IInstrumentationFactory InstrumentationFactory - Instrumentation factory interface
type IInstrumentationFactory interface {
	// CreateInstrumentation  creates instrumentation and returns err in case that the creation was failed.
	CreateInstrumentation() (*Instrumentation, error)
}

// InstrumentationFactory a factory for creating a instrumentation entry
type InstrumentationFactory struct {
	// configuration is the instrumentation configuration
	configuration *InstrumentationConfiguration
	//tracerFactory is the factory that creates the tracer.
	tracerFactory trace.ITracerFactory
	// metricSubmitterFactory is the factory that creates the metric submitter
	metricSubmitterFactory metric.IMetricSubmitterFactory
}

// Instrumentation struct manage the instrumentation of the server
type Instrumentation struct {
	// Tracer is the tracer (trace.ITracer)  of the instrumentation
	Tracer trace.ITracer
	// MetricSubmitter is the tracer (metric.IMetricSubmitter)  of the instrumentation
	MetricSubmitter metric.IMetricSubmitter
}

// InstrumentationConfiguration is the configuration of the instrumentation.
type InstrumentationConfiguration struct {
}

// NewInstrumentationFactory returns new InstrumentationFactory
func NewInstrumentationFactory(
	configuration *InstrumentationConfiguration,
	tracerFactory trace.ITracerFactory,
	metricSubmitterFactory metric.IMetricSubmitterFactory) (factory IInstrumentationFactory) {
	return &InstrumentationFactory{
		configuration:          configuration,
		tracerFactory:          tracerFactory,
		metricSubmitterFactory: metricSubmitterFactory,
	}
}

// CreateInstrumentation creates instrumentation using Tivan infra.
func (factory InstrumentationFactory) CreateInstrumentation() (newInstrumentation *Instrumentation, err error) {
	metricSubmitter := factory.metricSubmitterFactory.CreateMetricSubmitter()
	tracer, err := factory.tracerFactory.CreateTracer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tracer while creating instrumentation")
	}
	return &Instrumentation{Tracer: tracer, MetricSubmitter: metricSubmitter}, nil
}
