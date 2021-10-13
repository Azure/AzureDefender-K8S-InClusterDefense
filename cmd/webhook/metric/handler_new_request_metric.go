package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	admissionv1 "k8s.io/api/admission/v1"
)

// HandlerNewRequestMetric implements metric.IMetric  interface
var _ metric.IMetric = (*HandlerNewRequestMetric)(nil)

// HandlerNewRequestMetric is metric for the requests kinds of the handler
type HandlerNewRequestMetric struct {
	// requestKing is the kind that the handler was got in the request (e.g. Pod).
	requestKind string
	// requestOperation is the operation type of the request (create/update)
	requestOperation string
}

// NewHandlerNewRequestMetric Ctor
func NewHandlerNewRequestMetric(kind string, operation admissionv1.Operation) *HandlerNewRequestMetric {
	return &HandlerNewRequestMetric{
		requestKind:      kind,
		requestOperation: string(operation),
	}
}

func (m *HandlerNewRequestMetric) MetricName() string {
	return "HandlerNewRequest"
}

func (m *HandlerNewRequestMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "RequestKind", Value: m.requestKind},
		{Key: "RequestOperation", Value: m.requestOperation},
	}
}
