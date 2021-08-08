package instrumentation

import (
	instrumenation2 "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/logger"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	tivanInstrumentation "tivan.ms/libs/instrumentation"
)

// IInstrumentationFactory InstrumentationFactory - Instrumentation factory interface
type IInstrumentationFactory interface {
	CreateInstrumentation() (*Instrumentation, error)
}

// InstrumentationFactory a factory for creating a instrumentation entry
type InstrumentationFactory struct {
	Configuration          InstrumentationConfiguration
	tracerFactory          instrumenation2.ITracerFactory
	metricSubmitterFactory metric.IMetricSubmitterFactory
}

// Instrumentation struct manage the instrumentation of the server
type Instrumentation struct {
	Tracer          logr.Logger
	MetricSubmitter tivanInstrumentation.MetricSubmitter
}

type InstrumentationConfiguration struct {
}

// NewInstrumentationFactory returns new InstrumentationFactory
func NewInstrumentationFactory(
	configuration InstrumentationConfiguration,
	tracerFactory instrumenation2.ITracerFactory,
	metricSubmitterFactory metric.IMetricSubmitterFactory) (InstrumentationFactoryImpl IInstrumentationFactory) {
	return &InstrumentationFactory{
		Configuration:          configuration,
		tracerFactory:          tracerFactory,
		metricSubmitterFactory: metricSubmitterFactory,
	}
}

// CreateInstrumentation creates instrumentation using Tivan infra.
func (factory *InstrumentationFactory) CreateInstrumentation() (newInstrumentation *Instrumentation, err error) {
	newInstrumentation = &Instrumentation{}
	// Create Tivan's instrumentation:
	//configuration := factory.Configuration
	//instrumentationConfiguration := configuration.GetInstrumentationConfiguration()
	//instrumentationInitializer := tivanInstrumentation.NewInstrumentationInitializer(instrumentationConfiguration)
	//instrumentationInitializationResults, err := instrumentationInitializer.Initialize()
	//if err != nil {
	//	return nil, errors.Wrap(err, "error encountered during tracer initialization")
	//}
	//newInstrumentation.MetricSubmitter = instrumentationInitializationResults.MetricSubmitter
	newInstrumentation.MetricSubmitter = factory.metricSubmitterFactory.CreateMetricSubmitter()
	// Decorate Tivan's tracer with the logr.Logger interface.
	//newInstrumentation.Tracer, err = instrumenation.NewTracerFactory().CreateTracer(instrumentationInitializationResults.Tracer)
	newInstrumentation.Tracer, err = factory.tracerFactory.CreateTracer()
	if err != nil {
		return nil, errors.Wrap(err, "error encountered during tracer initialization")
	}

	return newInstrumentation, nil
}
