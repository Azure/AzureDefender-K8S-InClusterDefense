package tag2digest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registryauth "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/pkg/errors"
)

type ITag2DigestResolver interface {
	Resolve(imageReference string, authContext *registryauth.AuthContext) (string, error)
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

func (resolver *Tag2DigestResolver) Resolve(imageReference string, authContext *registryauth.AuthContext) (string, error) {
	tracer := resolver.tracerProvider.GetTracer("Resolve")
	tracer.Info("Received:", "imageReference", imageReference, "authContext", authContext)

	// First check if we can extract it from ref it self (digest based ref)
	isDigestBasedImageRef, digest, err := registryutils.TryExtractDigestFromImageRef(imageReference)
	if err != nil {
		// Report error
		err = errors.Wrap(err, "utils.TryExtractDigestFromImageRef:")
		tracer.Error(err, "")
		return "", err
	} else if isDigestBasedImageRef {
		tracer.Info("utils.TryExtractDigestFromImageRef return that it is digest based", "isDigestBasedImageRef", isDigestBasedImageRef, "digest", digest)
		return digest, nil
	} else {

		if registryutils.IsRegistryEndpointACR(authContext.RegistryEndpoint) {
			digest, err = resolver.registryClient.GetDigest(imageReference,
				&registryauth.AuthConfig{
					AuthType: registryauth.ACRAuth,
					Context:  authContext})
			if err != nil {
				tracer.Error(err, "Failed on ACR auth -> continue to other types of auth")
			}
		}

		digest, err = resolver.registryClient.GetDigest(imageReference,
			&registryauth.AuthConfig{
				AuthType: registryauth.K8SAuth,
				Context:  authContext,
			})

		if err != nil {
			err = errors.Wrap(err, "Tag2DigestResolver.Resolve: Failed to get digest on K8sAuth")
			tracer.Error(err, "")
			return "", err

		}
		return digest, nil
	}
}
