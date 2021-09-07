package metric

// IMetric interface for getting the metric name and metric dimensions
type IMetric interface {
	// MetricName - getter for the metric name
	MetricName() string
	// MetricDimension - getter for the metric dimensions
	MetricDimension() []Dimension
}
