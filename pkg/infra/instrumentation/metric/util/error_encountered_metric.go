package util

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
)

// ErrorEncounteredMetric implements metric.IMetric interface
var _ metric.IMetric = (*ErrorEncounteredMetric)(nil)

// ErrorEncounteredMetric implementation of metric.IMetric, for error encountered metric
type ErrorEncounteredMetric struct {
	errorType string
	context   string
}

// NewErrorEncounteredMetric Cto'r for ErrorEncounteredMetric
func NewErrorEncounteredMetric(err error, context string) *ErrorEncounteredMetric {
	return &ErrorEncounteredMetric{
		errorType: utils.GetType(err),
		context:   context,
	}
}

func (m *ErrorEncounteredMetric) MetricName() string {
	return "ErrorEncountered"
}

func (m *ErrorEncounteredMetric) MetricDimension() []metric.Dimension {
	return []metric.Dimension{
		{Key: "ErrorType", Value: m.errorType},
		{Key: "Context", Value: m.context},
	}
}
