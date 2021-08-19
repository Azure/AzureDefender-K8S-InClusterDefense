package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
)

type IARGDataProvider interface {
	TryGetImageVulnerabilityScanResults(registry string, repository string, digest string) (isScanFound bool,scanFindings []*contracts.ScanFinding, err error)
}

type ARGDataProvider struct {
	tracerProvider  trace.ITracerProvider
	metricSubmitter metric.IMetricSubmitter
	azureAuthorizerFactory azureauth.IAzureAuthorizerFactory
	argQueryGenerator *queries.ARGQueryGenerator
}

func NewARGDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider, azureAuthorizerFactory azureauth.IAzureAuthorizerFactory, queryGenerator *queries.ARGQueryGenerator) *ARGDataProvider{
	return &ARGDataProvider{
		tracerProvider:  instrumentationProvider.GetTracerProvider("NewARGDataProvider"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		azureAuthorizerFactory: azureAuthorizerFactory,
		argQueryGenerator: queryGenerator,

	}
}


func (provider *ARGDataProvider) TryGetImageVulnerabilityScanResults(registry string, repository string, digest string) (isScanFound bool,scanFindings []*contracts.ScanFinding, err error) {
	tracer := provider.tracerProvider.GetTracer("TryGetImageVulnerabilityScanResults")
	isScanFound = false





	return isScanFound, scanFindings, err

}