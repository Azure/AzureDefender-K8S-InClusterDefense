package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// CacheClientGetMetric implements metric.IMetric interface
var _ metric.IMetric = (*CacheClientRedisPingMetric)(nil)

// CacheClientRedisPingMetric is metric of ARGDataProvider that checks the latency according to some scanStatus
type CacheClientRedisPingMetric struct {
}

// NewCacheClientRedisPingMetric  Ctor for cacheClientRedisPingMetric
func NewCacheClientRedisPingMetric() *CacheClientRedisPingMetric {
	return &CacheClientRedisPingMetric{}
}

func (m *CacheClientRedisPingMetric) MetricName() string {
	return "CacheClientRedisPing"
}

func (m *CacheClientRedisPingMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
