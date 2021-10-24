package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/acrauth"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	craneerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers/errors"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var (
	_tag2DigestTimeoutErr      = errors.New("Tag2Digest timeout Error")
	_argDataProviderTimeoutErr = errors.New("ARGDataProvider timeout Error")
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

	// tag2DigestTimoutDuration is the duration the tag2digestResolver will try to resolve the image.
	//if the duration will exceed, the program will continue without waiting for the digest.
	//the digest still will be saved in the cache.
	tag2DigestTimoutDuration time.Duration
	// argDataProviderTimoutDuration is the duration the argDataProvider will try to fetch the results of some digest,
	//if the duration will exceed,  the program will continue without waiting for the results.
	//the results still will be saved in the cache.
	argDataProviderTimoutDuration time.Duration
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider,
	argDataProvider arg.IARGDataProvider,
	tag2digestResolver tag2digest.ITag2DigestResolver,
	tag2DigestTimeoutConfiguration *utils.TimeoutConfiguration,
	argDataProviderTimeoutConfiguration *utils.TimeoutConfiguration) *AzdSecInfoProvider {
	return &AzdSecInfoProvider{
		tracerProvider:                instrumentationProvider.GetTracerProvider("AzdSecInfoProvider"),
		metricSubmitter:               instrumentationProvider.GetMetricSubmitter(),
		argDataProvider:               argDataProvider,
		tag2digestResolver:            tag2digestResolver,
		tag2DigestTimoutDuration:      utils.ParseTimeoutConfigurationToDurationOrDefault(tag2DigestTimeoutConfiguration, 1*time.Second /*Default time.Duration*/),
		argDataProviderTimoutDuration: utils.ParseTimeoutConfigurationToDurationOrDefault(argDataProviderTimeoutConfiguration, 2*time.Second /*Default time.Duration*/),
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

	// Convert pull secrets from reference object to strings
	imagePullSecrets := make([]string, 0, len(podSpec.ImagePullSecrets))
	// element is a LocalObjectReference{Name}
	for _, element := range podSpec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, element.Name)
	}

	// Build resource deployment context
	resourceCtx := tag2digest.NewResourceContext(resourceMetadata.Namespace, imagePullSecrets, podSpec.ServiceAccountName)
	tracer.Info("resourceCtx", "resourceCtx", resourceCtx)

	// Initialize container vuln scan info list
	vulnSecInfoContainers := make([]*contracts.ContainerVulnerabilityScanInfo, 0, len(podSpec.InitContainers)+len(podSpec.Containers))

	// Get container vulnerability scan information for init containers
	for _, container := range podSpec.InitContainers {
		vulnerabilitySecInfo, err := provider.getSingleContainerVulnerabilityScanInfo(&container, resourceCtx)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo Init containers")
			tracer.Error(wrappedError, "")
			return nil, wrappedError
		}

		// Add it to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}

	// Get container vulnerability scan information for containers
	for _, container := range podSpec.Containers {
		vulnerabilitySecInfo, err := provider.getSingleContainerVulnerabilityScanInfo(&container, resourceCtx)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo Containers")
			tracer.Error(wrappedError, "")
			return nil, wrappedError
		}

		// Add it to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}
	return vulnSecInfoContainers, nil
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

	// Checks if the image registry is not ACR.
	if !registryutils.IsRegistryEndpointACR(imageRef.Registry()) {
		tracer.Info("Image from another registry than ACR received", "Registry", imageRef.Registry())
		return buildContainerVulnerabilityScanInfoUnScannedWithReason(container, contracts.ImageIsNotInACRRegistryUnscannedReason), nil
	}

	digest, err := provider.callTag2DigestResolveWithTimeout(imageRef, resourceCtx)
	if err != nil {
		// TODO wait until @maayaan merge his PR and then add tests for this method. ( Maayan already created IAZdSecInfoProvider mock)
		return provider.tryResolveErrorToUnscannedWithReason(container, err)
	}

	scanStatus, scanFindings, err := provider.callARGDataProviderWithTimeout(imageRef, digest)
	if err != nil {
		// TODO wait until @maayaan merge his PR and then add tests for this method. ( Maayan already created IAZdSecInfoProvider mock)
		return provider.tryResolveErrorToUnscannedWithReason(container, err)
	}

	tracer.Info("results from ARG data provider", "scanStatus", scanStatus, "scanFindings", scanFindings)
	// Build scan info from provided scan results
	info := buildContainerVulnerabilityScanInfoFromResult(container, digest, scanStatus, scanFindings)

	return info, nil
}

// callTag2DigestResolveWithTimeout wraps tag2digestResolver.Resolve with timeout.
// it runs this function in parallel to thread that sleeps constant time and returns the result of the first thread that is finish.
// We use it in order to ensure that if we couldn't get the digest in the first time, then we will return unscanned
// status with timeout reason and the thread that tries to resolve the digest ll keep running and insert the results to the cache.
// TODO @tomerwinberger, do you think that we should generalize this method (it will be changed between functions with diff
// 	returns values - for example, this and callArgDataProviderWithTimeout are needed 2 different methods).
// TODO Add tests to this method (once @maayan will merge his PR)
func (provider *AzdSecInfoProvider) callTag2DigestResolveWithTimeout(imageRef registry.IImageReference, resourceCtx *tag2digest.ResourceContext) (digest string, err error) {
	// Get image digest
	chanTimeout := make(chan int, 1)
	go func() {
		digest, err = provider.tag2digestResolver.Resolve(imageRef, resourceCtx)
		chanTimeout <- 0 // The value that we insert to the channel is not relevant.
	}()

	// Choose the first thread that is finish.
	select {
	case _ = <-chanTimeout:
		return digest, err
	case <-time.After(provider.tag2DigestTimoutDuration):
		return "", _tag2DigestTimeoutErr
	}
}

// callARGDataProviderWithTimeout wraps argDataProvider.GetImageVulnerabilityScanResults with timeout.
// it runs this function in parallel to thread that sleeps constant time and returns the result of the first thread that is finishes.
// We use it in order to verify that if we couldn't get the results in the first time, then we will return unscanned
//status with timeout reason and the thread that tries to fetch the results will keep running and insert the results to the cache.
// TODO Add tests to this method (once @maayan will merge his PR)
func (provider *AzdSecInfoProvider) callARGDataProviderWithTimeout(imageRef registry.IImageReference, digest string) (scanStatus contracts.ScanStatus, scanFindings []*contracts.ScanFinding, err error) {
	// Get image digest
	chanTimeout := make(chan int, 1)
	go func() {
		scanStatus, scanFindings, err = provider.argDataProvider.GetImageVulnerabilityScanResults(imageRef.Registry(), imageRef.Repository(), digest)
		chanTimeout <- 0 // The value that we insert to the channel is not relevant.
	}()

	// Choose the first thread that is finish.
	select {
	case _ = <-chanTimeout:
		return scanStatus, scanFindings, err
	case <-time.After(provider.argDataProviderTimoutDuration):
		return "", nil, _argDataProviderTimeoutErr
	}
}

// tryResolveErrorToUnscannedWithReason gets an error the container that the error encountered and returns the info and error according to the type of the error.
// If the error is expected error (e.g. image is not exists while trying to resolve the digest, unauthorized to arg) then
// this function create new contracts.ContainerVulnerabilityScanInfo that that status is unscanned and add in the additional metadata field
// the reason for unscanned - for example contracts.ImageDoesNotExistUnscannedReason.
// If the function doesn't recognize the error, then it returns nil, err
func (provider *AzdSecInfoProvider) tryResolveErrorToUnscannedWithReason(container *corev1.Container, err error) (*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("tryResolveErrorToUnscannedWithReason")
	// Check if the err is known error:
	// TODO - sort the errors (most frequent should be first error)
	// TODO add metrics for this method.
	// Checks if the error is image is not found error
	if utils.IsErrorIsTypeOf(err, craneerrors.GetImageIsNotFoundErrType()) { // Checks if the error is Image DoesNot Exist
		return buildContainerVulnerabilityScanInfoUnScannedWithReason(container, contracts.ImageDoesNotExistUnscannedReason), nil
		// Checks if th error is unauthorized error
	} else if utils.IsErrorIsTypeOf(err, craneerrors.GetUnauthorizedErrType()) { // Checks if the error is unauthorized
		return buildContainerVulnerabilityScanInfoUnScannedWithReason(container, contracts.RegistryUnauthorizedUnscannedReason), nil
		// Checks if the error is NoSuchHost - it means that the registry is not found.
	} else if acrauth.IsNoSuchHostErr(err) {
		return buildContainerVulnerabilityScanInfoUnScannedWithReason(container, contracts.RegistryDoesNotExistUnscannedReason), nil
		// Checks if there was timeout while trying to resolve digest using Tag2Digest
	} else if errors.Is(err, _tag2DigestTimeoutErr) {
		return buildContainerVulnerabilityScanInfoUnScannedWithReason(container, contracts.Tag2DigestTimeoutUnscannedReason), nil
		// Checks if there was timeout while trying to get image vulnerabilities result using ArgDataProvider.
	} else if errors.Is(err, _argDataProviderTimeoutErr) {
		return buildContainerVulnerabilityScanInfoUnScannedWithReason(container, contracts.ARGDataProviderTimeoutUnscannedReason), nil
		// Got unexpected error
	} else {
		err = errors.Wrap(err, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo.tag2digestResolver.Resolve")
		tracer.Error(err, "Unexpected error")
		return nil, err
	}
}

// buildContainerVulnerabilityScanInfoFromResult build the info object from data provided
func buildContainerVulnerabilityScanInfoFromResult(container *corev1.Container, digest string, scanStatus contracts.ScanStatus, scanFindigs []*contracts.ScanFinding) *contracts.ContainerVulnerabilityScanInfo {
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
func buildContainerVulnerabilityScanInfoUnScannedWithReason(container *corev1.Container, reason contracts.UnscannedReason) *contracts.ContainerVulnerabilityScanInfo {
	var additionalData = map[string]string{
		contracts.UnscannedReasonAnnotationKey: string(reason),
	}

	info := &contracts.ContainerVulnerabilityScanInfo{
		Name: container.Name,
		Image: &contracts.Image{
			Name:   container.Image,
			Digest: "",
		},
		ScanStatus:     contracts.Unscanned,
		ScanFindings:   nil,
		AdditionalData: additionalData,
	}
	return info
}
