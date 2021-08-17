package instrumentation

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
)

// IInstrumentationProvider is the instrumentation provider interface that provides tracer and metric submitter
type IInstrumentationProvider interface {
	// GetTracerProvider is function that gets a string as the context of the tracer and returns new tracer provider
	//with context of the struct.
	GetTracerProvider(context string) (tracer trace.ITracerProvider)
	// GetMetricSubmitter is function that returns metric submitter.
	GetMetricSubmitter() (tracer metric.IMetricSubmitter)
}
