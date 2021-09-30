package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/operations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
)

// CacheClientGetMetric implements metric.IMetric interface
var _ metric.IMetric = (*CacheClientGetMetric)(nil)

// CacheClientGetMetric is metric of ARGDataProvider that checks the latency according to some scanStatus
type CacheClientGetMetric struct {
	// cacheType is the type of the cache - e.g. FreeCache/ redisCache ...
	cacheType string

	// operationStatus is the status of the operation, e.g. hit or miss.
	operationStatus operations.OPERATION_STATUS
}

// NewCacheOperationMetric  Ctor for ArgDataProviderResponseLatencyMetric
func NewCacheOperationMetric(cacheType interface{}, operationStatus operations.OPERATION_STATUS) *CacheClientGetMetric {
	return &CacheClientGetMetric{
		cacheType:       utils.GetType(cacheType),
		operationStatus: operationStatus,
	}
}

func (m *CacheClientGetMetric) MetricName() string {
	return "CacheClientGet"
}

func (m *CacheClientGetMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "CacheType", Value: m.cacheType},
		{Key: "OperationStatus", Value: string(m.operationStatus)},
	}
}
