package azdsecinfo

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {
	// GetContainersVulnerabilityScanInfo receives pod template spec containing containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata *metav1.ObjectMeta, resourceKind *metav1.TypeMeta) ([]*contracts.ContainerVulnerabilityScanInfo, error)
}

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
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argDataProvider arg.IARGDataProvider, tag2digestResolver tag2digest.ITag2DigestResolver) *AzdSecInfoProvider {
	return &AzdSecInfoProvider{
		tracerProvider:     instrumentationProvider.GetTracerProvider("AzdSecInfoProvider"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		argDataProvider:    argDataProvider,
		tag2digestResolver: tag2digestResolver,
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
	vulnSecInfoContainers := make([]*contracts.ContainerVulnerabilityScanInfo, 0, len(podSpec.InitContainers) + len(podSpec.Containers))

	// insert container vulnerability scan information for init containers to vulnSecInfoContainers
	initContainersErr := provider.updateVulnSecInfoContainers(&podSpec.InitContainers, &vulnSecInfoContainers, resourceCtx)
	if initContainersErr != nil {
		wrappedError := errors.Wrap(initContainersErr, "Handler failed to updateVulnSecInfoContainers on Init containers")
		tracer.Error(wrappedError, "")
		return nil, wrappedError
	}

	// insert container vulnerability scan information for containers to vulnSecInfoContainers
	containersErr := provider.updateVulnSecInfoContainers(&podSpec.Containers, &vulnSecInfoContainers, resourceCtx)
	if containersErr != nil {
		wrappedError := errors.Wrap(containersErr, "Handler failed to updateVulnSecInfoContainers on Containers")
		tracer.Error(wrappedError, "")
		return nil, wrappedError
	}

	return vulnSecInfoContainers, nil
}

// updateVulnSecInfoContainers is updating vulnSecInfoContainers array with the scan results of the given containers .
// It runs each container scan in parallel and returns only when all the scans are finished and the array is updated
func (provider *AzdSecInfoProvider) updateVulnSecInfoContainers(containers *[]corev1.Container, vulnSecInfoContainers *[]*contracts.ContainerVulnerabilityScanInfo, resourceCtx *tag2digest.ResourceContext) error {
	tracer := provider.tracerProvider.GetTracer("updateVulnSecInfoContainers")
	azdSecInfoProviderSync := NewAzdSecInfoProviderSync()
	defer azdSecInfoProviderSync.cancelScans() // Release resources even if no errors

	// Get container vulnerability scan information in parallel
	// Each call send data to channel (azdSecInfoProviderSync.vulnerabilitySecInfoChannel)
	for _, container := range *containers {
		go provider.getSingleContainerVulnerabilityScanInfoSyncWrapper(container, resourceCtx, azdSecInfoProviderSync)
	}

	var vulnerabilitySecInfoWrapper *wrappers.ContainerVulnerabilityScanInfoWrapper = nil
	err := error(nil)
	for range *containers { // No deadlock as a result of the loop because the number of receivers is identical to the number of senders
		vulnerabilitySecInfoWrapper = <- azdSecInfoProviderSync.vulnerabilitySecInfoChannel // extract all data from channel to make sure there is no goroutine leak
		// If vulnerabilitySecInfoWrapper is nil it means that an error occurred previously - skip this section and wait for all goroutines to finish
		if vulnerabilitySecInfoWrapper != nil{
			if vulnerabilitySecInfoWrapper.Err != nil {
				wrappedError := errors.Wrap(vulnerabilitySecInfoWrapper.Err, "Handler failed in getSingleContainerVulnerabilityScanInfoSyncWrapper")
				tracer.Error(wrappedError, "")
				err = wrappedError
			} else {
			// Add it to slice
			*vulnSecInfoContainers = append(*vulnSecInfoContainers, vulnerabilitySecInfoWrapper.VulnerabilitySecInfo)
			}
		}
	}
	return err
}

//getSingleContainerVulnerabilityScanInfoSyncWrapper wrap getSingleContainerVulnerabilityScanInfo.
// It adds the ability to stop other goroutines from starting new jobs (getSingleContainerVulnerabilityScanInfo) in case one goroutine fail
func (provider *AzdSecInfoProvider) getSingleContainerVulnerabilityScanInfoSyncWrapper(container corev1.Container, resourceCtx *tag2digest.ResourceContext,  azdSecInfoProviderSync *AzdSecInfoProviderSync){
	select {
	// In case canceled has been called (Error occurred in another goroutine). Exit this goroutine.
	case <-azdSecInfoProviderSync.AzdSecInfoProviderCtx.Done():
		azdSecInfoProviderSync.vulnerabilitySecInfoChannel <- nil // Send nil to channel to avoid deadlock
		return
	default:
		vulnerabilitySecInfo, err := provider.getSingleContainerVulnerabilityScanInfo(&container, resourceCtx)
		if err != nil {
			// Call cancelScans to stop other goroutines from starting new jobs
			azdSecInfoProviderSync.cancelScans()
		}
		azdSecInfoProviderSync.vulnerabilitySecInfoChannel <- wrappers.NewContainerVulnerabilityScanInfoWrapper(vulnerabilitySecInfo, err)
	}
}

// getSingleContainerVulnerabilityScanInfo receives a container and it's belogned deployed resource context, and returns fetched ContainerVulnerabilityScanInfo
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

	//Set default values
	var scanStatus = contracts.Unscanned
	var scanFindings []*contracts.ScanFinding = nil
	var digest = ""
	var additionalData = make(map[string]string)

	// Checks if the image registry is not ACR.
	if !registryutils.IsRegistryEndpointACR(imageRef.Registry()) {
		tracer.Info("Image from another registry than ACR received", "Registry", imageRef.Registry())
		additionalData["unscannedReason"] = fmt.Sprintf("Registry of image \"%v\" is not an ACR", imageRef.Registry())
		info := buildContainerVulnerabilityScanInfoFromResult(container, digest, scanStatus, scanFindings, additionalData)
		return info, nil
	}

	// Get image digest
	digest, err = provider.tag2digestResolver.Resolve(imageRef, resourceCtx)
	if err != nil {
		err = errors.Wrap(err, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo.tag2digestResolver.Resolve")
		tracer.Error(err, "")

		// TODO support digest does not exists in registry or unauthorized to not fail...
		// Add indication in tag2digest resolver on unauthorized to image to set as unscanned.
		return nil, err
	}

	// 	Get image scan results for image
	scanStatus, scanFindings, err = provider.argDataProvider.GetImageVulnerabilityScanResults(imageRef.Registry(), imageRef.Repository(), digest)
	if err != nil {
		err = errors.Wrap(err, "AzdSecInfoProvider.getContainerVulnerabilityScanResults")
		tracer.Error(err, "")
		return nil, err
	}
	tracer.Info("results from ARG data provider", "scanStatus", scanStatus, "scanFindings", scanFindings)

	// Build scan info from provided scan results
	info := buildContainerVulnerabilityScanInfoFromResult(container, digest, scanStatus, scanFindings, additionalData)

	return info, nil
}

// buildContainerVulnerabilityScanInfoFromResult build the info object from data provided
func buildContainerVulnerabilityScanInfoFromResult(container *corev1.Container, digest string, scanStatus contracts.ScanStatus, scanFindigs []*contracts.ScanFinding, additionalData map[string]string) *contracts.ContainerVulnerabilityScanInfo {
	info := &contracts.ContainerVulnerabilityScanInfo{
		Name: container.Name,
		Image: &contracts.Image{
			Name:   container.Image,
			Digest: digest,
		},
		ScanStatus:     scanStatus,
		ScanFindings:   scanFindigs,
		AdditionalData: additionalData,
	}
	return info
}
