package azdsecinfo

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	//corev1 "k8s.io/api/core/v1"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// _timeoutPrefixForCacheKey is a prefix for PodSpecCacheKey for timeout keys. The prefix is used to separate timeout keys with containerVulnerabilityScanInfo keys
	_timeoutPrefixForCacheKey = "timeout"
	// _containerVulnerabilityScanInfoPrefixForCacheKey is a prefix for PodSpecCacheKey for containerVulnerabilityScanInfo keys. The prefix is used to separate timeout keys with containerVulnerabilityScanInfo keys
	_containerVulnerabilityScanInfoPrefixForCacheKey = "ContainerVulnerabilityScanInfo"
)

var (
	// _resetTimeoutTTL is the TTL for reseting timeout status in cache (after reset timestatus to noTimeoutEncountered the value should be in cache only for short period)
	_resetTimeoutTTL = time.Duration(1) // 1 nanosecond
)

// IAzdSecInfoProviderCacheClient cache client designated for AzdSecInfoProvider
type IAzdSecInfoProviderCacheClient interface {
	// GetContainerVulnerabilityScanInfofromCache try to get ContainerVulnerabilityScanInfo from cache.
	// It gets the results from the cache and parse it to containerVulnerabilityCacheResultsWrapper object.
	// If there is an error with the cache or the value is invalid returns an error.
	GetContainerVulnerabilityScanInfofromCache(podSpecCacheKey string) ([]*contracts.ContainerVulnerabilityScanInfo, error, error)

	// SetContainerVulnerabilityScanInfoInCache set ContainerVulnerabilityScanInfo in cache
	// No error is reported back, only tracing it
	SetContainerVulnerabilityScanInfoInCache(podSpecCacheKey string, containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) error

	// GetTimeOutStatus gets the timeout status of the podSpec from cache - how many times timeout has occurred for this podSpec
	GetTimeOutStatus(podSpecCacheKey string) (int, error)

	// SetTimeOutStatusAfterEncounteredTimeout update the timeout status if already exist in cache or set for the first time timeout status
	SetTimeOutStatusAfterEncounteredTimeout(podSpecCacheKey string, timeOutStatus int) error

	// ResetTimeOutInCacheAfterGettingScanResults resets the timeout status in cache after scanResults was received.
	// If scanResults was received the timeout is no longer relevant and needs to be reset.
	// If no timeout occurred before, do nothing.
	ResetTimeOutInCacheAfterGettingScanResults(podSpecCacheKey string) error

	// GetPodSpecCacheKey get the cache key without the prefix of a given podSpec
	GetPodSpecCacheKey(podSpec *admisionrequest.PodSpec) string
}

// AzdSecInfoProviderCacheClient implements IAzdSecInfoProviderCacheClient interface
var _ IAzdSecInfoProviderCacheClient = (*AzdSecInfoProviderCacheClient)(nil)

// AzdSecInfoProviderCacheClient cache client designated for AzdSecInfoProvider
// It wraps ICache client
type AzdSecInfoProviderCacheClient struct {
	//tracerProvider is tracer provider of AzdSecInfoProviderCacheClient
	tracerProvider trace.ITracerProvider
	//metricSubmitter is metric submitter of AzdSecInfoProviderCacheClient
	metricSubmitter metric.IMetricSubmitter
	// cacheClient is a cache for mapping digest to scan results and save timeout status
	cacheClient cache.ICacheClient
	// cacheExpirationTimeTimeout is the expiration time **IN MINUTES** for timout.
	cacheExpirationTimeTimeout time.Duration
	// cacheExpirationContainerVulnerabilityScanInfo is the expiration time **IN SECONDS** for ContainerVulnerabilityScanInfo.
	cacheExpirationContainerVulnerabilityScanInfo time.Duration
}

// containerVulnerabilityCacheResultsWrapper is a wrapper for ContainerVulnerabilityScanInfo
// It holds both data and error.
type containerVulnerabilityCacheResultsWrapper struct {
	// ContainerVulnerabilityScanInfo array of ContainerVulnerabilityScanInfo
	ContainerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo `json:"containerVulnerabilityScanInfo"`
	// err is an error occurred while getting ContainerVulnerabilityScanInfo
	ErrString string `json:"err"`
}

// NewAzdSecInfoProviderCacheClient - AzdSecInfoProviderCacheClient Ctor
func NewAzdSecInfoProviderCacheClient(instrumentationProvider instrumentation.IInstrumentationProvider, cacheClient cache.ICacheClient, azdSecInfoProviderConfiguration *AzdSecInfoProviderConfiguration) *AzdSecInfoProviderCacheClient {
	return &AzdSecInfoProviderCacheClient{
		tracerProvider:             instrumentationProvider.GetTracerProvider("AzdSecInfoProviderCacheClient"),
		metricSubmitter:            instrumentationProvider.GetMetricSubmitter(),
		cacheClient:                cacheClient,
		cacheExpirationTimeTimeout: utils.GetMinutes(azdSecInfoProviderConfiguration.CacheExpirationTimeTimeout),
		cacheExpirationContainerVulnerabilityScanInfo: utils.GetSeconds(azdSecInfoProviderConfiguration.CacheExpirationContainerVulnerabilityScanInfo),
	}
}

// GetContainerVulnerabilityScanInfofromCache try to get ContainerVulnerabilityScanInfo from cache.
// It gets the results from the cache and parse it to containerVulnerabilityCacheResultsWrapper object.
// Returns:
// []*contracts.ContainerVulnerabilityScanInfo - If scan results was STORED in cache as value from previous scans, otherwise nil
// error - If error was STORED in cache as value from previous scans, otherwise nil
//If there is an error with the cache or the value is invalid returns an error.
func (client *AzdSecInfoProviderCacheClient) GetContainerVulnerabilityScanInfofromCache(podSpecCacheKey string) ([]*contracts.ContainerVulnerabilityScanInfo, error, error) {
	tracer := client.tracerProvider.GetTracer("GetContainerVulnerabilityScanInfofromCache")
	// Get the key
	ContainerVulnerabilityScanInfoCacheKey := client.getContainerVulnerabilityScanInfoCacheKey(podSpecCacheKey)

	// get result from cache
	scanInfoWrapperStringFromCache, err := client.cacheClient.Get(ContainerVulnerabilityScanInfoCacheKey)
	if err != nil { // Key don't exist in cache or error with cache functionality
		// Check if the error is MissingKeyCacheError
		if cache.IsMissingKeyCacheError(err) { // The key not in the cache
			tracer.Info("Missing key. ContainerVulnerabilityScanInfoCacheKey is not in cache", "ContainerVulnerabilityScanInfoCacheKey", ContainerVulnerabilityScanInfoCacheKey)
			return nil, nil, err
		}
		// error with cache functionality
		err = errors.Wrap(err, "Failed to get ContainerVulnerabilityScanInfo from cache")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.GetContainerVulnerabilityScanInfofromCache"))
		return nil, nil, err
	}

	// Key exist in cache
	// Parse the results
	ContainersVulnerabilityScanInfoWrapperFromCache, err := client.unmarshalScanResults(scanInfoWrapperStringFromCache)
	if err != nil { // unmarshal failed
		err = errors.Wrap(err, "failed to unmarshalScanResults from cache")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.GetContainerVulnerabilityScanInfofromCache"))
		return nil, nil, err
	}

	// Get error stored in cache from previous runs
	errorStoredInCache := client.getErrorStoredInCache(ContainersVulnerabilityScanInfoWrapperFromCache.ErrString)
	// If there is an error stored in cache from previous runs
	if errorStoredInCache != nil {
		tracer.Info("Got error stored in cache", "errorStoredInCache", errorStoredInCache)
		return nil, errorStoredInCache, nil
	}

	// return results
	tracer.Info("Got ContainerVulnerabilityScanInfo from cache")
	return ContainersVulnerabilityScanInfoWrapperFromCache.ContainerVulnerabilityScanInfo, nil, nil
}

// SetContainerVulnerabilityScanInfoInCache set ContainerVulnerabilityScanInfo in cache
// No error is reported back, only tracing it
func (client *AzdSecInfoProviderCacheClient) SetContainerVulnerabilityScanInfoInCache(podSpecCacheKey string, containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) error {
	tracer := client.tracerProvider.GetTracer("SetContainerVulnerabilityScanInfoInCache")
	// Convert results to resultsString
	resultsString, err := client.marshalScanResults(containerVulnerabilityScanInfo, err)
	if err != nil { // Marshal failed
		err = errors.Wrap(err, "Failed to marshal ContainerVulnerabilityScanInfo")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.SetContainerVulnerabilityScanInfoInCache"))
		return err
	}
	// Try to set resultsString in cache
	if err = client.cacheClient.Set(client.getContainerVulnerabilityScanInfoCacheKey(podSpecCacheKey), resultsString, client.cacheExpirationContainerVulnerabilityScanInfo); err != nil {
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.SetContainerVulnerabilityScanInfoInCache"))
		return err
	} else {
		tracer.Info("Set ContainerVulnerabilityScanInfo in cache successfully")
	}
	return nil
}

// GetTimeOutStatus gets the timeout status of the podSpec from cache - how many times timeout has occurred for this podSpec
func (client *AzdSecInfoProviderCacheClient) GetTimeOutStatus(podSpecCacheKey string) (int, error) {
	tracer := client.tracerProvider.GetTracer("GetTimeOutStatus")

	// Get key for cache
	timeOutCacheKey := client.getTimeOutCacheKey(podSpecCacheKey)
	// Get timeoutStatus from cache
	timeoutStatusString, err := client.cacheClient.Get(timeOutCacheKey)
	// Key don't exist in cache or error with cache functionality
	if err != nil {
		// Check if the error is MissingKeyCacheError
		if cache.IsMissingKeyCacheError(err) { // The key not in the cache
			tracer.Info("Missing key. TimeOutCacheKey not in cache", "timeOutCacheKey", timeOutCacheKey)
			return _noTimeOutEncountered, nil
		}
		// error with cache functionality - _unknownTimeOutStatus.
		err = errors.Wrap(err, "Error while trying to get timeoutStatus from cache")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.GetTimeOutStatus"))
		return _unknownTimeOutStatus, err
	}

	// Make sure the string from the cache is a valid timeout status (int)
	timeoutStatus, err := strconv.Atoi(timeoutStatusString)
	if err != nil {
		err = errors.Wrapf(err, "Invalid value in cache for timeout status - should be valid int. got %s", timeoutStatusString)
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.GetTimeOutStatus"))
		return _unknownTimeOutStatus, err
	}
	// Valid int
	tracer.Info("timeout status is in cache", "timeoutStatus", timeoutStatus)
	return timeoutStatus, nil
}

// SetTimeOutStatusAfterEncounteredTimeout update the timeout status if already exist in cache or set for the first time timeout status
func (client *AzdSecInfoProviderCacheClient) SetTimeOutStatusAfterEncounteredTimeout(podSpecCacheKey string, timeOutStatus int) error {
	tracer := client.tracerProvider.GetTracer("SetTimeOutStatusAfterEncounteredTimeout")
	tracer.Info("Try to set timeOutStatus in cache", "timeOutStatus", timeOutStatus)
	// TODO handle race condition and locks for redis
	return client.setTimeOutStatus(podSpecCacheKey, timeOutStatus, client.cacheExpirationTimeTimeout)
}

// setTimeOutStatus set a given timeOutStatus in cache.
func (client *AzdSecInfoProviderCacheClient) setTimeOutStatus(podSpecCacheKey string, timeOutStatus int, expirationTime time.Duration) error {
	tracer := client.tracerProvider.GetTracer("setTimeOutStatus")

	// Get key for cache
	timeOutCacheKey := client.getTimeOutCacheKey(podSpecCacheKey)
	timeOutStatusString := strconv.Itoa(timeOutStatus)
	if err := client.cacheClient.Set(timeOutCacheKey, timeOutStatusString, expirationTime); err != nil {
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.SetTimeOutStatusAfterEncounteredTimeout"))
		return err
	}
	tracer.Info("Set timeout status in cache succeeded", "timeOutStatusString", timeOutStatusString)
	return nil
}

// ResetTimeOutInCacheAfterGettingScanResults resets the timeout status in cache after scanResults was received.
// If scanResults was received the timeout is no longer relevant and needs to be reset.
// If no timeout occurred before, do nothing.
func (client *AzdSecInfoProviderCacheClient) ResetTimeOutInCacheAfterGettingScanResults(podSpecCacheKey string) error {
	tracer := client.tracerProvider.GetTracer("ResetTimeOutInCacheAfterGettingScanResults")

	// Get key for cache
	timeOutCacheKey := client.getTimeOutCacheKey(podSpecCacheKey)
	// Check if the timeOutCacheKey is already in cache
	timeoutEncounteredString, err := client.cacheClient.Get(timeOutCacheKey)
	if err != nil {
		if 	cache.IsMissingKeyCacheError(err) {
			tracer.Info("Missing key. TimeOutCacheKey is not in cache", "timeOutCacheKey", timeOutCacheKey)
			return nil
		}
		// Error in cache functionality - return the error
		err = errors.Wrap(err, "error encountered while trying to get timeout status from cache.")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults"))
		return err
	}

	// Make sure the string from the cache is a valid timeout status (int)
	timeoutEncountered, err := strconv.Atoi(timeoutEncounteredString)
	if err != nil {
		err = errors.Wrapf(err, "Invalid value in cache for timeout status - should be valid int. got %s", timeoutEncounteredString)
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults"))
		return err
	}

	// In case that the podSpecCacheKey is already in cache.
	tracer.Info("timeout status exist in cache - reset timeout status", "podSpecCacheKey", podSpecCacheKey, "timeoutEncountered", timeoutEncountered)
	// In case value in cache indicate that _noTimeOutEncountered - no need to update the value in cache
	if timeoutEncountered == _noTimeOutEncountered {
		tracer.Info("No need to update timeOut status in cache because it was already set to no timeout encountered", "podSpecCacheKey", podSpecCacheKey)
		return nil
	}
	// In case timeout encountered  - reset timeout status (to _noTimeOutEncountered) because we succeeded to get results before timeout
	// Set in cache failed
	if err := client.setTimeOutStatus(podSpecCacheKey, _noTimeOutEncountered, _resetTimeoutTTL); err != nil {
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, err.Error()))
		err = errors.Wrap(err, "error encountered while trying to reset timeOut status in cache.")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults"))
		return err
	}
	//Set in cache succeeded
	tracer.Info("updated timeOut status in cache to no timeout encountered", "podSpecCacheKey", podSpecCacheKey)
	return nil
}

// marshalScanResults convert the given ContainerVulnerabilityScanInfo and error to containerVulnerabilityCacheResultsWrapper and marshaling the new object to a string
func (client *AzdSecInfoProviderCacheClient) marshalScanResults(containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) (string, error) {
	tracer := client.tracerProvider.GetTracer("marshalScanResults")
	// Get error msg. Error must be converted to string because marshal ignores unexported fields and error struct contains only unexported fields
	errToCacheString := client.convertErrorToString(err)

	// Create containerVulnerabilityCacheResultsWrapper
	containerVulnerabilityScanInfoWrapper := &containerVulnerabilityCacheResultsWrapper{
		ContainerVulnerabilityScanInfo: containerVulnerabilityScanInfo,
		ErrString:                      errToCacheString,
	}
	// Marshal object
	marshaledContainerVulnerabilityScanInfoWrapper, err := json.Marshal(containerVulnerabilityScanInfoWrapper)
	if err != nil {
		err = errors.Wrap(err, "Failed on json.Marshal containerVulnerabilityScanInfoWrapper")
		tracer.Error(err, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProviderCacheClient.marshalScanResults"))
		return "", err
	}

	// Cast to string
	ser := string(marshaledContainerVulnerabilityScanInfoWrapper)
	tracer.Info("containerVulnerabilityCacheResultsWrapper marshalled successfully")
	return ser, nil
}

// GetPodSpecCacheKey get the cache key without the prefix of a given podSpec
// The key is containerName:imageName as string seperate each containerName:imageName by comma.
// For example - 'myName1:alpine,myName2:nginx'
func (client *AzdSecInfoProviderCacheClient) GetPodSpecCacheKey(podSpec *admisionrequest.PodSpec) string {
	containers := utils.ExtractContainersFromPodSpecAsString(podSpec)
	// Sort the array - it is important for the cache to be sorted.
	sort.Strings(containers)
	podSpecCacheKey := strings.Join(containers, ",")
	return podSpecCacheKey
}

// unmarshalScanResults convert the given ContainerVulnerabilityScanInfo and error to containerVulnerabilityCacheResultsWrapper and marshaling the new object to a string
func (client *AzdSecInfoProviderCacheClient) unmarshalScanResults(ContainerVulnerabilityScanInfoString string) (*containerVulnerabilityCacheResultsWrapper, error) {
	tracer := client.tracerProvider.GetTracer("unMarshalScanResults")

	// Unmarshal object
	containerVulnerabilityScanInfoWrapper := new(containerVulnerabilityCacheResultsWrapper)
	unmarshalErr := json.Unmarshal([]byte(ContainerVulnerabilityScanInfoString), containerVulnerabilityScanInfoWrapper)
	if unmarshalErr != nil {
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on json.Unmarshal containerVulnerabilityScanInfoWrapper")
		tracer.Error(unmarshalErr, "")
		client.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(unmarshalErr, "AzdSecInfoProviderCacheClient.unmarshalScanResults"))
		return nil, unmarshalErr
	}
	return containerVulnerabilityScanInfoWrapper, nil
}

// getErrorStoredInCache check if the errorString is empty. If empty it means the error stored in cache is nil.
// Otherwise the create a new error with error msg equal to errorString
func (client *AzdSecInfoProviderCacheClient) getErrorStoredInCache(errorString string) error {
	// The error stored in cache is nil
	if errorString == "" {
		return nil
	}
	// If there is an error stored in cache from previous runs
	// Convert the error stored in cache as string to an error type with the same information
	errorStoredInCache := errors.New(errorString)
	return errorStoredInCache
}

// convertErrorToString check if the given error is nil. If nil returns the empty string Otherwise the error msg.
func (client *AzdSecInfoProviderCacheClient) convertErrorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// getTimeOutCacheKey returns the timeout cache key of a given podSpecCacheKey
func (client *AzdSecInfoProviderCacheClient) getTimeOutCacheKey(podSpecCacheKey string) string {
	return _timeoutPrefixForCacheKey + podSpecCacheKey
}

// getContainerVulnerabilityScanInfoCacheKey returns the ContainerVulnerabilityScanInfo cache key of a given podSpecCacheKey
func (client *AzdSecInfoProviderCacheClient) getContainerVulnerabilityScanInfoCacheKey(podSpecCacheKey string) string {
	return _containerVulnerabilityScanInfoPrefixForCacheKey + podSpecCacheKey
}
