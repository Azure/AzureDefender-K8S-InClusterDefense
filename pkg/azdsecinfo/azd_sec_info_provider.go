package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

const (
	_azureContainerRegistrySuffix = "azurecr.io"
)

var (
	_containerNullError = errors.New("Container received is null")
)

// IAzdSecInfoProvider represents interface for providing azure defender security information
type IAzdSecInfoProvider interface {

	// GetContainerVulnerabilityScanInfo receives containers list, and returns their fetched ContainersVulnerabilityScanInfo
	GetContainerVulnerabilityScanInfo(*corev1.Container) (*contracts.ContainerVulnerabilityScanInfo, error)
}

// AzdSecInfoProvider represents default implementation of IAzdSecInfoProvider interface
type AzdSecInfoProvider struct {
	tracerProvider  trace.ITracerProvider
	metricSubmitter metric.IMetricSubmitter
	argDataProvider arg.IARGDataProvider
	registryClient  registry.IRegistryClient
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

// GetContainerVulnerabilityScanInfo receives containers list, and returns their fetched ContainerVulnerabilityScanInfo
func (provider *AzdSecInfoProvider) GetContainerVulnerabilityScanInfo(container *corev1.Container) (*contracts.ContainerVulnerabilityScanInfo, error) {
	tracer := provider.tracerProvider.GetTracer("GetContainerVulnerabilityScanInfo")
	if container == nil {
		tracer.Error(_containerNullError, "")
		return nil, _containerNullError
	}

	tracer.Info("Container image ref", "container image ref", container.Image)

	imageRefContexct, err := registry.ExtractImageRefContext(container.Image)
	if err != nil {
		err = errors.Wrap(err, "AzdSecInfoProvider.GetContainerVulnerabilityScanInfo.registry.ExtractImageRefContext")
		tracer.Error(err, "")
		return nil, err
	}

	tracer.Info("Container image ref extracted context", "context", imageRefContexct)

	var scanStatus = contracts.Unscanned
	var scanFindings []*contracts.ScanFinding = nil
	var digest = ""

	if strings.HasSuffix(strings.ToLower(imageRefContexct.Registry), _azureContainerRegistrySuffix) {
		digest, err := provider.registryClient.GetDigest(container.Image)
		if err != nil {
			err = errors.Wrap(err, "AzdSecInfoProvider.GetContainerVulnerabilityScanInfo.registry.GetDigest")
			tracer.Error(err, "")

			// TODO support digest does not exists in registry or unauthorized to not fail...
			return nil, err

		}

		scanStatus, scanFindings, err = provider.getContainerVulnerbilityScanResults(imageRefContexct, digest)
		if err != nil {
			err = errors.Wrap(err, "AzdSecInfoProvider.getContainerVulnerbilityScanResults")
			tracer.Error(err, "")
			return nil, err
		}

	}

	info := buildContainerVulnerabilityScanInfoFromResult(container, digest, scanStatus, scanFindings)

	return info, nil
}

func (provider *AzdSecInfoProvider) getContainerVulnerbilityScanResults(imageRefContexct *registry.ImageRefContext, digest string) (scanStatus contracts.ScanStatus, scanFindings []*contracts.ScanFinding, err error) {
	tracer := provider.tracerProvider.GetTracer("getContainerVulnerbilityScanResults")

	isScanned, results, err := provider.argDataProvider.TryGetImageVulnerabilityScanResults(imageRefContexct.Registry, imageRefContexct.Repository, digest)
	if err != nil {
		err = errors.Wrap(err, "Error from getContainerVulnerbilityScanResults.argDataProvider.TryGetImageVulnerabilityScanResults")
		tracer.Error(err, "")
		return "", nil, err
	}

	if isScanned {
		if len(results) == 0 {
			scanStatus = contracts.HealthyScan
			scanFindings = []*contracts.ScanFinding{}
		} else {
			scanStatus = contracts.UnhealthyScan
			scanFindings = results
		}
	} else {
		scanStatus = contracts.Unscanned
	}
	return scanStatus, scanFindings, nil
}

func buildContainerVulnerabilityScanInfoFromResult(container *corev1.Container, digest string, scanStatus contracts.ScanStatus, scanFindigs []*contracts.ScanFinding) *contracts.ContainerVulnerabilityScanInfo {
	info := &contracts.ContainerVulnerabilityScanInfo{
		Name: container.Name,
		Image: &contracts.Image{
			Name:   container.Image,
			Digest: digest,
		},
		ScanStatus:   scanStatus,
		ScanFindings: scanFindigs,
	}
	return info
}
