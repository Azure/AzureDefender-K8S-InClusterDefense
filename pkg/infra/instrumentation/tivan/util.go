package tivan

import (
	"github.com/pkg/errors"
	tivanInstrumentation "tivan.ms/libs/instrumentation"
)

// TivanInstrumentationConfiguration is the configuration that is needed to create tivan's instrumentation
type TivanInstrumentationConfiguration struct {
	// componentName is the component name that Tivan's needs for the instrumentation.
	ComponentName string
	// mdmNamespace is the mdm name that Tivan's needs for the instrumentation.
	MdmNamespace string
}

// NewTivanInstrumentationResult Creates new tivanInstrumentation.InstrumentationInitializationResult - tivan instrumentation object.
// It is creating it by creating tivan's configuration, and then initialize the instrumentation
func NewTivanInstrumentationResult(configuration *TivanInstrumentationConfiguration) (instrumentationResult *tivanInstrumentation.InstrumentationInitializationResult, err error) {
	tivanConfiguration := newTivanInstrumentationConfiguration(configuration)
	instrumentationInitializer := tivanInstrumentation.NewInstrumentationInitializer(tivanConfiguration)
	instrumentationResult, err = instrumentationInitializer.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "tivan.NewTivanInstrumentationResult: error encountered during tracer initialization")
	}
	return instrumentationResult, nil
}

// newTivanInstrumentationConfiguration - Get Instrumentation Initialization Configuration
func newTivanInstrumentationConfiguration(configuration *TivanInstrumentationConfiguration) *tivanInstrumentation.InstrumentationConfiguration {
	instrumentationConfiguration := tivanInstrumentation.NewInstrumentationConfigurationFromEnv(configuration.ComponentName, configuration.MdmNamespace)
	return instrumentationConfiguration
}
