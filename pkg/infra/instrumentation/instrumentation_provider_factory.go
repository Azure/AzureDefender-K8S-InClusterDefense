package instrumentation

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
)

// IInstrumentationProviderFactory InstrumentationFactory - Instrumentation factory interface
type IInstrumentationProviderFactory interface {
	// CreateInstrumentation  creates instrumentation and returns err in case that the creation was failed.
	CreateInstrumentationProvider() (IInstrumentationProvider, error)
}

// InstrumentationProviderFactory a factory for creating a instrumentation entry
type InstrumentationProviderFactory struct {
	// configuration is the instrumentation configuration
	configuration *InstrumentationProviderConfiguration
	//tracerFactory is the factory that creates the tracer.
	tracerFactory trace.ITracerFactory
	// metricSubmitterFactory is the factory that creates the metric submitter
	metricSubmitterFactory metric.IMetricSubmitterFactory
}

// InstrumentationProviderConfiguration is the configuration of the instrumentation.
type InstrumentationProviderConfiguration struct {
}

// NewInstrumentationProviderFactory returns new InstrumentationFactory
func NewInstrumentationProviderFactory(
	configuration *InstrumentationProviderConfiguration,
	tracerFactory trace.ITracerFactory,
	metricSubmitterFactory metric.IMetricSubmitterFactory) (factory IInstrumentationProviderFactory) {
	return &InstrumentationProviderFactory{
		configuration:          configuration,
		tracerFactory:          tracerFactory,
		metricSubmitterFactory: metricSubmitterFactory,
	}
}

// CreateInstrumentation creates instrumentation using Tivan infra.
func (factory *InstrumentationProviderFactory) CreateInstrumentationProvider() (instrumentationProvider IInstrumentationProvider, err error) {
	metricSubmitter := factory.metricSubmitterFactory.CreateMetricSubmitter()
	tracer, err := factory.tracerFactory.CreateTracer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tracer while creating instrumentation")
	}
	return &InstrumentationProvider{Tracer: tracer, MetricSubmitter: metricSubmitter}, nil
}
