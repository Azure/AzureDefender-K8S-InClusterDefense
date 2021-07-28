package instrumentation

import (
	tivanInstrumentation "dev.azure.com/msazure/One/_git/Rome-Detection-Tivan-Libs.git/src/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/pkg/infra/configuration"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// Instrumentation struct manage the instrumentation of the server
type Instrumentation struct {
	Tracer          logr.Logger
	MetricSubmitter tivanInstrumentation.MetricSubmitter
}

// IInstrumentationFactory InstrumentationFactory - Instrumentation factory interface
type IInstrumentationFactory interface {
	CreateInstrumentation() (*Instrumentation, error)
}

// InstrumentationFactory a factory for creating a instrumentation entry
type InstrumentationFactory struct {
	//TODO add instrumentation configuration
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
		return nil, errors.New("error encountered during tracer initialization")
	}
	newInstrumentation.MetricSubmitter = instrumentationInitializationResults.MetricSubmitter
	// Decorate Tivan's tracer with the logr.Logger interface.
	newInstrumentation.Tracer = NewTracerFactory().CreateTracer(instrumentationInitializationResults.Tracer)

	return newInstrumentation, nil
}
