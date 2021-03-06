package arg

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	argmetric "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
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

// ARGDataProvider implements IARGDataProvider interface
var _ IARGDataProvider = (*ARGDataProvider)(nil)

// ARGDataProvider is a IARGDataProvider implementation
type ARGDataProvider struct {
	//tracerProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter
	metricSubmitter metric.IMetricSubmitter
	// argQueryGenerator is the generator for the are queries.
	argQueryGenerator queries.IARGQueryGenerator
	// argClient is the arg client of the ARGDataProvider
	argClient IARGClient
	// cacheClient is a cache for mapping digest to scan results
	cacheClient IARGDataProviderCacheClient
	// ARGDataProviderConfiguration is configuration data for ARGDataProvider
	argDataProviderConfiguration *ARGDataProviderConfiguration
}

// ARGDataProviderConfiguration is configuration data for ARGDataProvider
type ARGDataProviderConfiguration struct {
	// CacheExpirationTimeUnscannedResults is the expiration time **IN MINUTES** for unscanned results in the cache client
	CacheExpirationTimeUnscannedResults int
	// CacheExpirationTimeScannedResults is the expiration time **IN HOURS** for scan results in the cache client
	CacheExpirationTimeScannedResults int
}

// ScanFindingsInCache represents findings of image vulnerability scan with its scan status
type ScanFindingsInCache struct {
	//ScanStatus vulnerability scan status for image
	ScanStatus contracts.ScanStatus `json:"scanStatus"`
	// ScanFindings vulnerability scan findings for image
	ScanFindings []*contracts.ScanFinding `json:"scanFindings"`
}

// NewARGDataProvider Constructor
func NewARGDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argClient IARGClient, queryGenerator queries.IARGQueryGenerator, cacheClient IARGDataProviderCacheClient, configuration *ARGDataProviderConfiguration) *ARGDataProvider {
	return &ARGDataProvider{
		tracerProvider:               instrumentationProvider.GetTracerProvider("ARGDataProvider"),
		metricSubmitter:              instrumentationProvider.GetMetricSubmitter(),
		argQueryGenerator:            queryGenerator,
		argClient:                    argClient,
		cacheClient:                  cacheClient,
		argDataProviderConfiguration: configuration,
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

	// Try to get results from cache. If a key doesn't exist or an error occurred - continue without cache
	scanStatus, scanFindings, err := provider.cacheClient.GetResultsFromCache(digest)
	if err != nil { // Couldn't get ImageVulnerabilityScanResults from cache - skip and get results from provider
		if cache.IsMissingKeyCacheError(err){
			tracer.Info("Missin key. Couldn't get ImageVulnerabilityScanResults from cache: Digest not in cache", "digest", digest)
		}else{
			err = errors.Wrap(err, "Couldn't get ImageVulnerabilityScanResults from cache: error encountered")
			tracer.Error(err, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.GetImageVulnerabilityScanResults"))
		}
	} else { //  Key exist in cache
		tracer.Info("got ImageVulnerabilityScanResults from cache")
		return scanStatus, scanFindings, nil
	}

	// Try to get results from ARG
	scanStatus, scanFindings, err = provider.getResultsFromArg(registry, repository, digest)
	if err != nil {
		err = errors.Wrap(err, "Failed to get get results from Arg")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.GetImageVulnerabilityScanResults"))
		return "", nil, err
	}
	tracer.Info("got results from Arg")

	// Set scan findings in cache
	// In case error occurred - continue without cache
	go provider.cacheClient.SetScanFindingsInCache(scanFindings, scanStatus, digest)

	return scanStatus, scanFindings, nil
}

// getResultsFromArg gets scan results from arg
func (provider *ARGDataProvider) getResultsFromArg(registry string, repository string, digest string) (contracts.ScanStatus, []*contracts.ScanFinding, error) {
	tracer := provider.tracerProvider.GetTracer("getResultsFromArg")

	// Generate image scan result ARG query for this specific image
	query, err := provider.argQueryGenerator.GenerateImageVulnerabilityScanQuery(&queries.ContainerVulnerabilityScanResultsQueryParameters{
		Registry:   registry,
		Repository: repository,
		Digest:     digest,
	})

	if err != nil {
		err = errors.Wrap(err, "Failed on argQueryGenerator.GenerateImageVulnerabilityScanQuery")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.getResultsFromArg"))
		return "", nil, err
	}

	tracer.Info("Query", "Query", query)

	// Query arg for scan results for image
	results, err := provider.argClient.QueryResources(query)
	if err != nil {
		err = errors.Wrap(err, "Failed on argClient.QueryResources")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.getResultsFromArg"))
		return "", nil, err
	}

	// Parse ARG client generic results to scan results ARG query array
	scanResultsQueryResponseObjectList, err := provider.parseARGImageScanResults(results)
	if err != nil {
		err = errors.Wrap(err, "Failed on parseARGImageScanResults")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.getResultsFromArg"))
		return "", nil, err
	}
	tracer.Info("scanResultsQueryResponseObjectList", "list", scanResultsQueryResponseObjectList)

	// Get image scan data from the ARG query parsed results
	scanStatus, scanFindings, err := provider.getImageScanDataFromARGQueryScanResult(scanResultsQueryResponseObjectList)
	if err != nil {
		err = errors.Wrap(err, "Failed on getImageScanDataFromARGQueryScanResult")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.getResultsFromArg"))
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
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.parseARGImageScanResults"))
		return nil, err
	}

	// TODO check if there is more efficient way for this - might be a performance hit..(maybe with other client return value)
	marshaled, err := json.Marshal(argImageScanResults)
	if err != nil {
		err = errors.Wrap(err, "Failed on json.Marshal results")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.parseARGImageScanResults"))
		return nil, err
	}

	// Unmarshal to scan query results object
	containerVulnerabilityScanResultsQueryResponseObjectList := []*queries.ContainerVulnerabilityScanResultsQueryResponseObject{}
	err = json.Unmarshal(marshaled, &containerVulnerabilityScanResultsQueryResponseObjectList)
	if err != nil {
		err = errors.Wrap(err, "Failed on json.Unmarshal results")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.parseARGImageScanResults"))
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
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProvider.getImageScanDataFromARGQueryScanResult"))
		return "", nil, err
	}

	// Should set the scanStatus to unscanned?
	if len(scanResultsQueryResponseObjectList) == 0 {
		// Unscanned - no results found
		tracer.Info("Set to Unscanned scan data")
		//TODO Check that this metric in the correct place (The value always 0, move this metric to another place).
		provider.metricSubmitter.SendMetric(util.GetDurationMilliseconds(startTime), argmetric.NewArgDataProviderResponseLatencyMetricWithGetImageVulnerabilityScanResultsQuery(contracts.Unscanned))
		// Return unscanned and return nil array of findings
		return contracts.Unscanned, nil, nil
	}

	// Should set the scanStatus to healthy?
	if len(scanResultsQueryResponseObjectList) == 1 && strings.EqualFold(scanResultsQueryResponseObjectList[0].ScanStatus, _argScanHealthyStatus) {
		// Healthy Set to healthy scan
		tracer.Info("Set to Healthy scan data", "healthyReceivedFindings", scanResultsQueryResponseObjectList)

		provider.metricSubmitter.SendMetric(util.GetDurationMilliseconds(startTime), argmetric.NewArgDataProviderResponseLatencyMetricWithGetImageVulnerabilityScanResultsQuery(contracts.HealthyScan))
		// Return healthy scan status and empty array (initialized but empty)
		return contracts.HealthyScan, []*contracts.ScanFinding{}, nil
	}

	// Set the scanStatus to Unhealthy.
	tracer.Info("Set to Unhealthy scan data")
	scanFindings := make([]*contracts.ScanFinding, 0, len(scanResultsQueryResponseObjectList))
	for _, element := range scanResultsQueryResponseObjectList {
		scanFindings = append(scanFindings, &contracts.ScanFinding{
			Id:        element.FindingsIds,
			Patchable: element.Patchable,
			Severity:  element.ScanFindingSeverity})
	}
	// Send metrics
	provider.metricSubmitter.SendMetric(len(scanFindings), argmetric.NewArgDataProviderResponseNumOfRecordsMetric())
	provider.metricSubmitter.SendMetric(util.GetDurationMilliseconds(startTime), argmetric.NewArgDataProviderResponseLatencyMetricWithGetImageVulnerabilityScanResultsQuery(contracts.UnhealthyScan))
	return contracts.UnhealthyScan, scanFindings, nil
}
