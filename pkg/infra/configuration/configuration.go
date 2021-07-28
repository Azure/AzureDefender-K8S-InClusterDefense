package configuration

import tivanInstrumentation "dev.azure.com/msazure/One/_git/Rome-Detection-Tivan-Libs.git/src/instrumentation"

// Default instrumentation configuration values
const (
	componentName = "AzDProxy"
	mdmNamespace  = "Tivan.Collector.Pods" //TODO Check if I can change this mdmNameSpace ??
)

// GetInstrumentationConfiguration - Get Instrumentation Initialization Configuration
func GetInstrumentationConfiguration() *tivanInstrumentation.InstrumentationConfiguration {
	// TODO Use Tivan default configuration - should be changed when we will have our configuration
	instrumentationConfiguration := tivanInstrumentation.NewInstrumentationConfigurationFromEnv(componentName, mdmNamespace)
	return instrumentationConfiguration
}
