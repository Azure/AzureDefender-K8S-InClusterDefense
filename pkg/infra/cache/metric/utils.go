package metric

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
)

// NewSetErrEncounteredMetric returns *util.ErrorEncounteredMetric when Set operation failed of some ICacheClient failed.
func NewSetErrEncounteredMetric(err error, clientType string) *util.ErrorEncounteredMetric {
	errContext := clientType + "SetFailed"
	return util.NewErrorEncounteredMetric(err, errContext)
}

// NewGetErrEncounteredMetric returns *util.ErrorEncounteredMetric when Get operation failed of some ICacheClient failed.
func NewGetErrEncounteredMetric(err error, clientType string) *util.ErrorEncounteredMetric {
	errContext := clientType + "GetFailed"
	return util.NewErrorEncounteredMetric(err, errContext)
}
