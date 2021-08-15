package metric

// IMetricSubmitterFactory is factory for metric submitter
type IMetricSubmitterFactory interface {
	// CreateMetricSubmitter creates new IMetricSubmitter.
	CreateMetricSubmitter() (metricSubmitter IMetricSubmitter)
}
