package arg

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"time"
)

// IARGDataProviderCacheClient is a cache client designated for ARGDataProvider
type IARGDataProviderCacheClient interface {

	// GetResultsFromCache try to get ImageVulnerabilityScanResults from cache.
	// The cache mapping digest to scan results or to known errors.
	// If the digest exist in cache - return the value (scan results or error) and a flag _gotResultsFromCache
	// If the digest dont exist in cache or any other unknown error occurred - return "", nil, nil and _didntGotResultsFromCache
	GetResultsFromCache(digest string) (contracts.ScanStatus, []*contracts.ScanFinding, error)

	// SetScanFindingsInCache map digest to scan results
	SetScanFindingsInCache(scanFindings []*contracts.ScanFinding, scanStatus contracts.ScanStatus, digest string) error
}

// ARGDataProviderCacheClient implements IARGDataProviderCacheClient interface
var _ IARGDataProviderCacheClient = (*ARGDataProviderCacheClient)(nil)

// ARGDataProviderCacheClient is a cache client designated for ARGDataProvider
// It wraps ICache client
type ARGDataProviderCacheClient struct {
	//tracerProvider is tracer provider of ARGDataProviderCacheClient
	tracerProvider trace.ITracerProvider
	//metricSubmitter is metric submitter of ARGDataProviderCacheClient
	metricSubmitter metric.IMetricSubmitter
	// cacheClient is a cache for mapping digest to scan results and save timeout status
	cacheClient cache.ICacheClient
	// CacheExpirationTimeUnscannedResults is the expiration time **IN MINUTES** for unscanned results in the cache client
	cacheExpirationTimeUnscannedResults time.Duration
	// CacheExpirationTimeScannedResults is the expiration time **IN HOURS** for scan results in the cache client
	cacheExpirationTimeScannedResults time.Duration
}

// NewARGDataProviderCacheClient - ARGDataProviderCacheClient Ctor
func NewARGDataProviderCacheClient(instrumentationProvider instrumentation.IInstrumentationProvider, cacheClient cache.ICacheClient, argDataProviderConfiguration *ARGDataProviderConfiguration) *ARGDataProviderCacheClient {
	return &ARGDataProviderCacheClient{
		tracerProvider:                      instrumentationProvider.GetTracerProvider("ARGDataProviderCacheClient"),
		metricSubmitter:                     instrumentationProvider.GetMetricSubmitter(),
		cacheClient:                         cacheClient,
		cacheExpirationTimeUnscannedResults: utils.GetMinutes(argDataProviderConfiguration.CacheExpirationTimeUnscannedResults),
		cacheExpirationTimeScannedResults:   utils.GetHours(argDataProviderConfiguration.CacheExpirationTimeScannedResults),
	}
}

// GetResultsFromCache try to get ImageVulnerabilityScanResults from cache.
// The cache mapping digest to scan results or to known errors.
// If the digest exist in cache - return the value (scan results or error) and a flag _gotResultsFromCache
// If the digest dont exist in cache or any other unknown error occurred - return "", nil, nil and _didntGotResultsFromCache
func (client *ARGDataProviderCacheClient) GetResultsFromCache(digest string) (contracts.ScanStatus, []*contracts.ScanFinding, error) {
	tracer := client.tracerProvider.GetTracer("GetResultsFromCache")

	scanFindingsString, err := client.cacheClient.Get(digest)

	// Nothing found in cache for digest as key
	if err != nil { // Error as a result of key doesn't exist or other error from the cache functionality are treated the same (skip cache)
		if cache.IsMissingKeyCacheError(err) {
			tracer.Info("Missing key. Digest as key not in cache", "digest", digest)
			return "", nil, err
		}
		err = errors.Wrap(err, "scanFindings as value don't exist in cache or there is an error in cache functionality")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProviderCacheClient.GetResultsFromCache"))
		return "", nil, err
	}

	// Key exist in cache
	scanStatusFromCache, scanFindingsFromCache, unmarshalErr := client.parseScanFindingsFromCache(scanFindingsString)
	if unmarshalErr != nil { // json.unmarshall failed - trace the error and continue without cache
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on unmarshall scan results from cache")
		tracer.Error(unmarshalErr, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(unmarshalErr, "ARGDataProviderCacheClient.GetResultsFromCache"))
		return "", nil, unmarshalErr
	}

	// results successfully extracted from cache - return the results
	tracer.Info("scanFindings exist in cache", "digest", digest)
	return scanStatusFromCache, scanFindingsFromCache, nil
}

// SetScanFindingsInCache map digest to scan results
func (client *ARGDataProviderCacheClient) SetScanFindingsInCache(scanFindings []*contracts.ScanFinding, scanStatus contracts.ScanStatus, digest string) error {
	tracer := client.tracerProvider.GetTracer("SetScanFindingsInCache")

	// Convert results to string in order to set the results in the cache
	scanFindingsWrapper := &ScanFindingsInCache{ScanStatus: scanStatus, ScanFindings: scanFindings}
	scanFindingsBuffer, err := json.Marshal(scanFindingsWrapper)
	if err != nil {
		err = errors.Wrap(err, "Failed on json.Marshal scanFindingsWrapper")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProviderCacheClient.SetScanFindingsInCache"))
		return err
	}
	scanFindingsString := string(scanFindingsBuffer)

	// Set TTL. Different TTL for different scan status
	expirationTime := client.cacheExpirationTimeScannedResults // Default
	if scanStatus == contracts.Unscanned {
		expirationTime = client.cacheExpirationTimeUnscannedResults
	}

	// Set results in cache
	err = client.cacheClient.Set(digest, scanFindingsString, expirationTime)
	if err != nil {
		err = errors.Wrap(err, "Failed to set digest in cache")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGDataProviderCacheClient.SetScanFindingsInCache"))
		return err
	}

	tracer.Info("set scanFindings in cache successfully", "digest", digest)
	return nil
}

// parseScanFindingsFromCache parse scan results as string to contracts.ScanStatus and []*contracts.ScanFinding objects
func (client *ARGDataProviderCacheClient) parseScanFindingsFromCache(scanFindingsString string) (contracts.ScanStatus, []*contracts.ScanFinding, error) {
	tracer := client.tracerProvider.GetTracer("parseScanFindingsFromCache")

	scanFindingsFromCache := new(ScanFindingsInCache)
	unmarshalErr := json.Unmarshal([]byte(scanFindingsString), scanFindingsFromCache)
	if unmarshalErr != nil {
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on json.Unmarshal scanFindingsWrapper")
		tracer.Error(unmarshalErr, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(unmarshalErr, "ARGDataProviderCacheClient.parseScanFindingsFromCache"))
		return "", nil, unmarshalErr
	}
	return scanFindingsFromCache.ScanStatus, scanFindingsFromCache.ScanFindings, nil
}
