package instrumentation

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
)

// InstrumentationProvider struct manage the instrumentation of the server
type InstrumentationProvider struct {
	// Tracer is the tracer (trace.ITracer)  of the instrumentation
	Tracer trace.ITracer
	// MetricSubmitter is the tracer (metric.IMetricSubmitter)  of the instrumentation
	MetricSubmitter metric.IMetricSubmitter
}

// NewInstrumentationProvider  Ctor for InstrumentationProvider
func NewInstrumentationProvider(tracer trace.ITracer, metricSubmitter metric.IMetricSubmitter) (provider *InstrumentationProvider) {
	return &InstrumentationProvider{
		Tracer:          tracer,
		MetricSubmitter: metricSubmitter,
	}
}

// GetTracer implements IInstrumentationProvider.GetTracer method of IInstrumentationProvider interface - register the logger with the context.
func (provider *InstrumentationProvider) GetTracer(context string) (tracer trace.ITracer) {
	return provider.Tracer.WithName(context)
}

// GetMetricSubmitter implements IInstrumentationProvider.GetMetricSubmitter method of IInstrumentationProvider interface
func (provider *InstrumentationProvider) GetMetricSubmitter() (metricSubmitter metric.IMetricSubmitter) {
	return provider.MetricSubmitter
}
