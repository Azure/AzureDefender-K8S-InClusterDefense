package tag2digest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/pkg/errors"
)

var (
	nilArgError = errors.New("NilArgError")
)

type ITag2DigestResolver interface {
	Resolve(imageReference registry.IImageReference, authContext *ResourceContext) (string, error)
}

type ResourceContext struct {
	Namespace          string
	ImagePullSecrets   []string
	ServiceAccountName string
}

type Tag2DigestResolver struct {
	//tracerProvider is tracer provider of AzdSecInfoProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter is metric submitter of AzdSecInfoProvider
	metricSubmitter metric.IMetricSubmitter
	// registryClient is the client of the registry which is used to resolve image's digest
	registryClient registry.IRegistryClient
}

func NewTag2DigestResolver(instrumentationProvider instrumentation.IInstrumentationProvider, registryClient registry.IRegistryClient) *Tag2DigestResolver {
	return &Tag2DigestResolver{
		tracerProvider:  instrumentationProvider.GetTracerProvider("Tag2DigestResolver"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		registryClient:  registryClient}
}

func (resolver *Tag2DigestResolver) Resolve(imageReference registry.IImageReference, resourceCtx *ResourceContext) (string, error) {
	tracer := resolver.tracerProvider.GetTracer("Resolve")
	tracer.Info("Received:", "imageReference", imageReference, "resourceCtx", resourceCtx)

	if imageReference == nil || resourceCtx == nil {
		err := errors.Wrap(nilArgError, "Tag2DigestResolver.Resolve")
		tracer.Error(err, "")
		return "", err
	}

	// First check if we can extract it from ref it self (digest based ref)
	digestReference, ok := imageReference.(*registry.Digest)
	if ok {
		tracer.Info("ImageReference is digestReference return it is digest", "digestReference", digestReference, "digest", digestReference.Digest())
		return digestReference.Digest(), nil
	}

	// ACR auth
	if registryutils.IsRegistryEndpointACR(imageReference.Registry()) {
		digest, err := resolver.registryClient.GetDigestUsingACRAttachAuth(imageReference)
		if err != nil {
			// todo only on unauthorized and retry only on transient
			tracer.Error(err, "Failed on ACR auth -> continue to other types of auth")
		}else{
			return digest, nil
		}
	}

	// Fallback to K8S auth
	// TODO Add fallback on missing pull secret
	digest, err := resolver.registryClient.GetDigestUsingK8SAuth(imageReference, resourceCtx.Namespace, resourceCtx.ImagePullSecrets, resourceCtx.ServiceAccountName)
	if err != nil {
		err = errors.Wrap(err, "Tag2DigestResolver.Resolve: Failed to get digest on K8sAuth")
		tracer.Error(err, "")
		return "", err

	}
	return  digest, nil
}
