package azdsecinfo

import (
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
	// configuration of AzdSecInfoProviderConfiguration.
	configuration *AzdSecInfoProviderConfiguration
}

// AzdSecInfoProviderConfiguration is configuration data for AzdSecInfoProvider
type AzdSecInfoProviderConfiguration struct {
	// CacheExpirationTimeTimeout is the expiration time for timout.
	CacheExpirationTimeTimeout time.Duration
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider,
	argDataProvider arg.IARGDataProvider,
	tag2digestResolver tag2digest.ITag2DigestResolver,
	GetContainersVulnerabilityScanInfoTimeoutDuration *utils.TimeoutConfiguration,
	azdSecInfoProviderConfiguration *AzdSecInfoProviderConfiguration,
	cacheClient cache.ICacheClient) *AzdSecInfoProvider {

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
		configuration: azdSecInfoProviderConfiguration,
		cacheClient:   cacheClient,
	}
}

// GetContainersVulnerabilityScanInfo receives api-resource pod spec containing containers, resource deployed metadata and kind
// Function returns evaluated ContainerVulnerabilityScanInfo for pod spec's container list (pod spec can be related to template of any resource creates pods eventually)
func (provider *AzdSecInfoProvider) GetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta) (containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) {
	tracer := provider.tracerProvider.GetTracer("GetContainersVulnerabilityScanInfo")
	tracer.Info("Received:", "podSpec", podSpec, "resourceMetadata", resourceMetadata, "resourceKind", resourceKind)

	// Argument validation
	if podSpec == nil || resourceMetadata == nil || resourceKind == nil {
		err = errors.Wrap(utils.NilArgumentError, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		tracer.Error(err, "")
		return nil, err
	}

	// Try to get containers vulnerabilities in diff thread.
	chanTimeout := make(chan *utils.ChannelDataWrapper, 1)
	go provider.getContainersVulnerabilityScanInfoSyncWrapper(podSpec, resourceMetadata, resourceKind, chanTimeout)

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
		containerVulnerabilityScanInfo, err = provider.extractContainersVulnerabilityScanInfoFromChannelData(channelData)
		if err != nil {
			err = errors.Wrap(err, "failed to extract []*contracts.ContainerVulnerabilityScanInfo from channel data")
			tracer.Error(err, "")
			return nil, err
		}

		// Success path:
		close(chanTimeout)
		tracer.Info("GetContainersVulnerabilityScanInfo finished extract []*contracts.ContainerVulnerabilityScanInfo.", "podSpec", podSpec, "containerVulnerabilityScanInfo", containerVulnerabilityScanInfo)
		return containerVulnerabilityScanInfo, nil

	// Timeout case:
	case <-time.After(provider.getContainersVulnerabilityScanInfoTimeoutDuration):
		//TODO Implement cache that will check if it  the first time that got timeout error or it  the second time (if it  the second time then it should return error and don't add unscanned metadata!!)

		return provider.timeoutEncounteredGetContainersVulnerabilityScanInfo(podSpec)
	}
}

// getContainersVulnerabilityScanInfoSyncWrapper runs getContainersVulnerabilityScanInfo and insert the result into the channel.
func (provider *AzdSecInfoProvider) getContainersVulnerabilityScanInfoSyncWrapper(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta, chanTimeout chan *utils.ChannelDataWrapper) {
	tracer := provider.tracerProvider.GetTracer("getContainersVulnerabilityScanInfoSyncWrapper")
	containerVulnerabilityScanInfo, err := provider.getContainersVulnerabilityScanInfo(podSpec, resourceMetadata, resourceKind)
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
		unscannedReason, isErrParsedToUnscannedReason := provider.tryParseErrToUnscannedWithReason(err)
		if !isErrParsedToUnscannedReason {
			err = errors.Wrap(err, "Unexpected error while trying to resolve digest")
			tracer.Error(err, "")
			return nil, err
		}

		// err parsed successfully to known unscanned reason.
		tracer.Info("err from Tag2DigestResolver parsed successfully to known unscanned reason", "err", err, "unscannedReason", unscannedReason)
		return provider.buildContainerVulnerabilityScanInfoUnScannedWithReason(container, *unscannedReason), nil
	}

	scanStatus, scanFindings, err := provider.argDataProvider.GetImageVulnerabilityScanResults(imageRef.Registry(), imageRef.Repository(), digest)
	if err != nil {
		// TODO wait until @maayaan merge his PR and then add tests for this method. ( Maayan already created IAZdSecInfoProvider mock)
		unscannedReason, isErrParsedToUnscannedReason := provider.tryParseErrToUnscannedWithReason(err)
		if !isErrParsedToUnscannedReason {
			err = errors.Wrap(err, "Unexpected error while trying to get results from ARGDataProvider")
			tracer.Error(err, "")
			return nil, err
		}
		// err parsed successfully to known unscanned reason.
		tracer.Info("err from ARGDataProvider parsed successfully to known unscanned reason", "err", err, "unscannedReason", unscannedReason)
		return provider.buildContainerVulnerabilityScanInfoUnScannedWithReason(container, *unscannedReason), nil
	}

	tracer.Info("results from ARG data provider", "scanStatus", scanStatus, "scanFindings", scanFindings)
	// Build scan info from provided scan results
	info := provider.buildContainerVulnerabilityScanInfoFromResult(container, digest, scanStatus, scanFindings)

	return info, nil
}

// tryParseErrToUnscannedWithReason gets an error the container that the error encountered and returns the info and error according to the type of the error.
// If the error is expected error (e.g. image is not exists while trying to resolve the digest, unauthorized to arg) then
// this function create new contracts.ContainerVulnerabilityScanInfo that that status is unscanned and add in the additional metadata field
// the reason for unscanned - for example contracts.ImageDoesNotExistUnscannedReason.
// If the function doesn't recognize the error, then it returns nil, err
func (provider *AzdSecInfoProvider) tryParseErrToUnscannedWithReason(err error) (*contracts.UnscannedReason, bool) {
	tracer := provider.tracerProvider.GetTracer("tryParseErrToUnscannedWithReason")
	// Check if the err  known error:
	// TODO - sort the errors (most frequent should be first error)
	// TODO add metrics for this method.
	// Checks if the error  image  not found error
	cause := errors.Cause(err)
	tracer.Info("Extracted the cause of the error", "err", err, "cause", cause)
	// Try to parse the cause of the error to known error -> if true, resolve to unscanned reason.
	switch cause.(type) {
	case *registryerrors.ImageIsNotFoundErr: // Checks if the error  Image DoesNot Exist
		unscannedReason := contracts.ImageDoesNotExistUnscannedReason
		return &unscannedReason, true
	case *registryerrors.UnauthorizedErr: // Checks if the error  unauthorized
		unscannedReason := contracts.RegistryUnauthorizedUnscannedReason
		return &unscannedReason, true
	case *registryerrors.RegistryIsNotFoundErr: // Checks if the error  NoSuchHost - it means that the registry  not found.
		unscannedReason := contracts.RegistryDoesNotExistUnscannedReason
		return &unscannedReason, true
	default: // Unexpected error
		return nil, false
	}
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
func (provider *AzdSecInfoProvider) timeoutEncounteredGetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("timeoutEncounteredGetContainersVulnerabilityScanInfo")
	images := utils.ExtractImagesFromPodSpec(podSpec)
	// Sort the array - it is important for the cache to be sorted.
	sort.Strings(images)
	concatenatedImages := strings.Join(images, ",")

	// Check if the concatenatedImages is already in cache
	_, err := provider.cacheClient.Get(concatenatedImages)

	// In case that the concatenatedImages is already in cache.
	if err == nil {
		// TODO Add metric for number of timeouts.
		err = errors.Wrap(err, "Second time that timeout was encountered.")
		tracer.Error(err, "")
		return nil, err
	}

	// Check if the error is due to missing key.
	_, isFirstTimeOfTimeout := err.(*cache.MissingKeyCacheError)
	if !isFirstTimeOfTimeout {
		// TODO Add metric new error encountered
		err = errors.Wrap(err, "error encountered while trying to get result of timeout from cache.")
		tracer.Error(err, "")
		return nil, err
	}

	// In case that we got missing key error, it means that the concatenatedImages is not in cache - first time of timeout.
	// Add to cache and return unscanned and timeout.
	if err := provider.cacheClient.Set(concatenatedImages, "1", provider.configuration.CacheExpirationTimeTimeout); err != nil {
		// TODO Add metric new error encountered
		err = errors.Wrap(err, "error encountered while trying to set new timeout in cache.")
		tracer.Error(err, "")
		return nil, err
	}

	tracer.Info("GetContainersVulnerabilityScanInfo got timeout - returning unscanned", "timeDurationOfTimeout", provider.getContainersVulnerabilityScanInfoTimeoutDuration)
	return provider.buildListOfContainerVulnerabilityScanInfoWhenTimeout()
}
