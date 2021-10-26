package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
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
	"time"
)

const (
	_defaultTimeDurationGetContainersVulnerabilityScanInfo = 2500 * time.Millisecond // 2.5 seconds - can't multiply float in seconds
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
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider,
	argDataProvider arg.IARGDataProvider,
	tag2digestResolver tag2digest.ITag2DigestResolver,
	GetContainersVulnerabilityScanInfoTimeoutDuration *utils.TimeoutConfiguration) *AzdSecInfoProvider {
	return &AzdSecInfoProvider{
		tracerProvider:     instrumentationProvider.GetTracerProvider("AzdSecInfoProvider"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		argDataProvider:    argDataProvider,
		tag2digestResolver: tag2digestResolver,
		getContainersVulnerabilityScanInfoTimeoutDuration: utils.ParseTimeoutConfigurationToDurationOrDefault(GetContainersVulnerabilityScanInfoTimeoutDuration, _defaultTimeDurationGetContainersVulnerabilityScanInfo /*Default time.Duration*/),
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

	// Wrap  getContainersVulnerabilityScanInfo with timeout - return the result from the first thread that is finish
	// TODO use Maayan's wrapper.
	chanTimeout := make(chan bool, 1)
	go func() {
		containerVulnerabilityScanInfo, err = provider.getContainersVulnerabilityScanInfo(podSpec, resourceMetadata, resourceKind)
		chanTimeout <- true // The value that we insert to the channel is not relevant.
	}()

	// Choose the first thread that is finish.
	select {
	// No timeout case:
	case _ = <-chanTimeout:
		close(chanTimeout)
		if err != nil {
			err = errors.Wrap(err, "GetContainersVulnerabilityScanInfo finished to fetch resulsts and encountered with error.")
			tracer.Error(err, "")
			return nil, err
		}
		return containerVulnerabilityScanInfo, nil

	// Timeout case:
	case <-time.After(provider.getContainersVulnerabilityScanInfoTimeoutDuration):
		//TODO close channel!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		//TODO Implement cache that will check if it is the first time that got timeout error or it is the second time (if it is the second time then it should return error and don't add unscanned metadata!!)
		tracer.Info("GetContainersVulnerabilityScanInfo got timeout - returning unscanned", "timeDurationOfTimeout", provider.getContainersVulnerabilityScanInfoTimeoutDuration)
		return provider.buildListOfContainerVulnerabilityScanInfoWhenTimeout()
	}
}

func (provider *AzdSecInfoProvider) getContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("getContainersVulnerabilityScanInfo")

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
	// Check if the err is known error:
	// TODO - sort the errors (most frequent should be first error)
	// TODO add metrics for this method.
	// Checks if the error is image is not found error
	cause := errors.Cause(err)
	tracer.Info("Extracted the cause of the error", "err", err, "cause", cause)
	// Try to parse the cause of the error to known error -> if true, resolve to unscanned reason.
	switch cause.(type) {
	case *registryerrors.ImageIsNotFoundErr: // Checks if the error is Image DoesNot Exist
		unscannedReason := contracts.ImageDoesNotExistUnscannedReason
		return &unscannedReason, true
	case *registryerrors.UnauthorizedErr: // Checks if the error is unauthorized
		unscannedReason := contracts.RegistryUnauthorizedUnscannedReason
		return &unscannedReason, true
	case *registryerrors.RegistryIsNotFoundErr: // Checks if the error is NoSuchHost - it means that the registry is not found.
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
