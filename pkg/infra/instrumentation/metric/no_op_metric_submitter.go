package metric

// NoOpMetricSubmitter is implementation that does nothing of IMetricSubmitter
type NoOpMetricSubmitter struct {
}

func NewNoOpMetricSubmitter() IMetricSubmitter {
	return &NoOpMetricSubmitter{}
}

// SendMetric send metric
func (metricSubmitter *NoOpMetricSubmitter) SendMetric(value int, metric IMetric) {
}
