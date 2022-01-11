package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/operations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
)

// CacheClientGetMetric implements metric.IMetric interface
var _ metric.IMetric = (*CacheClientSetMetric)(nil)

// CacheClientSetMetric is metric of ARGDataProvider that checks the latency according to some scanStatus
type CacheClientSetMetric struct {
	// cacheType is the type of the cache - e.g. FreeCache/ redisCache ...
	cacheType string

	// operationStatus is the status of the operation.
	operationStatus operations.OPERATION_STATUS
}

// NewCacheClientSetMetric  Ctor for CacheClientSetMetric
func NewCacheClientSetMetric(cacheType interface{}, operationStatus operations.OPERATION_STATUS) *CacheClientSetMetric {
	return &CacheClientSetMetric{
		cacheType:       utils.GetType(cacheType),
		operationStatus: operationStatus,
	}
}

func (m *CacheClientSetMetric) MetricName() string {
	return "CacheClientGet"
}

func (m *CacheClientSetMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "CacheType", Value: m.cacheType},
		{Key: "OperationStatus", Value: string(m.operationStatus)},
	}
}
