package arg

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	argmetric "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	_argScanHealthyStatus = "Healthy"
)

// IARGDataProvider is a provider for any ARG data
type IARGDataProvider interface {
	// GetImageVulnerabilityScanResults fetch ARG based scan data information on image if exists from ARG
	// scanStatus to represent it stores a scan on image, and if so if it's healthy or not
	// If scanStatus is Unscanned, nil scan findings array
	// If scan status is Healthy, empty scan findings array
	// If scan status is Unhealthy, findings presented in scan findings array
	GetImageVulnerabilityScanResults(registry string, repository string, digest string) (scanStatus contracts.ScanStatus, scanFindings []*contracts.ScanFinding, err error)
}

// ARGDataProvider is a IARGDataProvider implementation
type ARGDataProvider struct {
	//tracerProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter
	metricSubmitter metric.IMetricSubmitter
	// argQueryGenerator is the generator for the are queries.
	argQueryGenerator *queries.ARGQueryGenerator
	// argClient is the arg client of the ARGDataProvider
	argClient IARGClient
}

// NewARGDataProvider Constructor
func NewARGDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argClient IARGClient, queryGenerator *queries.ARGQueryGenerator) *ARGDataProvider {
	return &ARGDataProvider{
		tracerProvider:    instrumentationProvider.GetTracerProvider("NewARGDataProvider"),
		metricSubmitter:   instrumentationProvider.GetMetricSubmitter(),
		argQueryGenerator: queryGenerator,
		argClient:         argClient,
	}
}

// GetImageVulnerabilityScanResults fetch ARG based scan data information on image if exists from ARG
// scanStatus to represent it stores a scan on image, and if so if it's healthy or not
// If scanStatus is Unscanned, nil scan findings array
// If scan status is Healthy, empty scan findings array
// If scan status is Unhealthy, findings presented in scan findings array
func (provider *ARGDataProvider) GetImageVulnerabilityScanResults(registry string, repository string, digest string) (contracts.ScanStatus, []*contracts.ScanFinding, error) {
	tracer := provider.tracerProvider.GetTracer("GetImageVulnerabilityScanResults")

	tracer.Info("Received", "registry", registry, "repository", repository, "digest", digest)

	// Generate image scan result ARG query for this specific image
	query, err := provider.argQueryGenerator.GenerateImageVulnerabilityScanQuery(&queries.ContainerVulnerabilityScanResultsQueryParameters{
		Registry:   registry,
		Repository: repository,
		Digest:     digest,
	})

	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.GetImageVulnerabilityScanResults failed on argQueryGenerator.GenerateImageVulnerabilityScanQuery")
		tracer.Error(err, "")
		return "", nil, err
	}

	tracer.Info("Query", "Query", query)

	// Query arg for scan results for image
	results, err := provider.argClient.QueryResources(query)
	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.GetImageVulnerabilityScanResults failed on argClient.QueryResources")
		tracer.Error(err, "")
		return "", nil, err
	}

	// Parse ARG client generic results to scan results ARG query array
	scanResultsQueryResponseObjectList, err := provider.parseARGImageScanResults(results)
	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.GetImageVulnerabilityScanResults failed on parseARGImageScanResults")
		tracer.Error(err, "")
		return "", nil, err
	}
	tracer.Info("scanResultsQueryResponseObjectList", "list", scanResultsQueryResponseObjectList)

	// Get image scan data from the ARG query parsed results
	scanStatus, scanFindings, err := provider.getImageScanDataFromARGQueryScanResult(scanResultsQueryResponseObjectList)
	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.GetImageVulnerabilityScanResults failed on getImageScanDataFromARGQueryScanResult")
		tracer.Error(err, "")
		return "", nil, err
	}

	return scanStatus, scanFindings, nil
}

// parseARGImageScanResults parse ARG client returnes results from scan results query to an array of ContainerVulnerabilityScanResultsQueryResponseObject
func (provider *ARGDataProvider) parseARGImageScanResults(argImageScanResults []interface{}) ([]*queries.ContainerVulnerabilityScanResultsQueryResponseObject, error) {
	tracer := provider.tracerProvider.GetTracer("parseARGImageScanResults")

	if argImageScanResults == nil {
		err := errors.Wrap(errors.New("Received results nil argument"), "ARGDataProvider.parseARGImageScanResults")
		tracer.Error(err, "")
		return nil, err
	}

	// TODO check if there is more efficient way for this - might be a performance hit..(maybe with other client return value)
	marshaled, err := json.Marshal(argImageScanResults)
	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.parseARGImageScanResults failed on json.Marshal results")
		tracer.Error(err, "")
		return nil, err
	}

	// Unmarshal to scan query results object
	containerVulnerabilityScanResultsQueryResponseObjectList := []*queries.ContainerVulnerabilityScanResultsQueryResponseObject{}
	err = json.Unmarshal(marshaled, &containerVulnerabilityScanResultsQueryResponseObjectList)
	if err != nil {
		err = errors.Wrap(err, "ARGDataProvider.parseARGImageScanResults failed on json.Unmarshal results")
		tracer.Error(err, "")
		return nil, err
	}

	return containerVulnerabilityScanResultsQueryResponseObjectList, nil
}

// getImageScanDataFromARGQueryScanResult build and analayze scan status and scan findings list from arg parsed reuslts of scan vulnerability ARG query
// scanStatus to represent it stores a scan on image, and if so if it's healthy or not
// If scanStatus is Unscanned, nil scan findings array
// If scan status is Healthy, empty scan findings array
// If scan status is Unhealthy, findings presented in scan findings array
func (provider *ARGDataProvider) getImageScanDataFromARGQueryScanResult(scanResultsQueryResponseObjectList []*queries.ContainerVulnerabilityScanResultsQueryResponseObject) (contracts.ScanStatus, []*contracts.ScanFinding, error) {
	tracer := provider.tracerProvider.GetTracer("getImageScanDataFromARGQueryScanResult")
	startTime := time.Now().UTC()
	if scanResultsQueryResponseObjectList == nil {
		err := errors.Wrap(errors.New("Received results nil argument"), "ARGDataProvider.getImageScanDataFromARGQueryScanResult")
		tracer.Error(err, "")
		return "", nil, err
	}

	// Should set the scanStatus to unscanned?
	if len(scanResultsQueryResponseObjectList) == 0 {
		// Unscanned - no results found
		tracer.Info("Set to Unscanned scan data")
		provider.metricSubmitter.SendMetric(int(time.Now().Sub(startTime).Nanoseconds()), argmetric.NewArgDataProviderResponseLatency(contracts.Unscanned))
		// Return unscanned and return nil array of findings
		return contracts.Unscanned, nil, nil
	}

	// Should set the scanStatus to healthy?
	if len(scanResultsQueryResponseObjectList) == 1 && strings.EqualFold(scanResultsQueryResponseObjectList[0].ScanStatus, _argScanHealthyStatus) {
		// Healthy Set to healthy scan
		tracer.Info("Set to Healthy scan data", "healthyReceivedFindings", scanResultsQueryResponseObjectList)

		provider.metricSubmitter.SendMetric(int(time.Now().Sub(startTime).Nanoseconds()), argmetric.NewArgDataProviderResponseLatency(contracts.HealthyScan))
		// Return healthy scan status and empty array (initialized but empty)
		return contracts.HealthyScan, []*contracts.ScanFinding{}, nil
	}

	// Set the scanStatus to Unhealthy.
	tracer.Info("Set to Unhealthy scan data")
	scanFindings := make([]*contracts.ScanFinding, 0, len(scanResultsQueryResponseObjectList))
	for _, element := range scanResultsQueryResponseObjectList {
		scanFindings = append(scanFindings, &contracts.ScanFinding{
			Id:        element.Id,
			Patchable: element.Patchable,
			Severity:  element.ScanFindingSeverity})
	}
	// Send metrics
	provider.metricSubmitter.SendMetric(int(time.Now().Sub(startTime).Nanoseconds()), argmetric.NewArgDataProviderResponseLatency(contracts.UnhealthyScan))
	provider.metricSubmitter.SendMetric(len(scanFindings), argmetric.NewArgDataProviderResponseNumOfRecordsMetric())
	return contracts.UnhealthyScan, scanFindings, nil
}
