package tivan

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	tivanInstrumentation "tivan.ms/libs/instrumentation"
)

// tivanMetric is struct that gets a metric.IMetric and warp it with implementations of tivanInstrumentation.Metric
type tivanMetric struct {
	wrappedMetric metric.IMetric
}

// NewTivanMetric Ctor for tivanMetric
func NewTivanMetric(metric metric.IMetric) tivanInstrumentation.Metric {
	return &tivanMetric{
		wrappedMetric: metric,
	}
}

// MetricName - getter for the metric name
func (m *tivanMetric) MetricName() string {
	return m.wrappedMetric.MetricName()
}

// MetricDimension - getter for the metric dimensions
func (m *tivanMetric) MetricDimension() []tivanInstrumentation.Dimension {
	// Convert dimensions
	dimensions := m.wrappedMetric.MetricDimension()
	tivanDimensions := make([]tivanInstrumentation.Dimension, 0, len(dimensions))
	for _, dimension := range dimensions {
		tivanDimensions = append(tivanDimensions, tivanInstrumentation.Dimension{Key: dimension.Key, Value: dimension.Value})
	}

	return tivanDimensions
}
