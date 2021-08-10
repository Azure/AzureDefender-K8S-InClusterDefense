package instrumentation

import (
	"github.com/pkg/errors"
	tivanInstrumentation "tivan.ms/libs/instrumentation"
)

// Default instrumentation configuration values
const (
	// componentName is the component name that Tivan's needs for the instrumentation.
	componentName = "AzDProxy"
	// mdmNamespace is the mdm name that Tivan's needs for the instrumentation.
	mdmNamespace = "Tivan.Collector.Pods" //TODO Check if I can change this mdmNameSpace ??
)

// GetTivanInstrumentationResult gets tivan instrumentation Result.
func GetTivanInstrumentationResult() (instrumentationResult *tivanInstrumentation.InstrumentationInitializationResult, err error) {
	tivanConfiguration := getInstrumentationConfiguration()
	instrumentationInitializer := tivanInstrumentation.NewInstrumentationInitializer(tivanConfiguration)
	instrumentationResult, err = instrumentationInitializer.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "error encountered during tracer initialization")
	}
	return instrumentationResult, nil
}

// getInstrumentationConfiguration - Get Instrumentation Initialization Configuration
func getInstrumentationConfiguration() *tivanInstrumentation.InstrumentationConfiguration {
	// TODO Use Tivan default configuration - should be changed when we will have our configuration
	instrumentationConfiguration := tivanInstrumentation.NewInstrumentationConfigurationFromEnv(componentName, mdmNamespace)
	return instrumentationConfiguration
}
