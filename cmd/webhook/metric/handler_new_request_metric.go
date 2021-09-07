package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// HandlerNewRequestMetric is metric for the requests kinds of the handler
type HandlerNewRequestMetric struct {
	// requestKing is the kind that the handler was got in the request (e.g. Pod).
	requestKind string
}

// NewHandlerNewRequestMetric Ctor
func NewHandlerNewRequestMetric(kind string) *HandlerNewRequestMetric {
	return &HandlerNewRequestMetric{
		requestKind: kind,
	}
}

func (m *HandlerNewRequestMetric) MetricName() string {
	return "HandlerNewRequest"
}

func (m *HandlerNewRequestMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "RequestKind", Value: m.requestKind},
	}
}
