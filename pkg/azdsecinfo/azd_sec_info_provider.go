package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	azdsecinfometrics "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	registryerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// Default time duration for GetContainersVulnerabilityScanInfo IN MILLISECONDS
const _defaultTimeDurationGetContainersVulnerabilityScanInfo = 2850 * time.Millisecond // 2.85 seconds - can't multiply float in seconds

// The status of timeout during the run
const (
	_unknownTimeOutStatus                = -1
	_noTimeOutEncountered                = 0
	_numberOfTimeOutEncounteredThreshold = 3
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {
	// GetContainersVulnerabilityScanInfo receives pod template spec containing containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainersVulnerabilityScanInfo(podSpec *admisionrequest.SpecRes, resourceMetadata *admisionrequest.MetadataRes) ([]*contracts.ContainerVulnerabilityScanInfo, error)
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
	// getContainersVulnerabilityScanInfoTimeoutDuration is the duration of  GetContainersVulnerabilityScanInfo that AzdSecInfoProvider
	//will try to fetch the results of some digest,
	//if the duration will exceed, the program will return result of the first container that unscanned reason .
	//the results still will be saved in the cache.
	getContainersVulnerabilityScanInfoTimeoutDuration time.Duration
	// cacheClient is a cache client for AzdSecInfoProvider (mapping podSpec to scan results and save timeout status)
	cacheClient IAzdSecInfoProviderCacheClient
}

// AzdSecInfoProviderConfiguration is configuration data for AzdSecInfoProvider
type AzdSecInfoProviderConfiguration struct {
	// CacheExpirationTimeTimeout is the expiration time **IN MINUTES** for timout.
	CacheExpirationTimeTimeout int
	// CacheExpirationContainerVulnerabilityScanInfo is the expiration time **IN SECONDS** for ContainerVulnerabilityScanInfo.
	CacheExpirationContainerVulnerabilityScanInfo int
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider,
	argDataProvider arg.IARGDataProvider,
	tag2digestResolver tag2digest.ITag2DigestResolver,
	GetContainersVulnerabilityScanInfoTimeoutDuration *utils.TimeoutConfiguration,
	cacheClient IAzdSecInfoProviderCacheClient) *AzdSecInfoProvider {

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
	}
}

// GetContainersVulnerabilityScanInfo receives api-resource pod spec containing containers, resource deployed metadata and kind
// Function returns evaluated ContainerVulnerabilityScanInfo for pod spec's container list (pod spec can be related to template of any resource creates pods eventually)
// Function Logic:
// 1. validate Arguments
// 2. Try to get ContainersVulnerabilityScanInfo from cache. If succeeded (got results from cache either valid results or invalid results and error) - return results
// 3. In a new thread, get ContainersVulnerabilityScanInfo. If it takes more than the defined timeout we must return a response to the API server.
// So if it is the first or second there is timeout for this podSpec - return unscanned result (meaning gatekeeper will block this request)
// Otherwise return an error and don't block the request
// If no timeout occurred - save the results in the cache, reset the timeout status and return the results
// For more information - see README
func (provider *AzdSecInfoProvider) GetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("GetContainersVulnerabilityScanInfo")
	tracer.Info("Received:", "podSpec", podSpec, "resourceMetadata", resourceMetadata)

	// Arguments validation
	if podSpec == nil || resourceMetadata == nil {
		err := errors.Wrap(utils.NilArgumentError, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo"))
		return nil, err
	}

	// The key to be set in cache for the pod spec (current request). Without prefix (timeout or ContainerVulnerabilityScanInfo)
	podSpecCacheKey := provider.cacheClient.GetPodSpecCacheKey(podSpec)

	// Try to get ContainersVulnerabilityScanInfo from cache
	// If error is not nil - There are two options: 1. missing key 2. functionality error from cache. In both cases continue to fetch results from provider.
	// If error is nil - There are results from cache. Two options for the results from cache:
	// 		1. The result is an error occurred in previous run. Return the error occurred
	// 		2. The result is ContainerVulnerabilityScanInfo  -  no errors occurred in previous run. Return the results
	ContainersVulnerabilityScanInfo, errorStoredInCache, err := provider.cacheClient.GetContainerVulnerabilityScanInfofromCache(podSpecCacheKey)
	if err != nil { // failed to get results from cache - skip and get results from providers
		if cache.IsMissingKeyCacheError(err){
			tracer.Info("Missing key. Couldn't get ContainerVulnerabilityScanInfo from cache")
		}else{
			err = errors.Wrap(err, "Couldn't get ContainersVulnerabilityScanInfo from cache: error encountered")
			tracer.Error(err, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo"))
		}
	} else { // No error means that there are results from cache
		// If an error was stored in cache (error from previous results) return the error in order to avoid multiple failed requests
		if errorStoredInCache != nil {
			errorStoredInCache = errors.Wrap(errorStoredInCache, "Got error from ContainerVulnerabilityScanInfo stored in cache")
			tracer.Error(errorStoredInCache, "")
			return nil, errorStoredInCache
		}
		// Results are valid - return the results
		tracer.Info("Got ContainersVulnerabilityScanInfo from cache successfully")
		return ContainersVulnerabilityScanInfo, nil
	}

	// Try to get containers vulnerabilities in diff thread.
	chanTimeout := make(chan *utils.ChannelDataWrapper, 1)
	go provider.getContainersVulnerabilityScanInfoSyncWrapper(podSpec, resourceMetadata, chanTimeout, podSpecCacheKey)

	// Choose the first thread that finish.
	select {
	// No timeout case:
	case channelData, isChannelOpen := <-chanTimeout:
		return provider.noTimeoutEncounteredGetContainersVulnerabilityScanInfo(podSpec, chanTimeout, channelData, isChannelOpen, podSpecCacheKey)
	// Timeout case:
	case <-time.After(provider.getContainersVulnerabilityScanInfoTimeoutDuration):
		return provider.timeoutEncounteredGetContainersVulnerabilityScanInfo(podSpec, podSpecCacheKey)
	}
}

// getContainersVulnerabilityScanInfoSyncWrapper runs getContainersVulnerabilityScanInfo and insert the result into the channel.
func (provider *AzdSecInfoProvider) getContainersVulnerabilityScanInfoSyncWrapper(podSpec *admisionrequest.SpecRes, resourceMetadata *admisionrequest.MetadataRes, chanTimeout chan *utils.ChannelDataWrapper, podSpecCacheKey string) {
	tracer := provider.tracerProvider.GetTracer("getContainersVulnerabilityScanInfoSyncWrapper")
	containerVulnerabilityScanInfo, err := provider.getContainersVulnerabilityScanInfo(podSpec, resourceMetadata)

	// Set both ContainersVulnerabilityScanInfo and err in cache
	// Set results in cache here and not in the upper function in order to set results in cache even if timeout has occurred.
	go func() {
		errFromCache := provider.cacheClient.SetContainerVulnerabilityScanInfoInCache(podSpecCacheKey, containerVulnerabilityScanInfo, err)
		if errFromCache != nil {
			errFromCache = errors.Wrap(errFromCache, "Failed to set containerVulnerabilityScanInfo in cache")
			tracer.Error(errFromCache, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.getContainersVulnerabilityScanInfoSyncWrapper failed to store results in cache"))
		} else {
			tracer.Info("Set containerVulnerabilityScanInfo in cache successfully")
		}
	}()

	// Send results to the channel
	channelData := utils.NewChannelDataWrapper(containerVulnerabilityScanInfo, err)
	chanTimeout <- channelData
	tracer.Info("channelData inserted to chanTimeout successfully", "channelData", channelData)
}

// extractContainersVulnerabilityScanInfoFromChannelData is method that gets *utils.ChannelDataWrapper and tries to extract the data from the channel.
func (provider *AzdSecInfoProvider) extractContainersVulnerabilityScanInfoFromChannelData(channelDataWrapper *utils.ChannelDataWrapper) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("extractContainersVulnerabilityScanInfoFromChannelData")
	if channelDataWrapper == nil {
		err := errors.Wrap(utils.NilArgumentError, "got nil channel data")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.extractContainersVulnerabilityScanInfoFromChannelData"))
		return nil, err
		// Check that the error of channel data is nil
	}
	// Extract data from channelDataWrapper
	data, err := channelDataWrapper.GetData()
	if err != nil {
		err = errors.Wrap(err, "returned error from channel")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.extractContainersVulnerabilityScanInfoFromChannelData"))
		return nil, err
	}
	// Try to cast.
	containerVulnerabilityScanInfo, canConvert := data.([]*contracts.ContainerVulnerabilityScanInfo)
	if !canConvert {
		err = errors.Wrap(utils.CantConvertChannelDataWrapper, "failed to convert ChannelDataWrapper.DataWrapper to []*contracts.ContainerVulnerabilityScanInfo")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.extractContainersVulnerabilityScanInfoFromChannelData"))
		return nil, err
	}
	// Successfully extract data from channel
	tracer.Info("extractContainersVulnerabilityScanInfoFromChannelData finished to extract data", "data", containerVulnerabilityScanInfo)
	return containerVulnerabilityScanInfo, nil
}

// getContainersVulnerabilityScanInfo try to get containers vulnerabilities scan info
func (provider *AzdSecInfoProvider) getContainersVulnerabilityScanInfo(podSpec *admisionrequest.SpecRes, resourceMetadata *admisionrequest.SpecRes) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
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
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(wrappedError, "AzdSecInfoProvider.getContainersVulnerabilityScanInfo"))
		return nil, wrappedError
	}
	//TODO sort result according to original containers order
	return vulnSecInfoContainers, nil
}

// getVulnSecInfoContainers gets vulnSecInfoContainers array with the scan results of the given containers.
// It runs each container scan in parallel and returns only when all the scans are finished and the array is updated
func (provider *AzdSecInfoProvider) getVulnSecInfoContainers(podSpec *corev1.PodSpec, resourceCtx *tag2digest.ResourceContext) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("getVulnSecInfoContainers")

	// Initialize container vuln scan info list
	vulnSecInfoContainers := make([]*contracts.ContainerVulnerabilityScanInfo, 0, len(podSpec.InitContainers)+len(podSpec.Containers))

	if podSpec == nil {
		err := errors.Wrap(utils.NilArgumentError, "failed in AzdSecInfoProvider.getVulnSecInfoContainers. Unexpected: pod.Spec is nil")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.getVulnSecInfoContainers"))
		return nil, err
	}

	// vulnerabilitySecInfoChannel is a channel for (*contracts.ContainerVulnerabilityScanInfo, error)
	vulnerabilitySecInfoChannel := make(chan *utils.ChannelDataWrapper, len(podSpec.InitContainers)+len(podSpec.Containers))
	// Get container vulnerability scan information in parallel
	// Each call send data to channel vulnerabilitySecInfoChannel
	for i := range podSpec.InitContainers {
		go provider.getSingleContainerVulnerabilityScanInfoSyncWrapper(&podSpec.InitContainers[i], resourceCtx, vulnerabilitySecInfoChannel)
	}
	for i := range podSpec.Containers {
		go provider.getSingleContainerVulnerabilityScanInfoSyncWrapper(&podSpec.Containers[i], resourceCtx, vulnerabilitySecInfoChannel)
	}

	for i := 0; i < len(podSpec.InitContainers)+len(podSpec.Containers); i++ { // No deadlock as a result of the loop because the number of receivers is identical to the number of senders
		vulnerabilitySecInfoWrapper, isChannelOpen := <-vulnerabilitySecInfoChannel // Because the channel is buffered all goroutines will finish executing (no goroutine leak)
		if !isChannelOpen {
			err := errors.Wrap(utils.ReadFromClosedChannelError, "failed in AzdSecInfoProvider.getVulnSecInfoContainers. Channel closed unexpectedly")
			tracer.Error(err, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.getVulnSecInfoContainers"))
			return nil, err
		}
		if vulnerabilitySecInfoWrapper == nil {
			err := errors.Wrap(utils.NilArgumentError, "failed in getSingleContainerVulnerabilityScanInfoSync. Unexpected: vulnerabilitySecInfoWrapper is nil")
			tracer.Error(err, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.getVulnSecInfoContainers"))
			return nil, err
		}
		// Check if an error occurred during getSingleContainerVulnerabilityScanInfo
		vulnerabilitySecInfoDataWrapper, err := vulnerabilitySecInfoWrapper.GetData()
		if err != nil {
			err = errors.Wrap(err, "failed in getSingleContainerVulnerabilityScanInfoSync.")
			tracer.Error(err, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.getVulnSecInfoContainers"))
			return nil, err
		}
		// Convert vulnerabilitySecInfoWrapper.DataWrapper to vulnerabilitySecInfo
		vulnerabilitySecInfo, canConvert := vulnerabilitySecInfoDataWrapper.(*contracts.ContainerVulnerabilityScanInfo)
		if !canConvert {
			err := errors.Wrap(utils.CantConvertChannelDataWrapper, "failed to convert ChannelDataWrapper.DataWrapper to *contracts.ContainerVulnerabilityScanInfo")
			tracer.Error(err, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.getVulnSecInfoContainers"))
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
func (provider *AzdSecInfoProvider) getSingleContainerVulnerabilityScanInfoSyncWrapper(container *corev1.Container, resourceCtx *tag2digest.ResourceContext, vulnerabilitySecInfoChannel chan *utils.ChannelDataWrapper) {
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

		// ErrString parsed successfully to known unscanned reason.
		tracer.Info("ErrString from Tag2DigestResolver parsed successfully to known unscanned reason", "ErrString", err, "unscannedReason", unscannedReason)
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
		// ErrString parsed successfully to known unscanned reason.
		tracer.Info("ErrString from ARGDataProvider parsed successfully to known unscanned reason", "ErrString", err, "unscannedReason", unscannedReason)
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

	provider.metricSubmitter.SendMetric(1, azdsecinfometrics.NewContainerVulnScanInfoMetric(scanStatus))

	return info
}

// buildContainerVulnerabilityScanInfoFromResult build the info object from data provided
func (provider *AzdSecInfoProvider) buildContainerVulnerabilityScanInfoUnScannedWithReason(container *corev1.Container, reason contracts.UnscannedReason) *contracts.ContainerVulnerabilityScanInfo {

	info := &contracts.ContainerVulnerabilityScanInfo{
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

	provider.metricSubmitter.SendMetric(1, azdsecinfometrics.NewContainerVulnScanInfoMetricWithUnscannedReason(contracts.Unscanned, reason))
	return info
}

// buildListOfContainerVulnerabilityScanInfoWhenTimeout is method that is called when the GetContainerVulnerabilityScanInfo
// got timeout (GetContainersVulnerabilityScanInfoTimeoutDuration) and returns list with one empty container with
//unscanned status and contracts.GetContainersVulnerabilityScanInfoTimeoutUnscannedReason.
// TODO In public preview, we should add timeout without empty container (bad UX) (should be changed also in the REGO).
func (provider *AzdSecInfoProvider) buildListOfContainerVulnerabilityScanInfoWhenTimeout(podSpec *corev1.PodSpec) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	var containerVulnerabilityScanInfoList[] *contracts.ContainerVulnerabilityScanInfo

	// Iterate over all podSpec containers
	containers := append(podSpec.InitContainers, podSpec.Containers...)
	for _, container := range containers{
		// For each container create info object containing the container name and image name with unscanned status.
		info := &contracts.ContainerVulnerabilityScanInfo{
			Name: container.Name,
			Image: &contracts.Image{
				Name:   container.Image,
				Digest: "",
			},
			ScanStatus:   contracts.Unscanned,
			ScanFindings: nil,
			AdditionalData: map[string]string{
				contracts.UnscannedReasonAnnotationKey: string(contracts.GetContainersVulnerabilityScanInfoTimeoutUnscannedReason),
			},
		}

		// Add info to list
		containerVulnerabilityScanInfoList = append(containerVulnerabilityScanInfoList, info)
	}
	provider.metricSubmitter.SendMetric(1, azdsecinfometrics.NewContainerVulnScanInfoMetricWithUnscannedReason(contracts.Unscanned, contracts.GetContainersVulnerabilityScanInfoTimeoutUnscannedReason))

	return containerVulnerabilityScanInfoList, nil
}

// timeoutEncounteredGetContainersVulnerabilityScanInfo is called when timeout is encountered in GetContainersVulnerabilityScanInfo function
// It checks if it is the first or second time that the request got an error (request is defined as the images of the request).
// If it is the first or the second time, it adds the images to the cache and returns unscanned with metadata.
// If it is the third time or there is an error in the communication with the cache, it returns an error.
// TODO Add tests for this behavior.
func (provider *AzdSecInfoProvider) timeoutEncounteredGetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, podSpecCacheKey string) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("timeoutEncounteredGetContainersVulnerabilityScanInfo")

	// Get the timeoutStatus from cache
	timeoutStatus, err := provider.cacheClient.GetTimeOutStatus(podSpecCacheKey)
	// If an error occurred while getting timeout status from cache return an error because we shouldn't block the pod request
	if err != nil {
		if cache.IsMissingKeyCacheError(err){
			tracer.Info("First timeout. Missing key. Couldn't get TimeOutStatus from cache.")
			return nil, err
		}
		err = errors.Wrap(err, "Timeout encountered but couldn't get previous timeout status from cache: error encountered")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.timeoutEncounteredGetContainersVulnerabilityScanInfo"))
		return nil, err
	}

	// Update timeoutStatus - increase the number of encountered timeouts by 1
	timeoutStatus += 1

	// In case that this is the third time encountered timeout (already encountered twice).
	if timeoutStatus == _numberOfTimeOutEncounteredThreshold {
		err = errors.Wrap(utils.TimeOutError, "Third time that timeout was encountered.")
		tracer.Error(err, "")
		return nil, err
	}
	// If this is the first or second time there is a timeout - set new timeout status in cache and return unscanned and timeout.
	// If an error occurred while setting timeout status in cache return an error because if we can't update timeout status we shouldn't block the pod request.
	if err := provider.cacheClient.SetTimeOutStatusAfterEncounteredTimeout(podSpecCacheKey, timeoutStatus); err != nil {
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.timeoutEncounteredGetContainersVulnerabilityScanInfo"))
		return nil, err
	}

	tracer.Info("GetContainersVulnerabilityScanInfo got timeout - returning unscanned", "timeDurationOfTimeout", provider.getContainersVulnerabilityScanInfoTimeoutDuration)
	return provider.buildListOfContainerVulnerabilityScanInfoWhenTimeout(podSpec)
}

// noTimeoutEncounteredGetContainersVulnerabilityScanInfo getting the scan results, set the results in the cache and reset timeout status
func (provider *AzdSecInfoProvider) noTimeoutEncounteredGetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, chanTimeout chan *utils.ChannelDataWrapper, channelData *utils.ChannelDataWrapper, isChannelOpen bool, podSpecCacheKey string) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("noTimeoutEncounteredGetContainersVulnerabilityScanInfo")
	// Check if channel is open
	if !isChannelOpen {
		err := errors.Wrap(utils.ReadFromClosedChannelError, "Channel closed unexpectedly")
		tracer.Error(err, "")
		return nil, err
	}

	// Try to extract []*contracts.ContainerVulnerabilityScanInfo from channelData
	containerVulnerabilityScanInfo, err := provider.extractContainersVulnerabilityScanInfoFromChannelData(channelData)
	if err != nil {
		err = errors.Wrap(err, "failed to extract []*contracts.ContainerVulnerabilityScanInfo from channel data")
		tracer.Error(err, "")
		provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.noTimeoutEncounteredGetContainersVulnerabilityScanInfo"))
		return nil, err
	}

	// Success path:
	// Update cache to no timeout encountered for this pod
	go func() {
		err := provider.cacheClient.ResetTimeOutInCacheAfterGettingScanResults(podSpecCacheKey)
		if err != nil {
			err = errors.Wrap(err, "failed ResetTimeOutInCacheAfterGettingScanResults")
			tracer.Error(err, "")
			provider.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "AzdSecInfoProvider.noTimeoutEncounteredGetContainersVulnerabilityScanInfo"))
		}
	}()

	close(chanTimeout)
	tracer.Info("GetContainersVulnerabilityScanInfo finished extract []*contracts.ContainerVulnerabilityScanInfo.", "podSpec", podSpec, "ContainerVulnerabilityScanInfo", containerVulnerabilityScanInfo)
	return containerVulnerabilityScanInfo, nil
}
