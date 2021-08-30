package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// HandlerNumOfRequestsMetric is metric for the requests kinds of the handler
type HandlerNumOfRequestsMetric struct {
	// requestKing is the kind that the handler was got in the request (e.g. Pod).
	requestKind string
}

// NewHandlerNumOfRequestsMetric Ctor
func NewHandlerNumOfRequestsMetric(kind string) *HandlerNumOfRequestsMetric {
	return &HandlerNumOfRequestsMetric{
		requestKind: kind,
	}
}

func (m *HandlerNumOfRequestsMetric) MetricName() string {
	return "HandlerNumOfRequests"
}

func (m *HandlerNumOfRequestsMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "RequestKind", Value: m.requestKind},
	}
}
