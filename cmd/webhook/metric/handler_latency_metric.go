package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"strconv"
)

// HandlerHandleLatencyMetric implements metric.IMetric  interface
var _ metric.IMetric = (*HandlerHandleLatencyMetric)(nil)

// HandlerHandleLatencyMetric is the latency metric of the handler
type HandlerHandleLatencyMetric struct {
	requestKind string
	responseAllowed bool
	responseReason string
	responseStatusCode int32
}

// NewHandlerHandleLatencyMetric Ctor
func NewHandlerHandleLatencyMetric(kind string, responseAllowed bool, responseReason string, responseStatusCode int32) *HandlerHandleLatencyMetric {
	return &HandlerHandleLatencyMetric{
		requestKind: kind,
		responseAllowed: responseAllowed,
		responseReason: responseReason,
		responseStatusCode: responseStatusCode,
	}
}

func (m *HandlerHandleLatencyMetric) MetricName() string {
	return "HandlerHandleLatency"
}

func (m *HandlerHandleLatencyMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "RequestKind", Value: m.requestKind},
		{Key: "ResponseAllowed", Value: strconv.FormatBool(m.responseAllowed)},
		{Key: "ResponseReason", Value: m.responseReason},
		{Key: "ResponseReason", Value: string(m.responseStatusCode)},
	}
}
