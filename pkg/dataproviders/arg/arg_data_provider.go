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
	registryerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	_argScanHealthyStatus = "Healthy"
	_gotResultsFromCache      = true
	_didntGotResultsFromCache = false
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
	// argDataProviderCacheClient is the cache ARGDataProvider uses and its time expiration
	argDataProviderCacheClient *ARGDataProviderCacheClient
}

// ARGDataProviderCacheClient is a cache client for ARGDataProvider that contains the needed cache client configurations
type ARGDataProviderCacheClient struct {
	// cacheClient is a cache for mapping digest to scan results
	cacheClient cache.ICacheClient
	// CacheExpirationTimeUnscannedResults is the expiration time **in ms** for unscanned results in the cache client
	CacheExpirationTimeUnscannedResults time.Duration
	// CacheExpirationTimeScannedResults is the expiration time **in ms** for scan results in the cache client
	CacheExpirationTimeScannedResults time.Duration
	// cacheExpirationTime is the expiration time in the cache client for errors occurred during GetImageVulnerabilityScanResults
	CacheExpirationTimeForErrors time.Duration
}

// ARGDataProviderConfiguration is configuration data for ARGDataProvider
type ARGDataProviderConfiguration struct {
	// CacheExpirationTimeUnscannedResults is the expiration time **in ms** for unscanned results in the cache client
	CacheExpirationTimeUnscannedResults string
	// CacheExpirationTimeScannedResults is the expiration time **in ms** for scan results in the cache client
	CacheExpirationTimeScannedResults string
	// cacheExpirationTime is the expiration time in the cache client for errors occurred during GetImageVulnerabilityScanResults
	CacheExpirationTimeForErrors string
}

// ScanFindingsInCache represents findings of image vulnerability scan with its scan status
type ScanFindingsInCache struct {
	//ScanStatus vulnerability scan status for image
	ScanStatus contracts.ScanStatus `json:"scanStatus"`
	// ScanFindings vulnerability scan findings for image
	ScanFindings []*contracts.ScanFinding `json:"scanFindings"`
}

// NewARGDataProvider Constructor
func NewARGDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argClient IARGClient, queryGenerator queries.IARGQueryGenerator, argDataProviderCacheClient *ARGDataProviderCacheClient) *ARGDataProvider {
	return &ARGDataProvider{
		tracerProvider:    instrumentationProvider.GetTracerProvider("ARGDataProvider"),
		metricSubmitter:   instrumentationProvider.GetMetricSubmitter(),
		argQueryGenerator: queryGenerator,
		argClient:         argClient,
		argDataProviderCacheClient: argDataProviderCacheClient,
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

	// Try to get results from cache
	// The results can be scanResults or error (if in previous run a know error occurred)
	// If a known error was saved in cache - return it
	gotResultsFromCache, scanStatus, scanFindings, err := provider.getResultsFromCache(digest)
	if gotResultsFromCache{
		if err != nil{ // a known error was saved in cache
			err = errors.Wrap(err, "Get error from cache as value")
			tracer.Error(err, "")
			return "", nil, err
		}
		tracer.Info("got ImageVulnerabilityScanResults from cache")
		return scanStatus, scanFindings, nil
	}

	// Try to get results from ARG
	scanStatus, scanFindings, err = provider.getResultsFromArg(registry, repository, digest)
	if err != nil {
		provider.setErrorInCache(digest, err)
		err = errors.Wrap(err, "Failed to get get results from Arg")
		tracer.Error(err, "")
		return "", nil, err
	}
	tracer.Info("got results from Arg")

	// Set scan findings in cache
	err = provider.setScanFindingsInCache(scanFindings, scanStatus, digest)
	if err != nil { // in case error occurred - continue without cache
		err = errors.Wrap(err, "Failed on getImageScanDataFromARGQueryScanResult")
		tracer.Error(err, "")
	}

	return scanStatus, scanFindings, nil
}

// getResultsFromCache try to get ImageVulnerabilityScanResults from cache.
// The cache mapping digest to scan results or to known errors.
// If the digest exist in cache - return the value (scan results or error) and a flag _gotResultsFromCache
// If the digest dont exist in cache or any other unknown error occurred - return "", nil, nil and _didntGotResultsFromCache
func (provider *ARGDataProvider) getResultsFromCache(digest string) (bool, contracts.ScanStatus, []*contracts.ScanFinding, error){
	tracer := provider.tracerProvider.GetTracer("getResultsFromCache")

	scanFindingsString, err := provider.argDataProviderCacheClient.cacheClient.Get(digest)

	// Nothing found in cache for digest as key
	if err != nil{ // Error as a result of key doesn't exist or other error from the cache functionality are treated the same (skip cache)
		tracer.Info("scanFindings don't exist in cache", "digest", digest)
		return _didntGotResultsFromCache, "", nil, nil
	}

	// digest exist in cache (as key) -> scanFindingsString exist in cache -> scanFindingsString is scan results or an error occurred during GetImageVulnerabilityScanResults
	err, isKnownError := registryerrors.TryParseStringToUnscannedWithReasonErr(scanFindingsString)
	if isKnownError{ // error found in cache as value
		err = errors.Wrap(err, "got an error value instead of scan results from cache")
		tracer.Error(err, "")
		return _gotResultsFromCache, "", nil, err
	}

	// scanFindingsString is scan results
	scanStatusFromCache, scanFindingsFromCache , unmarshalErr := provider.parseScanFindingsFromCache(scanFindingsString)
	if unmarshalErr != nil{ // json.unmarshall failed - trace the error and continue without cache
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on unmarshall scan results from cache")
		tracer.Error(unmarshalErr, "")
		return _didntGotResultsFromCache, "", nil, nil
	}

	// results successfully extracted from cache - return the results
	tracer.Info("scanFindings exist in cache", "digest", digest)
	return _gotResultsFromCache, scanStatusFromCache, scanFindingsFromCache, nil
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

// setErrorInCache map digest to a known error. If the given error is a known error, set it in cache.
func (provider *ARGDataProvider) setErrorInCache(digest string, err error){
	tracer := provider.tracerProvider.GetTracer("setErrorInCache")
	errorAsString, isErrParsedToUnscannedReason := registryerrors.TryParseErrToUnscannedWithReason(err)
	if !isErrParsedToUnscannedReason {
		err = errors.Wrap(err, "Unexpected error while trying to get results from ARGDataProvider - don't save in cache")
		tracer.Error(err, "")
		return
	}
	err = provider.argDataProviderCacheClient.cacheClient.Set(digest, string(*errorAsString), provider.argDataProviderCacheClient.CacheExpirationTimeForErrors)
	if err != nil{
		err = errors.Wrap(err, "Failed to set error in cache")
		tracer.Error(err, "")
	}
	tracer.Info("Set error in cache", "digest", digest, "error", errorAsString)
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
	expirationTime := provider.argDataProviderCacheClient.CacheExpirationTimeScannedResults
	if scanStatus == contracts.Unscanned{
		expirationTime =  provider.argDataProviderCacheClient.CacheExpirationTimeUnscannedResults
	}
	err = provider.argDataProviderCacheClient.cacheClient.Set(digest, scanFindingsString, expirationTime)
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
