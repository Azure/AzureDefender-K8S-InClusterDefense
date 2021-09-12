package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// HandlerHandleLatencyMetric implements metric.IMetric  interface
var _ metric.IMetric = (*HandlerHandleLatencyMetric)(nil)

// HandlerHandleLatencyMetric is the latency metric of the handler
type HandlerHandleLatencyMetric struct {
}

// NewHandlerHandleLatencyMetric Ctor
func NewHandlerHandleLatencyMetric() *HandlerHandleLatencyMetric {
	return &HandlerHandleLatencyMetric{}
}

func (m *HandlerHandleLatencyMetric) MetricName() string {
	return "HandlerHandleLatency"
}

func (m *HandlerHandleLatencyMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
