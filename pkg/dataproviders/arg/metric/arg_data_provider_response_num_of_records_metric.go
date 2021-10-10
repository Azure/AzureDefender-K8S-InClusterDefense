package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// ArgDataProviderResponseNumOfRecordsMetric implements metric.IMetric interface
var _ metric.IMetric = (*ArgDataProviderResponseNumOfRecordsMetric)(nil)

// ArgDataProviderResponseNumOfRecordsMetric is metric that counts how many records returned
type ArgDataProviderResponseNumOfRecordsMetric struct {
}

func NewArgDataProviderResponseNumOfRecordsMetric() *ArgDataProviderResponseNumOfRecordsMetric {
	return &ArgDataProviderResponseNumOfRecordsMetric{}
}

func (m *ArgDataProviderResponseNumOfRecordsMetric) MetricName() string {
	return "ArgDataProviderResponseNumOfRecords"
}

func (m *ArgDataProviderResponseNumOfRecordsMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
