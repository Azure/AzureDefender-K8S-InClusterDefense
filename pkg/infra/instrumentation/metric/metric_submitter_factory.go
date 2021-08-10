package metric

import tivanInstrumentation "tivan.ms/libs/instrumentation"

// IMetricSubmitterFactory is factory for metric submitter
type IMetricSubmitterFactory interface {
	// CreateMetricSubmitter creates new IMetricSubmitter.
	CreateMetricSubmitter() (metricSubmitter IMetricSubmitter)
}

// MetricSubmitterFactory implement the IMetricSubmitterFactory and creates the default metric submitter.
// It wraps Tivan's MetricSubmitter
type MetricSubmitterFactory struct {
	// configuration is the configuration of the metric submitter
	configuration *MetricSubmitterConfiguration
	// metricsubmitter is the metric submitter of tivan
	metricSubmitter *tivanInstrumentation.MetricSubmitter
}

// MetricSubmitterConfiguration is the configuration for the metric submitter.
type MetricSubmitterConfiguration struct {
}

// NewMetricSubmitterFactory creates MetricSubmitterFactory that creates metric submitter by wrapping the metric submitter of Tivan
func NewMetricSubmitterFactory(configuration *MetricSubmitterConfiguration, metricSubmitter *tivanInstrumentation.MetricSubmitter) (factory IMetricSubmitterFactory) {
	return MetricSubmitterFactory{
		configuration:   configuration,
		metricSubmitter: metricSubmitter,
	}
}

// CreateMetricSubmitter creates new IMetricSubmitter by using the metric submitter of Tivan
func (factory MetricSubmitterFactory) CreateMetricSubmitter() (metricSubmitter IMetricSubmitter) {
	return *factory.metricSubmitter
}
