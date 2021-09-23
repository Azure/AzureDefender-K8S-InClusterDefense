package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/operations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
)

// CacheOperationMetric implements metric.IMetric interface
var _ metric.IMetric = (*CacheOperationMetric)(nil)

// CacheOperationMetric is metric of ARGDataProvider that checks the latency according to some scanStatus
type CacheOperationMetric struct {
	// cacheType is the type of the cache - e.g. FreeCache/ redisCache ...
	cacheType string

	// operation is the operation that was executed - get or set.
	operation operations.OPERATION

	// operationStatus is the status of the operation, e.g. hit or miss.
	operationStatus operations.OPERATION_STATUS
}

// NewCacheOperationMetric  Ctor for ArgDataProviderResponseLatencyMetric
func NewCacheOperationMetric(cacheType interface{}, operation operations.OPERATION, operationStatus operations.OPERATION_STATUS) *CacheOperationMetric {
	return &CacheOperationMetric{
		cacheType:       utils.GetType(cacheType),
		operation:       operation,
		operationStatus: operationStatus,
	}
}

func (m *CacheOperationMetric) MetricName() string {
	return "CacheOperation"
}

func (m *CacheOperationMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "CacheType", Value: m.cacheType},
		{Key: "Operation", Value: string(m.operation)},
		{Key: "OperationStatus", Value: string(m.operationStatus)},
	}
}
