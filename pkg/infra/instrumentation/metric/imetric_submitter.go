package metric

type IMetricSubmitter interface {
	// SendMetric - send metric by name with provided dimensions
	SendMetric(value int, metric IMetric)
}
