package metric

import tivanInstrumentation "tivan.ms/libs/instrumentation"

type IMetricSubmitterFactory interface {
	CreateMetricSubmitter() (metricSubmitter IMetricSubmitter)
}

type MetricSubmitterFactory struct {
	Configuration   MetricSubmitterConfiguration
	metricSubmitter *tivanInstrumentation.MetricSubmitter
}

type MetricSubmitterConfiguration struct {
}

func (factory *MetricSubmitterFactory) MetricSubmitterFactory(results *tivanInstrumentation.InstrumentationInitializationResult) {
	factory.metricSubmitter = &results.MetricSubmitter
}

func (factory MetricSubmitterFactory) CreateMetricSubmitter() (metricSubmitter IMetricSubmitter) {
	return *factory.metricSubmitter
}
