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
	cacheClient cache.ICacheClient
	// ARGDataProviderConfiguration is configuration data for ARGDataProvider
	argDataProviderConfiguration *ARGDataProviderConfiguration
}

// ARGDataProviderConfiguration is configuration data for ARGDataProvider
type ARGDataProviderConfiguration struct {
	// CacheExpirationTimeUnscannedResults is the expiration time **in minutes** for unscanned results in the cache client
	CacheExpirationTimeUnscannedResults int
	// CacheExpirationTimeScannedResults is the expiration time **in hours** for scan results in the cache client
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
func NewARGDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argClient IARGClient, queryGenerator queries.IARGQueryGenerator, cacheClient cache.ICacheClient, configuration *ARGDataProviderConfiguration) *ARGDataProvider {
	return &ARGDataProvider{
		tracerProvider:    instrumentationProvider.GetTracerProvider("ARGDataProvider"),
		metricSubmitter:   instrumentationProvider.GetMetricSubmitter(),
		argQueryGenerator: queryGenerator,
		argClient:         argClient,
		cacheClient: cache.NewSafeCacheClient(cacheClient),
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

	// Try to get results from cache. If an key dosen't exist or an error occurred continue without cache
	scanStatus, scanFindings, err := provider.getResultsFromCache(digest)
	if err == nil{ // Key exist in cache
		tracer.Info("got ImageVulnerabilityScanResults from cache")
		return scanStatus, scanFindings, nil
	}

	// Try to get results from ARG
	scanStatus, scanFindings, err = provider.getResultsFromArg(registry, repository, digest)
	if err != nil {
		err = errors.Wrap(err, "Failed to get get results from Arg")
		tracer.Error(err, "")
		return "", nil, err
	}
	tracer.Info("got results from Arg")

	// Set scan findings in cache
	err = provider.setScanFindingsInCache(scanFindings, scanStatus, digest)
	if err != nil { // in case error occurred - continue without cache
		err = errors.Wrap(err, "Failed to set scan findings in cache")
		tracer.Error(err, "")
	}

	return scanStatus, scanFindings, nil
}

// getResultsFromCache try to get ImageVulnerabilityScanResults from cache.
// The cache mapping digest to scan results or to known errors.
// If the digest exist in cache - return the value (scan results or error) and a flag _gotResultsFromCache
// If the digest dont exist in cache or any other unknown error occurred - return "", nil, nil and _didntGotResultsFromCache
func (provider *ARGDataProvider) getResultsFromCache(digest string) (contracts.ScanStatus, []*contracts.ScanFinding, error){
	tracer := provider.tracerProvider.GetTracer("getResultsFromCache")

	scanFindingsString, err := provider.cacheClient.Get(digest)

	// Nothing found in cache for digest as key
	if err != nil{ // Error as a result of key doesn't exist or other error from the cache functionality are treated the same (skip cache)
		err = errors.Wrap(err, "scanFindings as value don't exist in cache or there is an error in cache functionality")
		tracer.Error(err, "")
		return  "", nil, err
	}

	// Key exist in cache
	scanStatusFromCache, scanFindingsFromCache , unmarshalErr := provider.parseScanFindingsFromCache(scanFindingsString)
	if unmarshalErr != nil{ // json.unmarshall failed - trace the error and continue without cache
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on unmarshall scan results from cache")
		tracer.Error(unmarshalErr, "")
		return  "", nil, err
	}

	// results successfully extracted from cache - return the results
	tracer.Info("scanFindings exist in cache", "digest", digest)
	return scanStatusFromCache, scanFindingsFromCache, nil
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
		return "", nil, err
	}

	tracer.Info("Query", "Query", query)

	// Query arg for scan results for image
	results, err := provider.argClient.QueryResources(query)
	if err != nil {
		err = errors.Wrap(err, "Failed on argClient.QueryResources")
		tracer.Error(err, "")
		return "", nil, err
	}

	// Parse ARG client generic results to scan results ARG query array
	scanResultsQueryResponseObjectList, err := provider.parseARGImageScanResults(results)
	if err != nil {
		err = errors.Wrap(err, "Failed on parseARGImageScanResults")
		tracer.Error(err, "")
		return "", nil, err
	}
	tracer.Info("scanResultsQueryResponseObjectList", "list", scanResultsQueryResponseObjectList)

	// Get image scan data from the ARG query parsed results
	scanStatus, scanFindings, err := provider.getImageScanDataFromARGQueryScanResult(scanResultsQueryResponseObjectList)
	if err != nil {
		err = errors.Wrap(err, "Failed on getImageScanDataFromARGQueryScanResult")
		tracer.Error(err, "")
		return "", nil, err
	}
	return scanStatus, scanFindings, nil
}

// setScanFindingsInCache map digest to scan results
func (provider *ARGDataProvider) setScanFindingsInCache(scanFindings []*contracts.ScanFinding, scanStatus contracts.ScanStatus, digest string) error {
	tracer := provider.tracerProvider.GetTracer("setScanFindingsInCache")

	scanFindingsWrapper := &ScanFindingsInCache{ScanStatus: scanStatus, ScanFindings: scanFindings}
	scanFindingsBuffer, err := json.Marshal(scanFindingsWrapper)
	if err != nil {
		err = errors.Wrap(err, "Failed on json.Marshal scanFindingsWrapper")
		tracer.Error(err, "")
		return err
	}

	scanFindingsString := string(scanFindingsBuffer)
	expirationTime := provider.argDataProviderConfiguration.GetCacheExpirationTimeScannedResults()
	if scanStatus == contracts.Unscanned{
		expirationTime =  provider.argDataProviderConfiguration.GetCacheExpirationTimeUnscannedResults()
	}
	err = provider.cacheClient.Set(digest, scanFindingsString, expirationTime)
	if err != nil{
		err = errors.Wrap(err, "Failed to set digest in cache")
		tracer.Error(err, "")
		return err
	}

	tracer.Info("set scanFindings in cache", "digest", digest)
	return nil
}

// getScanFindingsFromCache get the scan results of the given digest from cache
// return keyNotFound err if key don't exist
func (provider *ARGDataProvider) parseScanFindingsFromCache(scanFindingsString string) (contracts.ScanStatus, []*contracts.ScanFinding, error) {
	tracer := provider.tracerProvider.GetTracer("parseScanFindingsFromCache")

	scanFindingsFromCache :=  new(ScanFindingsInCache)
	unmarshalErr := json.Unmarshal([]byte(scanFindingsString), scanFindingsFromCache)
	if unmarshalErr != nil {
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on json.Unmarshal scanFindingsWrapper")
		tracer.Error(unmarshalErr, "")
		return "", nil, unmarshalErr
	}
	return scanFindingsFromCache.ScanStatus, scanFindingsFromCache.ScanFindings, nil
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
		err = errors.Wrap(err, "Failed on json.Marshal results")
		tracer.Error(err, "")
		return nil, err
	}

	// Unmarshal to scan query results object
	containerVulnerabilityScanResultsQueryResponseObjectList := []*queries.ContainerVulnerabilityScanResultsQueryResponseObject{}
	err = json.Unmarshal(marshaled, &containerVulnerabilityScanResultsQueryResponseObjectList)
	if err != nil {
		err = errors.Wrap(err, "Failed on json.Unmarshal results")
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

// GetCacheExpirationTimeUnscannedResults uses ARGDataProviderConfiguration instance's CacheExpirationTimeUnscannedResults (int)
// to a return a time.Duration object
func (configuration *ARGDataProviderConfiguration) GetCacheExpirationTimeUnscannedResults() time.Duration {
	return time.Duration(configuration.CacheExpirationTimeUnscannedResults) * time.Minute
}

// GetCacheExpirationTimeScannedResults uses ARGDataProviderConfiguration instance's CacheExpirationTimeScannedResults (int)
// to a return a time.Duration object
func (configuration *ARGDataProviderConfiguration) GetCacheExpirationTimeScannedResults() time.Duration {
	return time.Duration(configuration.CacheExpirationTimeScannedResults) * time.Hour
}

