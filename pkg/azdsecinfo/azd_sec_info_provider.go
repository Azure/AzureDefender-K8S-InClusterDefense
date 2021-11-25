package azdsecinfo

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	registryerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
	"time"
)

const (
	_defaultTimeDurationGetContainersVulnerabilityScanInfo = 2850 * time.Millisecond // 2.85 seconds - can't multiply float in seconds
	_timeoutPrefixForCacheKey = "timeout"
	_containerVulnerabilityScanInfoPrefixForCacheKey = "ContainerVulnerabilityScanInfo"
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {
	// GetContainersVulnerabilityScanInfo receives pod template spec containing containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta) ([]*contracts.ContainerVulnerabilityScanInfo, error)
}

// AzdSecInfoProvider implements IAzdSecInfoProvider interface
var _ IAzdSecInfoProvider = (*AzdSecInfoProvider)(nil)

// AzdSecInfoProvider represents default implementation of IAzdSecInfoProvider interface
type AzdSecInfoProvider struct {
	//tracerProvider is tracer provider of AzdSecInfoProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter is metric submitter of AzdSecInfoProvider
	metricSubmitter metric.IMetricSubmitter
	// argDataProvider is the ARG provider which provides any ARG data
	argDataProvider arg.IARGDataProvider
	// tag2digestResolver is the resolver of images to their digests
	tag2digestResolver tag2digest.ITag2DigestResolver

	// GetContainersVulnerabilityScanInfoTimeoutDuration is the duration of  GetContainersVulnerabilityScanInfo that AzdSecInfoProvider
	//will try to fetch the results of some digest,
	//if the duration will exceed, the program will return result of the first container that unscanned reason .
	//the results still will be saved in the cache.
	getContainersVulnerabilityScanInfoTimeoutDuration time.Duration
	// cacheClient is a cache for mapping digest to scan results
	cacheClient cache.ICacheClient
	// azdSecInfoProviderConfiguration is configuration data for AzdSecInfoProvider
	azdSecInfoProviderConfiguration *AzdSecInfoProviderConfiguration
}

// AzdSecInfoProviderConfiguration is configuration data for AzdSecInfoProvider
type AzdSecInfoProviderConfiguration struct {
	// CacheExpirationTimeTimeout is the expiration time **in hours** for timout.
	CacheExpirationTimeTimeout int
	// CacheExpirationContainerVulnerabilityScanInfo is the expiration time **in seconds** for ContainerVulnerabilityScanInfo.
	CacheExpirationContainerVulnerabilityScanInfo int
}

type ContainerVulnerabilityScanInfoWrapper struct {
	ContainerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo
	Err                            error
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider,
	argDataProvider arg.IARGDataProvider,
	tag2digestResolver tag2digest.ITag2DigestResolver,
	GetContainersVulnerabilityScanInfoTimeoutDuration *utils.TimeoutConfiguration,
 	cacheClient cache.ICacheClient, azdSecInfoProviderConfiguration *AzdSecInfoProviderConfiguration) *AzdSecInfoProvider {

	// In case that GetContainersVulnerabilityScanInfoTimeoutDuration.TimeDurationInMS is empty (zero) - use default value.
	getContainersVulnerabilityScanInfoTimeoutDuration := _defaultTimeDurationGetContainersVulnerabilityScanInfo
	if GetContainersVulnerabilityScanInfoTimeoutDuration.TimeDurationInMS > 0 {
		getContainersVulnerabilityScanInfoTimeoutDuration = GetContainersVulnerabilityScanInfoTimeoutDuration.ParseTimeoutConfigurationToDuration()
	}
	return &AzdSecInfoProvider{
		tracerProvider:     instrumentationProvider.GetTracerProvider("AzdSecInfoProvider"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		argDataProvider:    argDataProvider,
		tag2digestResolver: tag2digestResolver,
		getContainersVulnerabilityScanInfoTimeoutDuration: getContainersVulnerabilityScanInfoTimeoutDuration,
		cacheClient: cacheClient,
		azdSecInfoProviderConfiguration:   azdSecInfoProviderConfiguration,
	}
}

// GetContainersVulnerabilityScanInfo receives api-resource pod spec containing containers, resource deployed metadata and kind
// Function returns evaluated ContainerVulnerabilityScanInfo for pod spec's container list (pod spec can be related to template of any resource creates pods eventually)
func (provider *AzdSecInfoProvider) GetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("GetContainersVulnerabilityScanInfo")
	tracer.Info("Received:", "podSpec", podSpec, "resourceMetadata", resourceMetadata, "resourceKind", resourceKind)

	// Argument validation
	if podSpec == nil || resourceMetadata == nil || resourceKind == nil {
		err := errors.Wrap(utils.NilArgumentError, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		tracer.Error(err, "")
		return nil, err
	}

	concatenatedImages := provider.getConcatenatedImages(podSpec)

	// Try to get ContainersVulnerabilityScanInfo from cache
	ContainersVulnerabilityScanInfoWrapperFromCache, err := provider.getResultsfromCache(concatenatedImages)
	if err == nil{ // got returns from cache
		containerVulnerabilityScanInfo, err := ContainersVulnerabilityScanInfoWrapperFromCache.ContainerVulnerabilityScanInfo, ContainersVulnerabilityScanInfoWrapperFromCache.Err
		if err != nil{
			err = errors.Wrap(err, "Got error from ContainerVulnerabilityScanInfo stored in cache")
			tracer.Error(err, "")
			return nil, err
		}
		tracer.Info("ContainerVulnerabilityScanInfo unmarshalled successfully")
		return containerVulnerabilityScanInfo, nil
	} // failed to get results from cache - skip and get results from providers
	tracer.Info("No ContainersVulnerabilityScanInfo as value in cache")


	// Try to get containers vulnerabilities in diff thread.
	chanTimeout := make(chan *utils.ChannelDataWrapper, 1)
	go provider.getContainersVulnerabilityScanInfoSyncWrapper(podSpec, resourceMetadata, resourceKind, chanTimeout, concatenatedImages)

	// Choose the first thread that finish.
	select {
	// No timeout case:
	case channelData, isChannelOpen := <-chanTimeout:
		// Check if channel is open
		if !isChannelOpen {
			err = errors.Wrap(utils.ReadFromClosedChannelError, "Channel closed unexpectedly")
			tracer.Error(err, "")
			return nil, err
		}

		// Try to extract []*contracts.ContainerVulnerabilityScanInfo from channelData
		containerVulnerabilityScanInfo, err := provider.extractContainersVulnerabilityScanInfoFromChannelData(channelData)
		if err != nil {
			err = errors.Wrap(err, "failed to extract []*contracts.ContainerVulnerabilityScanInfo from channel data")
			tracer.Error(err, "")
			return nil, err
		}

		// Success path:
		// Update cache to no timeout encountered for this pod
		provider.updateCacheAfterGettingResults(concatenatedImages)
		close(chanTimeout)
		tracer.Info("GetContainersVulnerabilityScanInfo finished extract []*contracts.ContainerVulnerabilityScanInfo.", "podSpec", podSpec, "ContainerVulnerabilityScanInfo", containerVulnerabilityScanInfo)
		return containerVulnerabilityScanInfo, nil

	// Timeout case:
	case <-time.After(provider.getContainersVulnerabilityScanInfoTimeoutDuration):
		return provider.timeoutEncounteredGetContainersVulnerabilityScanInfo(concatenatedImages)
	}
}

// getContainersVulnerabilityScanInfoSyncWrapper runs getContainersVulnerabilityScanInfo and insert the result into the channel.
func (provider *AzdSecInfoProvider) getContainersVulnerabilityScanInfoSyncWrapper(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta, chanTimeout chan *utils.ChannelDataWrapper, concatenatedImages string) {
	tracer := provider.tracerProvider.GetTracer("getContainersVulnerabilityScanInfoSyncWrapper")
	containerVulnerabilityScanInfo, err := provider.getContainersVulnerabilityScanInfo(podSpec, resourceMetadata, resourceKind)

	// set ContainersVulnerabilityScanInfo, Err in cache
	provider.setResultsInCache(concatenatedImages, containerVulnerabilityScanInfo, err)

	channelData := utils.NewChannelDataWrapper(containerVulnerabilityScanInfo, err)
	chanTimeout <- channelData
	tracer.Info("channelData inserted to chanTimeout sucessfully", "channelData", channelData)
}

// extractContainersVulnerabilityScanInfoFromChannelData is method that gets *utils.ChannelDataWrapper and tries to extract the data from the channel.
func (provider *AzdSecInfoProvider) extractContainersVulnerabilityScanInfoFromChannelData(channelDataWrapper *utils.ChannelDataWrapper) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("extractContainersVulnerabilityScanInfoFromChannelData")
	if channelDataWrapper == nil {
		err := errors.Wrap(utils.NilArgumentError, "got nil channel data")
		tracer.Error(err, "")
		return nil, err
		// Check that the error of channel data is nil
	}
	// Extract data from channelDataWrapper
	data, err := channelDataWrapper.GetData()
	if err != nil {
		err = errors.Wrap(err, "returned error from channel")
		tracer.Error(err, "")
		return nil, err
	}
	// Try to cast.
	containerVulnerabilityScanInfo, canConvert := data.([]*contracts.ContainerVulnerabilityScanInfo)
	if !canConvert {
		err = errors.Wrap(utils.CantConvertChannelDataWrapper, "failed to convert ChannelDataWrapper.DataWrapper to []*contracts.ContainerVulnerabilityScanInfo")
		tracer.Error(err, "")
		return nil, err
	}
	// Successfully extract data from channel
	tracer.Info("extractContainersVulnerabilityScanInfoFromChannelData finished to extract data", "data", containerVulnerabilityScanInfo)
	return containerVulnerabilityScanInfo, nil
}

// getContainersVulnerabilityScanInfo try to get containers vulnerabilities scan info
func (provider *AzdSecInfoProvider) getContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("getContainersVulnerabilityScanInfo")

	// Convert pull secrets from reference object to strings
	imagePullSecrets := make([]string, 0, len(podSpec.ImagePullSecrets))
	// element  a LocalObjectReference{Name}
	for _, element := range podSpec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, element.Name)
	}

	// Build resource deployment context
	resourceCtx := tag2digest.NewResourceContext(resourceMetadata.Namespace, imagePullSecrets, podSpec.ServiceAccountName)
	tracer.Info("resourceCtx", "resourceCtx", resourceCtx)

	// insert container vulnerability scan information for init containers and containers to vulnSecInfoContainers
	vulnSecInfoContainers, err := provider.getVulnSecInfoContainers(podSpec, resourceCtx)
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to getVulnSecInfoContainers")
		tracer.Error(wrappedError, "")
		return nil, wrappedError
	}

	return vulnSecInfoContainers, nil
}

// getVulnSecInfoContainers gets vulnSecInfoContainers array with the scan results of the given containers.
// It runs each container scan in parallel and returns only when all the scans are finished and the array is updated
func (provider *AzdSecInfoProvider) getVulnSecInfoContainers(podSpec *corev1.PodSpec, resourceCtx *tag2digest.ResourceContext)  ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("getVulnSecInfoContainers")

	// Initialize container vuln scan info list
	vulnSecInfoContainers := make([]*contracts.ContainerVulnerabilityScanInfo, 0, len(podSpec.InitContainers)+len(podSpec.Containers))

	if podSpec == nil {
		err := errors.Wrap(utils.NilArgumentError, "failed in AzdSecInfoProvider.getVulnSecInfoContainers. Unexpected: pod.Spec is nil")
		tracer.Error(err, "")
		return nil, err
	}

	// vulnerabilitySecInfoChannel is a channel for (*contracts.ContainerVulnerabilityScanInfo, error)
	vulnerabilitySecInfoChannel := make(chan *utils.ChannelDataWrapper, len(podSpec.InitContainers) + len(podSpec.Containers))
	// Get container vulnerability scan information in parallel
	// Each call send data to channel vulnerabilitySecInfoChannel
	for i := range podSpec.InitContainers {
		go provider.getSingleContainerVulnerabilityScanInfoSyncWrapper(&podSpec.InitContainers[i], resourceCtx, vulnerabilitySecInfoChannel)
	}
	for i := range podSpec.Containers {
		go provider.getSingleContainerVulnerabilityScanInfoSyncWrapper(&podSpec.Containers[i], resourceCtx, vulnerabilitySecInfoChannel)
	}

	for i := 0; i <  len(podSpec.InitContainers) + len(podSpec.Containers); i++ { // No deadlock as a result of the loop because the number of receivers is identical to the number of senders
		vulnerabilitySecInfoWrapper, isChannelOpen := <- vulnerabilitySecInfoChannel // Because the channel is buffered all goroutines will finish executing (no goroutine leak)
		if !isChannelOpen {
			err := errors.Wrap(utils.ReadFromClosedChannelError, "failed in AzdSecInfoProvider.getVulnSecInfoContainers. Channel closed unexpectedly")
			tracer.Error(err, "")
			return nil, err
		}
		if vulnerabilitySecInfoWrapper == nil {
			err := errors.Wrap(utils.NilArgumentError, "failed in getSingleContainerVulnerabilityScanInfoSync. Unexpected: vulnerabilitySecInfoWrapper is nil")
			tracer.Error(err, "")
			return nil, err
		}
		// Check if an error occurred during getSingleContainerVulnerabilityScanInfo
		vulnerabilitySecInfoDataWrapper, err := vulnerabilitySecInfoWrapper.GetData()
		if err != nil {
			err = errors.Wrap(err, "failed in getSingleContainerVulnerabilityScanInfoSync.")
			tracer.Error(err, "")
			return nil, err
		}
		// Convert vulnerabilitySecInfoWrapper.DataWrapper to vulnerabilitySecInfo
		vulnerabilitySecInfo, canConvert := vulnerabilitySecInfoDataWrapper.(*contracts.ContainerVulnerabilityScanInfo)
		if !canConvert {
			err := errors.Wrap(utils.CantConvertChannelDataWrapper, "failed to convert ChannelDataWrapper.DataWrapper to *contracts.ContainerVulnerabilityScanInfo")
			tracer.Error(err, "")
			return nil, err
		}

		// No errors - add scan info to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}
	close(vulnerabilitySecInfoChannel)
	return vulnSecInfoContainers, nil
}

//getSingleContainerVulnerabilityScanInfoSyncWrapper wrap getSingleContainerVulnerabilityScanInfo.
// It sends getSingleContainerVulnerabilityScanInfo results to the channel
func (provider *AzdSecInfoProvider) getSingleContainerVulnerabilityScanInfoSyncWrapper(container *corev1.Container, resourceCtx *tag2digest.ResourceContext,  vulnerabilitySecInfoChannel chan *utils.ChannelDataWrapper){
	info, err := provider.getSingleContainerVulnerabilityScanInfo(container, resourceCtx)
	vulnerabilitySecInfoChannel <- utils.NewChannelDataWrapper(info, err)
}

// getSingleContainerVulnerabilityScanInfo receives a container, and it's belonged deployed resource context, and returns fetched ContainerVulnerabilityScanInfo
func (provider *AzdSecInfoProvider) getSingleContainerVulnerabilityScanInfo(container *corev1.Container, resourceCtx *tag2digest.ResourceContext) (*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("getSingleContainerVulnerabilityScanInfo")
	tracer.Info("Received:", "container image ref", container.Image, "resourceCtx", resourceCtx)

	if container == nil || resourceCtx == nil {
		err := errors.Wrap(utils.NilArgumentError, "AzdSecInfoProvider.getSingleContainerVulnerabilityScanInfo")
		tracer.Error(err, "")
		return nil, err
	}

	// Get image ref
	imageRef, err := registryutils.GetImageReference(container.Image)
	if err != nil {
		err = errors.Wrap(err, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo.registry.GetImageReference")
		tracer.Error(err, "")
		return nil, err
	}
	tracer.Info("Container image ref extracted", "imageRef", imageRef)

	// Checks if the image registry  not ACR.
	if !registryutils.IsRegistryEndpointACR(imageRef.Registry()) {
		tracer.Info("Image from another registry than ACR received", "Registry", imageRef.Registry())
		return provider.buildContainerVulnerabilityScanInfoUnScannedWithReason(container, contracts.ImageIsNotInACRRegistryUnscannedReason), nil
	}

	digest, err := provider.tag2digestResolver.Resolve(imageRef, resourceCtx)
	if err != nil {
		// TODO wait until @maayaan merge his PR and then add tests for this method. ( Maayan already created IAZdSecInfoProvider mock)
		unscannedReason, isErrParsedToUnscannedReason := registryerrors.TryParseErrToUnscannedWithReason(err)
		if !isErrParsedToUnscannedReason {
			err = errors.Wrap(err, "Unexpected error while trying to resolve digest")
			tracer.Error(err, "")
			return nil, err
		}

		// Err parsed successfully to known unscanned reason.
		tracer.Info("Err from Tag2DigestResolver parsed successfully to known unscanned reason", "Err", err, "unscannedReason", unscannedReason)
		return provider.buildContainerVulnerabilityScanInfoUnScannedWithReason(container, *unscannedReason), nil
	}

	scanStatus, scanFindings, err := provider.argDataProvider.GetImageVulnerabilityScanResults(imageRef.Registry(), imageRef.Repository(), digest)
	if err != nil {
		// TODO wait until @maayaan merge his PR and then add tests for this method. ( Maayan already created IAZdSecInfoProvider mock)
		unscannedReason, isErrParsedToUnscannedReason := registryerrors.TryParseErrToUnscannedWithReason(err)
		if !isErrParsedToUnscannedReason {
			err = errors.Wrap(err, "Unexpected error while trying to get results from ARGDataProvider")
			tracer.Error(err, "")
			return nil, err
		}
		// Err parsed successfully to known unscanned reason.
		tracer.Info("Err from ARGDataProvider parsed successfully to known unscanned reason", "Err", err, "unscannedReason", unscannedReason)
		return provider.buildContainerVulnerabilityScanInfoUnScannedWithReason(container, *unscannedReason), nil
	}

	tracer.Info("results from ARG data provider", "scanStatus", scanStatus, "scanFindings", scanFindings)
	// Build scan info from provided scan results
	info := provider.buildContainerVulnerabilityScanInfoFromResult(container, digest, scanStatus, scanFindings)

	return info, nil
}

// buildContainerVulnerabilityScanInfoFromResult build the info object from data provided
func (provider *AzdSecInfoProvider) buildContainerVulnerabilityScanInfoFromResult(container *corev1.Container, digest string, scanStatus contracts.ScanStatus, scanFindigs []*contracts.ScanFinding) *contracts.ContainerVulnerabilityScanInfo {
	info := &contracts.ContainerVulnerabilityScanInfo{
		Name: container.Name,
		Image: &contracts.Image{
			Name:   container.Image,
			Digest: digest,
		},
		ScanStatus:     scanStatus,
		ScanFindings:   scanFindigs,
		AdditionalData: nil,
	}
	return info
}

// buildContainerVulnerabilityScanInfoFromResult build the info object from data provided
func (provider *AzdSecInfoProvider) buildContainerVulnerabilityScanInfoUnScannedWithReason(container *corev1.Container, reason contracts.UnscannedReason) *contracts.ContainerVulnerabilityScanInfo {
	return &contracts.ContainerVulnerabilityScanInfo{
		Name: container.Name,
		Image: &contracts.Image{
			Name:   container.Image,
			Digest: "",
		},
		ScanStatus:   contracts.Unscanned,
		ScanFindings: nil,
		AdditionalData: map[string]string{
			contracts.UnscannedReasonAnnotationKey: string(reason),
		},
	}
}

// buildListOfContainerVulnerabilityScanInfoWhenTimeout is method that is called when the GetContainerVulnerabilityScanInfo
// got timeout (GetContainersVulnerabilityScanInfoTimeoutDuration) and returns list with one empty container with
//unscanned status and contracts.GetContainersVulnerabilityScanInfoTimeoutUnscannedReason.
// TODO In public preview, we should add timeout without empty container (bad UX) (should be changed also in the REGO).
func (provider *AzdSecInfoProvider) buildListOfContainerVulnerabilityScanInfoWhenTimeout() ([]*contracts.ContainerVulnerabilityScanInfo, error) {

	info := &contracts.ContainerVulnerabilityScanInfo{
		Name: "",
		Image: &contracts.Image{
			Name:   "",
			Digest: "",
		},
		ScanStatus:   contracts.Unscanned,
		ScanFindings: nil,
		AdditionalData: map[string]string{
			contracts.UnscannedReasonAnnotationKey: string(contracts.GetContainersVulnerabilityScanInfoTimeoutUnscannedReason),
		},
	}

	containerVulnerabilityScanInfoList := []*contracts.ContainerVulnerabilityScanInfo{info}
	return containerVulnerabilityScanInfoList, nil
}

// timeoutEncounteredGetContainersVulnerabilityScanInfo is called when timeout is encountered in GetContainersVulnerabilityScanInfo function
// It checks if it is the first time that the request got an error (request is defined as the images of the request).
// If it is the first time, it adds the images to the cache and returns unscanned with metadata.
// If it is not the first time (images already are in cache) or there is an error in the communication with the cache, it returns error.
// TODO Add tests for this behavior.
func (provider *AzdSecInfoProvider) timeoutEncounteredGetContainersVulnerabilityScanInfo(concatenatedImages string) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("timeoutEncounteredGetContainersVulnerabilityScanInfo")

	// Get key for cache
	timeOutCacheKey := provider.getTimeOutCacheKey(concatenatedImages)
	// Check if the concatenatedImages is already in cache
	timeoutEncountered, err := provider.cacheClient.Get(timeOutCacheKey)
	// In case that the concatenatedImages is already in cache.
	if err == nil {
		// TODO Add metric for number of timeouts.
		// if timeoutEncountered == false it means that it is the first time we had timeout - ignore
		// if timeoutEncountered == "true" it means that we encountered timeout - should return an error
		if timeoutEncountered == "true"{
			err = errors.Wrap(err, "Second time that timeout was encountered.")
			tracer.Error(err, "")
			return nil, err
		}
	}
	if err != nil {
		// Check if the error is due to missing key.
		_, isFirstTimeOfTimeout := err.(*cache.MissingKeyCacheError)
		if !isFirstTimeOfTimeout {
			// TODO Add metric new error encountered
			err = errors.Wrap(err, "error encountered while trying to get result of timeout from cache.")
			tracer.Error(err, "")
			return nil, err
		}
	}

	// In case that we got missing key error, it means that the concatenatedImages is not in cache - first time of timeout.
	// Add to cache and return unscanned and timeout.
	if err := provider.cacheClient.Set(timeOutCacheKey, "true", provider.azdSecInfoProviderConfiguration.GetCacheExpirationTimeTimeout()); err != nil {
		// TODO Add metric new error encountered
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
		return nil, err
	}

	tracer.Info("GetContainersVulnerabilityScanInfo got timeout - returning unscanned", "timeDurationOfTimeout", provider.getContainersVulnerabilityScanInfoTimeoutDuration)
	return provider.buildListOfContainerVulnerabilityScanInfoWhenTimeout()
}

func (provider *AzdSecInfoProvider) updateCacheAfterGettingResults(concatenatedImages string){
	tracer := provider.tracerProvider.GetTracer("updateCacheAfterGettingResults")

	// Get key for cache
	timeOutCacheKey := provider.getTimeOutCacheKey(concatenatedImages)
	// Check if the concatenatedImages is already in cache
	timeoutEncountered, err := provider.cacheClient.Get(timeOutCacheKey)
	// In case that the concatenatedImages is already in cache.
	if err == nil {
		tracer.Info("value exist in cache", "concatenatedImages", concatenatedImages, "timeoutEncountered", timeoutEncountered)
		// in case timeoutEncountered is true - change value to false because we succeeded to get results before timeout
		if timeoutEncountered == "true" {
			if err := provider.cacheClient.Set(concatenatedImages, "false", provider.azdSecInfoProviderConfiguration.GetCacheExpirationTimeTimeout()); err != nil {
				// TODO Add metric new error encountered
				err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
				tracer.Error(err, "")
			}
			tracer.Info("updated cache to no timeout encountered", "concatenatedImages", concatenatedImages)
		}
	}
}

func (provider *AzdSecInfoProvider) getConcatenatedImages(podSpec *corev1.PodSpec) string{
	images := utils.ExtractImagesFromPodSpec(podSpec)
	// Sort the array - it is important for the cache to be sorted.
	sort.Strings(images)
	concatenatedImages := strings.Join(images, ",")
	return concatenatedImages
}

// GetCacheExpirationTimeTimeout uses AzdSecInfoProviderConfiguration instance's CacheExpirationTimeTimeout (int)
// to a return a time.Duration object
// In case of invalid argument, use default values (0) which means the value expires immediately.
func (configuration *AzdSecInfoProviderConfiguration) GetCacheExpirationTimeTimeout() time.Duration {
	return time.Duration(configuration.CacheExpirationTimeTimeout) * time.Hour
}

// GetCacheExpirationContainerVulnerabilityScanInfo uses AzdSecInfoProviderConfiguration instance's CacheExpirationContainerVulnerabilityScanInfo (int)
// to a return a time.Duration object
// In case of invalid argument, use default values (0) which means the value expires immediately.
func (configuration *AzdSecInfoProviderConfiguration) GetCacheExpirationContainerVulnerabilityScanInfo() time.Duration {
	return time.Duration(configuration.CacheExpirationTimeTimeout) * time.Second
}

// marshalScanResults convert the given ContainerVulnerabilityScanInfo and error to ContainerVulnerabilityScanInfoWrapper and marshaling the new object to a string
func (provider *AzdSecInfoProvider) marshalScanResults(containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) (string, error) {
	tracer := provider.tracerProvider.GetTracer("marshalScanResults")
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
func (provider *AzdSecInfoProvider) unmarshalScanResults(ContainerVulnerabilityScanInfoString string) (*ContainerVulnerabilityScanInfoWrapper, error) {
	tracer := provider.tracerProvider.GetTracer("unMarshalScanResults")
	containerVulnerabilityScanInfoWrapper :=  new(ContainerVulnerabilityScanInfoWrapper)
	unmarshalErr := json.Unmarshal([]byte(ContainerVulnerabilityScanInfoString), containerVulnerabilityScanInfoWrapper)
	if unmarshalErr != nil {
		unmarshalErr = errors.Wrap(unmarshalErr, "Failed on json.Unmarshal containerVulnerabilityScanInfoWrapper")
		tracer.Error(unmarshalErr, "")
		return nil, unmarshalErr
	}
	return containerVulnerabilityScanInfoWrapper, nil
}

func (provider *AzdSecInfoProvider) getResultsfromCache(concatenatedImages string) (*ContainerVulnerabilityScanInfoWrapper, error) {
	tracer := provider.tracerProvider.GetTracer("getResultsfromCache")
	scanInfoWrapperStringFromCache, err := provider.cacheClient.Get(provider.getContainerVulnerabilityScanInfoCacheKey(concatenatedImages))
	if err != nil { // Key don't exist in cache
		err = errors.Wrap(err, "falied to get unmarshalScanResults from cache")
		tracer.Error(err, "")
		return nil, err
	}
	ContainersVulnerabilityScanInfoWrapperFromCache, err := provider.unmarshalScanResults(scanInfoWrapperStringFromCache)
	if err != nil{ // unmarshal failed
		err = errors.Wrap(err, "failed to unmarshalScanResults from cache")
		tracer.Error(err, "")
		return nil, err
	}
	return ContainersVulnerabilityScanInfoWrapperFromCache, nil
}



func (provider *AzdSecInfoProvider) setResultsInCache(concatenatedImages string, containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) {
	tracer := provider.tracerProvider.GetTracer("setResultsInCache")
	resultsString, err := provider.marshalScanResults(containerVulnerabilityScanInfo, err)
	if err != nil{
		err = errors.Wrap(err, "Failed to marshal ContainerVulnerabilityScanInfo")
		tracer.Error(err, "")
	}else if err = provider.cacheClient.Set(provider.getContainerVulnerabilityScanInfoCacheKey(concatenatedImages), resultsString,provider.azdSecInfoProviderConfiguration.GetCacheExpirationContainerVulnerabilityScanInfo() ); err != nil {
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
	}else{
		tracer.Info("Set ContainerVulnerabilityScanInfo in cache")
	}
}

// getTimeOutCacheKey returns the timeout cache key of a given concatenatedImages
func (provider *AzdSecInfoProvider) getTimeOutCacheKey (concatenatedImages string) string {
	return _timeoutPrefixForCacheKey + concatenatedImages
}

// getContainerVulnerabilityScanInfoCacheKey returns the ContainerVulnerabilityScanInfo cache key of a given concatenatedImages
func (provider *AzdSecInfoProvider) getContainerVulnerabilityScanInfoCacheKey (concatenatedImages string) string {
	return _containerVulnerabilityScanInfoPrefixForCacheKey + concatenatedImages
}