package instrumentation

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/configuration"
	"github.com/go-logr/logr"
	tivanInstrumentation "tivan.ms/libs/instrumentation"
)

// IInstrumentationFactory InstrumentationFactory - Instrumentation factory interface
type IInstrumentationFactory interface {
	CreateInstrumentation() (*Instrumentation, error)
}

// InstrumentationFactory a factory for creating a instrumentation entry
type InstrumentationFactory struct {
	//TODO add instrumentation configuration
}

// Instrumentation struct manage the instrumentation of the server
type Instrumentation struct {
	Tracer          logr.Logger
	MetricSubmitter tivanInstrumentation.MetricSubmitter
}

// NewInstrumentationFactory returns new InstrumentationFactory
func NewInstrumentationFactory() (InstrumentationFactoryImpl *InstrumentationFactory) {
	return &InstrumentationFactory{}
}

// CreateInstrumentation creates instrumentation using Tivan infra.
func (f *InstrumentationFactory) CreateInstrumentation() (newInstrumentation *Instrumentation, err error) {
	newInstrumentation = &Instrumentation{}
	// Create Tivan's instrumentation:
	instrumentationConfiguration := configuration.GetInstrumentationConfiguration()
	instrumentationInitializer := tivanInstrumentation.NewInstrumentationInitializer(instrumentationConfiguration)
	instrumentationInitializationResults, err := instrumentationInitializer.Initialize()
	if err != nil {
		return nil, fmt.Errorf("error encountered during tracer initialization: %v", err)
	}
	newInstrumentation.MetricSubmitter = instrumentationInitializationResults.MetricSubmitter
	// Decorate Tivan's tracer with the logr.Logger interface.
	newInstrumentation.Tracer = NewTracerFactory().CreateTracer(instrumentationInitializationResults.Tracer)

	return newInstrumentation, nil
}
