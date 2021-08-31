package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// ArgDataProviderResponseLatencyMetric is metric of ARGDataProvider that checks the latency according to some scanStatus
type ArgDataProviderResponseLatencyMetric struct {
	// scanStatus is the scan status of the metric that would be deployed as ScanStatus metric.Dimension.
	scanStatus contracts.ScanStatus
}

// NewArgDataProviderResponseLatency  Ctor for ArgDataProviderResponseLatencyMetric
func NewArgDataProviderResponseLatency(status contracts.ScanStatus) *ArgDataProviderResponseLatencyMetric {
	return &ArgDataProviderResponseLatencyMetric{
		scanStatus: status,
	}
}

func (m *ArgDataProviderResponseLatencyMetric) MetricName() string {
	return "ArgDataProviderResponseLatency"
}

func (m *ArgDataProviderResponseLatencyMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "ScanStatus", Value: string(m.scanStatus)},
	}
}
