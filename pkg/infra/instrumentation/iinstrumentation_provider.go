package instrumentation

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
)

// IInstrumentationProvider is the instrumentation provider interface that provides tracer and metric submitter
type IInstrumentationProvider interface {
	// GetTracer is function that gets a string as the context of the tracer and returns new tracer that starts its context start with context
	GetTracer(context string) (tracer trace.ITracer)
	// GetMetricSubmitter is function that returns metric submitter.
	GetMetricSubmitter() (metricSubmitter metric.IMetricSubmitter)
}
