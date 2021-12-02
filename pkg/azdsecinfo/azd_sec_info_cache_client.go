package azdsecinfo

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sort"
	"strings"
	"time"
)
// AzdSecInfoProviderCacheClient is cache client designated for AzdSecInfoProvider
// It wraps ICache client
type AzdSecInfoProviderCacheClient struct {
	//tracerProvider is tracer provider of AzdSecInfoProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter is metric submitter of AzdSecInfoProvider
	metricSubmitter metric.IMetricSubmitter
	// cacheClient is a cache for mapping digest to scan results and save timeout status
	cacheClient cache.ICacheClient
	// cacheExpirationTimeTimeout is the expiration time **IN MINUTES** for timout.
	cacheExpirationTimeTimeout time.Duration
	// cacheExpirationContainerVulnerabilityScanInfo is the expiration time **IN SECONDS** for ContainerVulnerabilityScanInfo.
	cacheExpirationContainerVulnerabilityScanInfo time.Duration
}

// NewAzdSecInfoProviderCacheClient - AzdSecInfoProviderCacheClient Ctor
func NewAzdSecInfoProviderCacheClient(instrumentationProvider instrumentation.IInstrumentationProvider, cacheClient cache.ICacheClient, azdSecInfoProviderConfiguration *AzdSecInfoProviderConfiguration) *AzdSecInfoProviderCacheClient {
	return &AzdSecInfoProviderCacheClient{
		tracerProvider:     instrumentationProvider.GetTracerProvider("AzdSecInfoProviderCacheClient"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		cacheClient: cacheClient,
		cacheExpirationTimeTimeout: utils.GetMinutes(azdSecInfoProviderConfiguration.CacheExpirationTimeTimeout),
		cacheExpirationContainerVulnerabilityScanInfo: utils.GetSeconds(azdSecInfoProviderConfiguration.CacheExpirationContainerVulnerabilityScanInfo),
	}
}

// getContainerVulnerabilityScanInfofromCache try to get ContainerVulnerabilityScanInfo from cache.
// It gets the results from the cache and parse it to ContainerVulnerabilityScanInfoWrapper object.
// If there is an error with the cache or the value is invalid returns an error.
func (client *AzdSecInfoProviderCacheClient) getContainerVulnerabilityScanInfofromCache(podSpecCacheKey string) ( *ContainerVulnerabilityScanInfoWrapper, error) {
	tracer := client.tracerProvider.GetTracer("getContainerVulnerabilityScanInfofromCache")
	// Get the key
	ContainerVulnerabilityScanInfoCacheKey := client.getContainerVulnerabilityScanInfoCacheKey(podSpecCacheKey)

	// get result from cache
	scanInfoWrapperStringFromCache, err := client.cacheClient.Get(ContainerVulnerabilityScanInfoCacheKey)
	if err != nil { // Key don't exist in cache or error with cache functionality
		// Check if the error is MissingKeyCacheError
		_, isKeyNotFound := err.(*cache.MissingKeyCacheError)
		if isKeyNotFound{ // The key not in the cache
			tracer.Info("ContainerVulnerabilityScanInfoCacheKey is not in cache", "ContainerVulnerabilityScanInfoCacheKey", ContainerVulnerabilityScanInfoCacheKey)
			return nil, err
		}
		// error with cache functionality
		err = errors.Wrap(err, "falied to get ContainerVulnerabilityScanInfo from cache")
		tracer.Error(err, "")
		return nil, err
	}
	// Key exist in cache
	// Parse the results
	ContainersVulnerabilityScanInfoWrapperFromCache, err := client.unmarshalScanResults(scanInfoWrapperStringFromCache)
	if err != nil{ // unmarshal failed
		err = errors.Wrap(err, "failed to unmarshalScanResults from cache")
		tracer.Error(err, "")
		return  nil, err
	}

	return ContainersVulnerabilityScanInfoWrapperFromCache, nil
}

// setContainerVulnerabilityScanInfoInCache set ContainerVulnerabilityScanInfo in cache
// No error is reported back, only tracing it
func (client *AzdSecInfoProviderCacheClient) setContainerVulnerabilityScanInfoInCache(podSpecCacheKey string, containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) {
	tracer := client.tracerProvider.GetTracer("setContainerVulnerabilityScanInfoInCache")
	// Convert results to resultsString
	resultsString, err := client.marshalScanResults(containerVulnerabilityScanInfo, err)
	if err != nil{ // Marshal failed
		err = errors.Wrap(err, "Failed to marshal ContainerVulnerabilityScanInfo")
		tracer.Error(err, "")
	// Try to set resultsString in cache
	}else if err = client.cacheClient.Set(client.getContainerVulnerabilityScanInfoCacheKey(podSpecCacheKey), resultsString, client.cacheExpirationContainerVulnerabilityScanInfo); err != nil {
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
	}else{
		tracer.Info("Set ContainerVulnerabilityScanInfo in cache")
	}
}

// getTimeOutStatus gets the timeout status of the podSpec from cache - how many times timeout occurred for this podSpec
func (client *AzdSecInfoProviderCacheClient) getTimeOutStatus(podSpecCacheKey string) (string, error) {
	tracer := client.tracerProvider.GetTracer("getTimeOutStatus")

	// Get key for cache
	timeOutCacheKey := client.getTimeOutCacheKey(podSpecCacheKey)
	// Get timeoutStatus from cache
	timeoutStatus, err := client.cacheClient.Get(timeOutCacheKey)
	// Key don't exist in cache or error with cache functionality
	if err != nil {
		// Check if the error is MissingKeyCacheError
		_, isKeyNotFound := err.(*cache.MissingKeyCacheError)
		if isKeyNotFound{  // The key not in the cache
			tracer.Info("timeOutCacheKey is not in cache", "timeOutCacheKey", timeOutCacheKey)
			return noTimeOutEncountered, nil
		}
		// error with cache functionality - unknownTimeOutStatus.
		err = errors.Wrap(err, "Error while trying to get timeoutStatus from cache")
		tracer.Error(err, "")
		return unknownTimeOutStatus, err
	}

	// Make sure the string from the cache is a valid timeout status
	switch timeoutStatus {
	case noTimeOutEncountered, oneTimeOutEncountered, twoTimeOutEncountered:
		tracer.Info("timeout status is in cache", "timeoutStatus", timeoutStatus)
		return timeoutStatus, nil
	default:
		err = errors.Wrap(err, "Invalid value in cache")
		tracer.Error(err, "")
		return unknownTimeOutStatus, err
	}
}

// setTimeOutStatusAfterEncounteredTimeout update the timeout status if already exist in cache or set for the first time timeout status
func (client *AzdSecInfoProviderCacheClient) setTimeOutStatusAfterEncounteredTimeout(podSpecCacheKey string, oldTimeOutStatus string) error{
	newTimeOutStatus := oneTimeOutEncountered // default value
	// If already one timeout in cache
	if oldTimeOutStatus == oneTimeOutEncountered{
		newTimeOutStatus = twoTimeOutEncountered
	}
	return client.setTimeOutStatus(podSpecCacheKey, newTimeOutStatus)
}

// setTimeOutStatus set a given timeOutStatus in cache.
func (client *AzdSecInfoProviderCacheClient) setTimeOutStatus(podSpecCacheKey string, timeOutStatus string) error {
	tracer := client.tracerProvider.GetTracer("setTimeOutStatus")

	// Get key for cache
	timeOutCacheKey := client.getTimeOutCacheKey(podSpecCacheKey)
	if err := client.cacheClient.Set(timeOutCacheKey, timeOutStatus, client.cacheExpirationTimeTimeout); err != nil {
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
		return err
	}
	return nil
}

// resetTimeOutInCacheAfterGettingScanResults resets the timeout status in cache after scanResults was received.
// If scanResults was received the timeout is no longer relevant and needs to be reset.
// If no timeout occurred before, do nothing.
func (client *AzdSecInfoProviderCacheClient) resetTimeOutInCacheAfterGettingScanResults(podSpecCacheKey string) {
	tracer := client.tracerProvider.GetTracer("resetTimeOutInCacheAfterGettingScanResults")

	// Get key for cache
	timeOutCacheKey := client.getTimeOutCacheKey(podSpecCacheKey)
	// Check if the timeOutCacheKey is already in cache
	timeoutEncountered, err := client.cacheClient.Get(timeOutCacheKey)
	if err != nil{
		err = errors.Wrap(err, "error encountered while trying to get timeout status from cache.")
		tracer.Error(err, "")
		return
	}

	// In case that the podSpecCacheKey is already in cache.
	tracer.Info("timeout status exist in cache", "podSpecCacheKey", podSpecCacheKey, "timeoutEncountered", timeoutEncountered)
	// in case timeoutEncountered is true - change value to false because we succeeded to get results before timeout
	if timeoutEncountered != noTimeOutEncountered {
		if err := client.setTimeOutStatus(podSpecCacheKey, noTimeOutEncountered); err != nil {
			// TODO Add metric new error encountered
			err = errors.Wrap(err, "error encountered while trying to update timeOut status in cache.")
			tracer.Error(err, "")
		}else{ //Set in cache succeeded
			tracer.Info("updated timeOut status in cache to no timeout encountered", "podSpecCacheKey", podSpecCacheKey)
		}
		return
	}
	tracer.Info("No need to update timeOut status in cache because it is already set to no timeout encountered", "podSpecCacheKey", podSpecCacheKey)
}

// marshalScanResults convert the given ContainerVulnerabilityScanInfo and error to ContainerVulnerabilityScanInfoWrapper and marshaling the new object to a string
func (client *AzdSecInfoProviderCacheClient) marshalScanResults(containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) (string, error) {
	tracer := client.tracerProvider.GetTracer("marshalScanResults")
	// Create ContainerVulnerabilityScanInfoWrapper
	containerVulnerabilityScanInfoWrapper := &ContainerVulnerabilityScanInfoWrapper{
		ContainerVulnerabilityScanInfo: containerVulnerabilityScanInfo,
		Err:                            err,
	}
	// Marshal object
	marshaledContainerVulnerabilityScanInfoWrapper, err := json.Marshal(containerVulnerabilityScanInfoWrapper)
	if err != nil {
		err = errors.Wrap(err, "Failed on json.Marshal containerVulnerabilityScanInfoWrapper")
		tracer.Error(err, "")
		return "", err
	}

	// Cast to string
	ser := string(marshaledContainerVulnerabilityScanInfoWrapper)
	tracer.Info("ContainerVulnerabilityScanInfoWrapper marshalled successfully")
	return ser, nil
}

// unmarshalScanResults convert the given ContainerVulnerabilityScanInfo and error to ContainerVulnerabilityScanInfoWrapper and marshaling the new object to a string
func (client *AzdSecInfoProviderCacheClient) unmarshalScanResults(ContainerVulnerabilityScanInfoString string) (*ContainerVulnerabilityScanInfoWrapper, error) {
	tracer := client.tracerProvider.GetTracer("unMarshalScanResults")

	// Unmarshal object
	containerVulnerabilityScanInfoWrapper :=  new(ContainerVulnerabilityScanInfoWrapper)
	unmarshalErr := json.Unmarshal([]byte(ContainerVulnerabilityScanInfoString), containerVulnerabilityScanInfoWrapper)
	if unmarshalErr != nil {
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on json.Unmarshal containerVulnerabilityScanInfoWrapper")
		tracer.Error(unmarshalErr, "")
		return nil, unmarshalErr
	}
	return containerVulnerabilityScanInfoWrapper, nil
}

// getPodSpecCacheKey get the cache key without the prefix of a given podSpec
func (client *AzdSecInfoProviderCacheClient) getPodSpecCacheKey(podSpec *corev1.PodSpec) string{
	images := utils.ExtractImagesFromPodSpec(podSpec)
	// Sort the array - it is important for the cache to be sorted.
	sort.Strings(images)
	podSpecCacheKey := strings.Join(images, ",")
	return podSpecCacheKey
}

// getTimeOutCacheKey returns the timeout cache key of a given podSpecCacheKey
func (client *AzdSecInfoProviderCacheClient) getTimeOutCacheKey (podSpecCacheKey string) string {
	return _timeoutPrefixForCacheKey + podSpecCacheKey
}

// getContainerVulnerabilityScanInfoCacheKey returns the ContainerVulnerabilityScanInfo cache key of a given podSpecCacheKey
func (client *AzdSecInfoProviderCacheClient) getContainerVulnerabilityScanInfoCacheKey (podSpecCacheKey string) string {
	return _containerVulnerabilityScanInfoPrefixForCacheKey + podSpecCacheKey
}
