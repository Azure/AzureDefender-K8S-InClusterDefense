package metric

import "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"

// ArgDataProviderResponseNumOfRecordsMetric implements metric.IMetric interface
var _ metric.IMetric = (*ArgDataProviderRedisCacheFailuresMetric)(nil)

// ArgDataProviderRedisCacheFailuresMetric is metric that counts how many failures of set a key in redis cache occurred.
type ArgDataProviderRedisCacheFailuresMetric struct {
}

func NewArgDataProviderRedisCacheFailuresMetric() *ArgDataProviderRedisCacheFailuresMetric {
	return &ArgDataProviderRedisCacheFailuresMetric{}
}

func (m *ArgDataProviderRedisCacheFailuresMetric) MetricName() string {
	return "ArgDataProviderRedisCacheFailuresMetric"
}

func (m *ArgDataProviderRedisCacheFailuresMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
