package tivan

import (
	tivanInstrumentation "github.com/Azure/ASC-go-libs/pkg/instrumentation"
	"github.com/pkg/errors"
)

// NewTivanInstrumentationResult Creates new tivanInstrumentation.InstrumentationInitializationResult - tivan instrumentation object.
// It is creating it by creating tivan's configuration, and then initialize the instrumentation
func NewTivanInstrumentationResult(configuration *tivanInstrumentation.InstrumentationConfiguration) (instrumentationResult *tivanInstrumentation.InstrumentationInitializationResult, err error) {
	instrumentationInitializer := tivanInstrumentation.NewInstrumentationInitializer(configuration)
	instrumentationResult, err = instrumentationInitializer.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "tivan.NewTivanInstrumentationResult: error encountered during tracer initialization")
	}
	return instrumentationResult, nil
}
