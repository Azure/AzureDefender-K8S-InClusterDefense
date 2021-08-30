package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// HandlerLatencyMetric is the latency metric of the handler
type HandlerLatencyMetric struct {
}

// NewHandlerLatencyMetric Ctor
func NewHandlerLatencyMetric() *HandlerLatencyMetric {
	return &HandlerLatencyMetric{}
}

func (m *HandlerLatencyMetric) MetricName() string {
	return "ArgDataProviderResponseLatency"
}

func (m *HandlerLatencyMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
