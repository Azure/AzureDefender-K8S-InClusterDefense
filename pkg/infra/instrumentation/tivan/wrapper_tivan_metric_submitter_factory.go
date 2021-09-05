package tivan

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	tivanInstrumentation "tivan.ms/libs/instrumentation"
)

// WrapperTivanMetricSubmitterFactory implement the IMetricSubmitterFactory and creates the default metric submitter.
// It wraps Tivan's MetricSubmitter
type WrapperTivanMetricSubmitterFactory struct {
	// configuration is the configuration of the metric submitter
	configuration *MetricSubmitterConfiguration
	// metricSubmitter is the metric submitter of tivan
	metricSubmitter tivanInstrumentation.MetricSubmitter
}

// MetricSubmitterConfiguration is the configuration for the metric submitter.
type MetricSubmitterConfiguration struct {
}

// NewMetricSubmitterFactory creates WrapperTivanMetricSubmitterFactory that creates metric submitter by wrapping the metric submitter of Tivan
func NewMetricSubmitterFactory(configuration *MetricSubmitterConfiguration, metricSubmitter tivanInstrumentation.MetricSubmitter) (factory metric.IMetricSubmitterFactory) {
	return &WrapperTivanMetricSubmitterFactory{
		configuration:   configuration,
		metricSubmitter: metricSubmitter,
	}
}

// CreateMetricSubmitter creates new IMetricSubmitter by using the metric submitter of Tivan
func (factory *WrapperTivanMetricSubmitterFactory) CreateMetricSubmitter() (metricSubmitter metric.IMetricSubmitter) {
	return NewWrapperTivanMetricSubmitter(factory.metricSubmitter)
}
