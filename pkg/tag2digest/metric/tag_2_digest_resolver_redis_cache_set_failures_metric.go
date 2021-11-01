package metric

import "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"

// Tag2DigestRedisCacheFailuresMetric implements metric.IMetric interface
var _ metric.IMetric = (*Tag2DigestRedisCacheFailuresMetric)(nil)

// Tag2DigestRedisCacheFailuresMetric is metric that counts how many failures of set a key in redis cache occurred.
type Tag2DigestRedisCacheFailuresMetric struct {
}

func NewTag2DigestRedisCacheFailuresMetric() *Tag2DigestRedisCacheFailuresMetric {
	return &Tag2DigestRedisCacheFailuresMetric{}
}

func (m *Tag2DigestRedisCacheFailuresMetric) MetricName() string {
	return "Tag2DigestRedisCacheFailuresMetric"
}

func (m *Tag2DigestRedisCacheFailuresMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}