package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// ArgDataProviderResponseLatencyMetric implements metric.IMetric interface
var _ metric.IMetric = (*ArgDataProviderResponseLatencyMetric)(nil)

// ArgDataProviderResponseLatencyMetric is metric of ARGDataProvider that checks the latency according to some scanStatus
type ArgDataProviderResponseLatencyMetric struct {
	// scanStatus is the scan status of the metric that would be deployed as ScanStatus metric.Dimension.
	scanStatus contracts.ScanStatus

	// queryName is the query name - e.g. GetImageVulnerabilityScanResults
	queryName string
}

// NewArgDataProviderResponseLatencyMetric  Ctor for ArgDataProviderResponseLatencyMetric
func NewArgDataProviderResponseLatencyMetric(status contracts.ScanStatus, queryName string) *ArgDataProviderResponseLatencyMetric {
	return &ArgDataProviderResponseLatencyMetric{
		scanStatus: status,
		queryName:  queryName,
	}
}

func NewArgDataProviderResponseLatencyMetricWithGetImageVulnerabilityScanResultsQuery(status contracts.ScanStatus) *ArgDataProviderResponseLatencyMetric {
	return NewArgDataProviderResponseLatencyMetric(status, "GetImageVulnerabilityScanResults")
}

func (m *ArgDataProviderResponseLatencyMetric) MetricName() string {
	return "ArgDataProviderResponseLatency"
}

func (m *ArgDataProviderResponseLatencyMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "ScanStatus", Value: string(m.scanStatus)},
	}
}
