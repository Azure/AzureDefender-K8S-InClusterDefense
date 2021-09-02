package azdsecinfo

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registryauth "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	// _azureContainerRegistrySuffix is the suffix of ACR public (todo extract per env maybe?)
	_azureContainerRegistrySuffix = ".azurecr.io"
)

var (
	// _containerNullError Null container error
	_containerNullError = errors.New("Container received is null")
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {
	// GetContainersVulnerabilityScanInfo receives pod template spec containing containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata  *metav1.ObjectMeta, resourceKind *metav1.TypeMeta ) ([]*contracts.ContainerVulnerabilityScanInfo, error)
}

// AzdSecInfoProvider represents default implementation of IAzdSecInfoProvider interface
type AzdSecInfoProvider struct {
	//tracerProvider is tracer provider of AzdSecInfoProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter is metric submitter of AzdSecInfoProvider
	metricSubmitter metric.IMetricSubmitter
	// argDataProvider is the ARG provider which provides any ARG data
	argDataProvider arg.IARGDataProvider
	// registryClient is the client of the registry which is used to resolve image's digest
	registryClient registry.IRegistryClient
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argDataProvider arg.IARGDataProvider, registryClient registry.IRegistryClient) *AzdSecInfoProvider {
	return &AzdSecInfoProvider{
		tracerProvider:  instrumentationProvider.GetTracerProvider("AzdSecInfoProvider"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		argDataProvider: argDataProvider,
		registryClient:  registryClient,
	}
}


// GetContainersVulnerabilityScanInfo receives api-resource pod spec containing containers list, and returns pod spec containers fetched ContainerVulnerabilityScanInfo
// Pod template spec represents contianers to be deployed for all api-resources
func (provider *AzdSecInfoProvider) GetContainersVulnerabilityScanInfo(podSpec *corev1.PodSpec, resourceMetadata  *metav1.ObjectMeta, resourceKind *metav1.TypeMeta ) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("GetContainersVulnerabilityScanInfo")
	tracer.Info("Received:", "podSpec", podSpec, "resourceMetadata", resourceMetadata, "resourceKind", resourceKind)

	if podSpec == nil || resourceMetadata == nil || resourceKind == nil {
		tracer.Error(_containerNullError, "")
		return nil, _containerNullError
	}

	namespace := resourceMetadata.Namespace
	imagePullSecrets := []string{}
	for _, element := range podSpec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, element.Name)
	}

	registryAuthCtx := &registryauth.AuthContext{Namespace: namespace, ImagePullSecrets: imagePullSecrets, ServiceAccountName: podSpec.ServiceAccountName}
	tracer.Info("registryAuthCtx:", "registryAuthCtx",registryAuthCtx)

	vulnSecInfoContainers := []*contracts.ContainerVulnerabilityScanInfo{}
	for _, container := range podSpec.InitContainers {

		// Get container vulnerability scan information for init containers
		vulnerabilitySecInfo, err := provider.getSingleContainerVulnerabilityScanInfo(&container, registryAuthCtx)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo Init containers")
			tracer.Error(wrappedError, "")
			return nil, wrappedError
		}

		// Add it to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}

	for _, container := range podSpec.Containers {

		// Get container vulnerability scan information for containers
		vulnerabilitySecInfo, err := provider.getSingleContainerVulnerabilityScanInfo(&container, registryAuthCtx)
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


// getSingleContainerVulnerabilityScanInfo receives container , and returns fetched ContainerVulnerabilityScanInfo
func (provider *AzdSecInfoProvider) getSingleContainerVulnerabilityScanInfo(container *corev1.Container, registryAuthCtx *registryauth.AuthContext) (*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("getSingleContainerVulnerabilityScanInfo")
	if container == nil {
		tracer.Error(_containerNullError, "")
		return nil, _containerNullError
	}

	tracer.Info("Received:", "container image ref", container.Image, "registryAuthCtx", registryAuthCtx)

	// Extracts image context
	imageRefContext, err := registry.ExtractImageRefContext(container.Image)
	if err != nil {
		err = errors.Wrap(err, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo.registry.ExtractImageRefContext")
		tracer.Error(err, "")
		return nil, err
	}
	tracer.Info("Container image ref extracted context", "context", imageRefContext)

	//Set default values
	var scanStatus = contracts.Unscanned
	var scanFindings []*contracts.ScanFinding = nil
	var digest = ""
	var additionalData = make(map[string]string)

	// Checks if the image registry is not ACR.
	if !strings.HasSuffix(strings.ToLower(imageRefContext.Registry), _azureContainerRegistrySuffix) {
		tracer.Info("Image from another registry than ACR received", "Registry", imageRefContext.Registry)
		additionalData["unscannedReason"] = fmt.Sprintf("Registry of image \"%v\" is not an ACR", imageRefContext.Registry)
		info := buildContainerVulnerabilityScanInfoFromResult(container, digest, scanStatus, scanFindings, additionalData)
		return info, nil
	}

	// Set auth to this registry context
	registryAuthCtx.RegistryEndpoint = imageRefContext.Registry

	// Get image digest
	digest, err = provider.registryClient.GetDigest(container.Image, registryAuthCtx)
	if err != nil {
		err = errors.Wrap(err, "AzdSecInfoProvider.GetContainersVulnerabilityScanInfo.registry.GetDigest")
		tracer.Error(err, "")

		// TODO support digest does not exists in registry or unauthorized to not fail...
		return nil, err
	}

	// Tries to get image scan results for image
	scanStatus, scanFindings, err = provider.argDataProvider.GetImageVulnerabilityScanResults(imageRefContext.Registry, imageRefContext.Repository, digest)
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
