package metric

// NoOpMetricSubmitter implements IMetricSubmitter  interface
var _ IMetricSubmitter = (*NoOpMetricSubmitter)(nil)

// NoOpMetricSubmitter is implementation that does nothing of IMetricSubmitter
// NoOp is used for testing/debugging.
type NoOpMetricSubmitter struct {
}

// NewNoOpMetricSubmitter Ctor for NoOpMetricSubmitter
func NewNoOpMetricSubmitter() IMetricSubmitter {
	return &NoOpMetricSubmitter{}
}

// SendMetric send metric
func (metricSubmitter *NoOpMetricSubmitter) SendMetric(value int, metric IMetric) {
}
