package metric

import tivanInstrumentation "tivan.ms/libs/instrumentation"

type IMetricSubmitter interface {
	// SendMetric - send metric by name with provided dimensions
	SendMetric(value int, metric tivanInstrumentation.Metric)
}
