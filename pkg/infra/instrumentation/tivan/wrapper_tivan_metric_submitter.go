package tivan

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	tivanInstrumentation "tivan.ms/libs/instrumentation"
)

// WrapperTivanMetricSubmitter implements metric.IMetricSubmitter  interface
var _ metric.IMetricSubmitter = (*WrapperTivanMetricSubmitter)(nil)

type WrapperTivanMetricSubmitter struct {
	//tivanMetricSubmitter is wrapper for tivan's metric submitter.
	tivanMetricSubmitter tivanInstrumentation.MetricSubmitter
}

// NewWrapperTivanMetricSubmitter  Ctor for WrapperTivanMetricSubmitter
func NewWrapperTivanMetricSubmitter(tivanMetricSubmitter tivanInstrumentation.MetricSubmitter) *WrapperTivanMetricSubmitter {
	return &WrapperTivanMetricSubmitter{tivanMetricSubmitter: tivanMetricSubmitter}
}

// SendMetric send metric using tivan's metric submitter.
func (wrapper *WrapperTivanMetricSubmitter) SendMetric(value int, metric metric.IMetric) {
	wrapper.tivanMetricSubmitter.SendMetric(value, NewTivanMetric(metric))
}
