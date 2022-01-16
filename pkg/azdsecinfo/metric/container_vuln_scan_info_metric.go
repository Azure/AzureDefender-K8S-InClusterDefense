package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// ContainerVulnScanInfoMetric implements metric.IMetric interface
var _ metric.IMetric = (*ContainerVulnScanInfoMetric)(nil)

const _notApplicable = "N\\A"

// ContainerVulnScanInfoMetric is metric of AzDSecInfo to report per contianer info vuln scan returned
type ContainerVulnScanInfoMetric struct {
	// Scan Status
	scanStatus contracts.ScanStatus
	// Reason additional info required - for unscanned reason or other if scanned
	reason string
}

// NewContainerVulnScanInfoMetric  Ctor for ArgDataProviderResponseLatencyMetric
func NewContainerVulnScanInfoMetric(scanStatus contracts.ScanStatus) *ContainerVulnScanInfoMetric {
	return &ContainerVulnScanInfoMetric{
		scanStatus: scanStatus,
		reason:     _notApplicable,
	}
}

func NewContainerVulnScanInfoMetricWithUnscannedReason(scanStatus contracts.ScanStatus, unscannedReason contracts.UnscannedReason) *ContainerVulnScanInfoMetric {
	reason := string(unscannedReason)
	if reason == "" {
		reason = _notApplicable
	}

	return &ContainerVulnScanInfoMetric{
		scanStatus: scanStatus,
		reason:     reason,
	}
}

func (m *ContainerVulnScanInfoMetric) MetricName() string {
	return "ContainerVulnScanInfo"
}

func (m *ContainerVulnScanInfoMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "ScanStatus", Value: string(m.scanStatus)},
		{Key: "Reason", Value: m.reason},
	}
}
