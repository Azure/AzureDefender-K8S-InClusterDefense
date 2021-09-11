package crane_metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
)

// CraneWrapperNumOfDigestRetryAttempts is metric that counts how many attempts digest method has been executed
type CraneWrapperNumOfDigestRetryAttempts struct {
}

func NewCraneWrapperNumOfRetryAttempts() *CraneWrapperNumOfDigestRetryAttempts {
	return &CraneWrapperNumOfDigestRetryAttempts{}
}

func (m *CraneWrapperNumOfDigestRetryAttempts) MetricName() string {
	return "CraneWrapperNumOfDigestRetryAttempts"
}

func (m *CraneWrapperNumOfDigestRetryAttempts) MetricDimension() []metric.Dimension {
	return []metric.Dimension{}
}
