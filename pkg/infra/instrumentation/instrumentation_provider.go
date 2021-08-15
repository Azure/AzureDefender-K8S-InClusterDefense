package instrumentation

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
)

// InstrumentationProvider struct manage the instrumentation of the server
type InstrumentationProvider struct {
	// tracer is the tracer (trace.ITracer)  of the instrumentation
	tracer trace.ITracer
	// metricSubmitter is the tracer (metric.IMetricSubmitter)  of the instrumentation
	metricSubmitter metric.IMetricSubmitter
}

// NewInstrumentationProvider  Ctor for InstrumentationProvider
func NewInstrumentationProvider(tracer trace.ITracer, metricSubmitter metric.IMetricSubmitter) (provider *InstrumentationProvider) {
	return &InstrumentationProvider{
		tracer:          tracer,
		metricSubmitter: metricSubmitter,
	}
}

// GetTracerProvider implements IInstrumentationProvider.GetTracer method of IInstrumentationProvider interface - register the logger with the context.
func (provider *InstrumentationProvider) GetTracerProvider(context string) (tracer trace.ITracerProvider) {
	return trace.NewTracerProvider(provider.tracer, context)
}

// GetMetricSubmitter implements IInstrumentationProvider.GetMetricSubmitter method of IInstrumentationProvider interface
func (provider *InstrumentationProvider) GetMetricSubmitter() (metricSubmitter metric.IMetricSubmitter) {
	return provider.metricSubmitter
}
