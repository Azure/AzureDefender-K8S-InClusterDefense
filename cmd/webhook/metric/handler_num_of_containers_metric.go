package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// HandlerNumOfContainersMetric is dimensionless that counts how many containers
type HandlerNumOfContainersMetric struct {
}

// NewHandlerNumOfContainersMetric Ctor
func NewHandlerNumOfContainersMetric() *HandlerNumOfContainersMetric {
	return &HandlerNumOfContainersMetric{}
}

func (m *HandlerNumOfContainersMetric) MetricName() string {
	return "HandlerNumOfContainers"
}

func (m *HandlerNumOfContainersMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
