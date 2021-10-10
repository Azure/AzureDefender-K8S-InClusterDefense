package metric

// DimensionlessMetric implements metric.IMetric interface
var _ IMetric = (*DimensionlessMetric)(nil)

// DimensionlessMetric implementation of metric.IMetric, for metric without dimensions
type DimensionlessMetric struct {
	metricName string
}

// NewDimensionlessMetric Cto'r for DimensionlessMetric
func NewDimensionlessMetric(metricName string) *DimensionlessMetric {
	return &DimensionlessMetric{
		metricName: metricName,
	}
}

func (metric *DimensionlessMetric) MetricName() string {
	return metric.metricName
}

func (metric *DimensionlessMetric) MetricDimension() []Dimension {
	return []Dimension{}
}
