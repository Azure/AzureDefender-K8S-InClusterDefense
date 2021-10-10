package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// HandlerNumOfContainersPerPodMetric implements metric.IMetric  interface
var _ metric.IMetric = (*HandlerNumOfContainersPerPodMetric)(nil)

// HandlerNumOfContainersPerPodMetric is dimensionless that counts how many containers
type HandlerNumOfContainersPerPodMetric struct {
}

// NewHandlerNumOfContainersPerPodMetric Ctor
func NewHandlerNumOfContainersPerPodMetric() *HandlerNumOfContainersPerPodMetric {
	return &HandlerNumOfContainersPerPodMetric{}
}

func (m *HandlerNumOfContainersPerPodMetric) MetricName() string {
	return "HandlerNumOfContainersPerPod"
}

func (m *HandlerNumOfContainersPerPodMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
