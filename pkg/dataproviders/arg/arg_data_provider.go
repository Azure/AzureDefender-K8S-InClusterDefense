package arg

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
)

type IARGDataProvider interface {
	TryGetImageVulnerabilityScanResults(registry string, repository string, digest string) (isScanFound bool, scanFindings []*contracts.ScanFinding, err error)
}

type ARGDataProvider struct {
	tracerProvider    trace.ITracerProvider
	metricSubmitter   metric.IMetricSubmitter
	argQueryGenerator *queries.ARGQueryGenerator
	argClient         IARGClient
}

func NewARGDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argClient IARGClient, queryGenerator *queries.ARGQueryGenerator) *ARGDataProvider {
	return &ARGDataProvider{
		tracerProvider:    instrumentationProvider.GetTracerProvider("NewARGDataProvider"),
		metricSubmitter:   instrumentationProvider.GetMetricSubmitter(),
		argQueryGenerator: queryGenerator,
		argClient:         argClient,
	}
}

func (provider *ARGDataProvider) TryGetImageVulnerabilityScanResults(registry string, repository string, digest string) (isScanFound bool, scanFindings []*contracts.ScanFinding, err error) {
	tracer := provider.tracerProvider.GetTracer("TryGetImageVulnerabilityScanResults")
	isScanFound = false

	tracer.Info("Received", "registry", registry, "repository", repository, "digest", digest)

	query, err := provider.argQueryGenerator.GenerateImageVulnerabilityScanQuery(&queries.ContainerVulnerabilityScanResultsQueryParameters{
		Registry:   registry,
		Repository: repository,
		Digest:     digest,
	})

	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.TryGetImageVulnerabilityScanResults failed on argQueryGenerator.GenerateImageVulnerabilityScanQuery")
		tracer.Error(err, "")
		return false, nil, err
	}

	tracer.Info("Query", "Query", query)

	results, err := provider.argClient.QueryResources(query)
	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.TryGetImageVulnerabilityScanResults failed on argClient.QueryResources")
		tracer.Error(err, "")
		return false, nil, err
	}

	if len(results) > 0 {
		isScanFound = true
		// TODO check if there is more efficient way for this - might be a performance hit..(maybe with other client return value)
		marshaled, err := json.Marshal(results)
		if err != nil {
			err = errors.Wrap(err, "ARGDataProvider.TryGetImageVulnerabilityScanResults failed on json.Marshal results")
			tracer.Error(err, "")
			return false, nil, err
		}

		containerVulnerabilityScanResultsQueryResponseObjectList := []*queries.ContainerVulnerabilityScanResultsQueryResponseObject{}
		err = json.Unmarshal(marshaled, &containerVulnerabilityScanResultsQueryResponseObjectList)
		if err != nil {
			err = errors.Wrap(err, "ARGDataProvider.TryGetImageVulnerabilityScanResults failed on json.Unmarshal results")
			tracer.Error(err, "")
			return false, nil, err
		}

		for _, element := range containerVulnerabilityScanResultsQueryResponseObjectList {
			scanFindings = append(scanFindings, &contracts.ScanFinding{
				Id:        element.Id,
				Patchable: element.Patchable,
				Severity:  element.ScanFindingSeverity,
			})
		}

	} else {
		isScanFound = false
	}

	return isScanFound, scanFindings, nil
}

//
