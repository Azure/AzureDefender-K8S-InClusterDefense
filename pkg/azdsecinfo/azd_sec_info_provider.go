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
}

// NewAzdSecInfoProvider - AzdSecInfoProvider Ctor
func NewAzdSecInfoProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argDataProvider arg.IARGDataProvider) *AzdSecInfoProvider {
	return &AzdSecInfoProvider{
		tracerProvider:  instrumentationProvider.GetTracerProvider("AzdSecInfoProvider"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		argDataProvider: argDataProvider,
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
		tracer.Error(err, "Error from GetRegistryAndRepositoryFromImageReference")
		return nil, err
	}

	tracer.Info("Container image ref extracted context", "context", imageRefContexct)

	digest := "sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108"

	scanStatus := contracts.Unscanned
	scanFindigs := []*contracts.ScanFinding{}

	if strings.HasSuffix(strings.ToLower(imageRefContexct.Registry), _azureContainerRegistrySuffix) {
		v, err := provider.argDataProvider.TryGetImageVulnerabilityScanResults(imageRefContexct.Registry, imageRefContexct.Repository, digest)
		if err != nil {
			tracer.Error(err, "Error from argDataProvider.TryGetImageVulnerabilityScanResults")
			return nil, err
		}

		if len(v) == 0 {
			scanStatus = contracts.HealthyScan
		} else {
			scanStatus = contracts.UnhealthyScan
			scanFindigs = v
		}

	}

	info := &contracts.ContainerVulnerabilityScanInfo{
		Name: container.Name,
		Image: &contracts.Image{
			Name:   container.Image,
			Digest: digest,
		},
		ScanStatus:   scanStatus,
		ScanFindings: scanFindigs,
	}

	return info, nil
}
