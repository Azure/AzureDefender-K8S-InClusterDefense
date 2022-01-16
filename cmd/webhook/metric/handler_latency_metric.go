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
	responseStatusCode int
	patchCount int
}

// NewHandlerHandleLatencyMetric Ctor
func NewHandlerHandleLatencyMetric(kind string, responseAllowed bool, responseReason string, responseStatusCode int32, patchCount int) *HandlerHandleLatencyMetric {
	return &HandlerHandleLatencyMetric{
		requestKind: kind,
		responseAllowed: responseAllowed,
		responseReason: responseReason,
		responseStatusCode: int(responseStatusCode),
		patchCount: patchCount,
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
		{Key: "ResponseReasonStatusCode", Value: strconv.Itoa(m.responseStatusCode)},
		{Key: "PatchCount", Value: strconv.Itoa(m.patchCount)},
	}
}
