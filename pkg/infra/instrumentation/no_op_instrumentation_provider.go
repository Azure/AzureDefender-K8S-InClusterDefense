package instrumentation

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
)

// NoOpInstrumentationProvider is implementation that does nothing of IInstrumentationProvider
type NoOpInstrumentationProvider struct {
}

// NewNoOpInstrumentationProvider  Ctor for NoOpInstrumentationProvider
func NewNoOpInstrumentationProvider() (provider IInstrumentationProvider) {
	return &NoOpInstrumentationProvider{}
}

func (n *NoOpInstrumentationProvider) GetTracerProvider(context string) (tracer trace.ITracerProvider) {
	return trace.NewNoOpTracerProvider()
}

func (n *NoOpInstrumentationProvider) GetMetricSubmitter() (metricSubmitter metric.IMetricSubmitter) {
	return metric.NewNoOpMetricSubmitter()
}
