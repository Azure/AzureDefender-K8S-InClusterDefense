package tag2digest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registryerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"time"
)

// ITag2DigestResolver responsible to resolve resource's image to it's digest
type ITag2DigestResolver interface {
	// Resolve receives an image reference and the resource deployed context and resturns image digest
	Resolve(imageReference registry.IImageReference, authContext *ResourceContext) (string, error)
}

// Tag2DigestResolver implements ITag2DigestResolver interface
var _ ITag2DigestResolver = (*Tag2DigestResolver)(nil)

const (
	// _gotResultsFromCache is a flag representing that the results are from cache
	_gotResultsFromCache = true
	// _didntGetResultsFromCache is a flag representing that the results aren't from cache
	_didntGetResultsFromCache = false
)

// Tag2DigestResolver represents basic implementation of ITag2DigestResolver interface
type Tag2DigestResolver struct {
	//tracerProvider is tracer provider of AzdSecInfoProvider
	tracerProvider trace.ITracerProvider
	//metricSubmitter is metric submitter of AzdSecInfoProvider
	metricSubmitter metric.IMetricSubmitter
	// registryClient is the client of the registry which is used to resolve image's digest
	registryClient registry.IRegistryClient
	// tag2DigestResolverCacheClient is the cache Tag2DigestResolver uses and its time expiration
	tag2DigestResolverCacheClient *Tag2DigestResolverCacheClient
}

// Tag2DigestResolverConfiguration is configuration data for Tag2DigestResolver
type Tag2DigestResolverConfiguration struct {
	// cacheExpirationTime is the expiration time for digests in the cache client
	CacheExpirationTimeForResults string
	// cacheExpirationTime is the expiration time in the cache client for errors occurred during get digest
	CacheExpirationTimeForErrors string
}

// Tag2DigestResolverCacheClient is a cache client for Tag2DigestResolver that contains the needed cache client configurations
type Tag2DigestResolverCacheClient struct {
	// cacheClient is a cache for mapping image full name to its digest
	cacheClient cache.ICacheClient
	// cacheExpirationTime is the expiration time for digests in the cache client
	CacheExpirationTimeForResults time.Duration
	// cacheExpirationTime is the expiration time in the cache client for errors occurred during get digest
	CacheExpirationTimeForErrors time.Duration
}

// NewTag2DigestResolver Ctor
func NewTag2DigestResolver(instrumentationProvider instrumentation.IInstrumentationProvider, registryClient registry.IRegistryClient, tag2DigestResolverCacheClient *Tag2DigestResolverCacheClient) *Tag2DigestResolver {
	return &Tag2DigestResolver{
		tracerProvider:  instrumentationProvider.GetTracerProvider("Tag2DigestResolver"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		registryClient:  registryClient,
		tag2DigestResolverCacheClient: tag2DigestResolverCacheClient,
	}
}

// Resolve receives an image reference and the resource deployed context and returns image digest
// Saves digest in cache. The format is key - image original name, value - digest
func (resolver *Tag2DigestResolver) Resolve(imageReference registry.IImageReference, resourceCtx *ResourceContext) (string, error) {
	tracer := resolver.tracerProvider.GetTracer("Resolve")
	tracer.Info("Received:", "imageReference", imageReference, "resourceCtx", resourceCtx)

	// Argument validation
	if imageReference == nil || resourceCtx == nil {
		err := errors.Wrap(utils.NilArgumentError, "Tag2DigestResolver.Resolve")
		tracer.Error(err, "")
		return "", err
	}

	// Try to get digest from cache
	gotResultsFromCache, digest, err := resolver.getDigestFromCache(imageReference)
	if gotResultsFromCache{
		if err != nil{
			err = errors.Wrap(err, "Get error from cache as value")
			tracer.Error(err, "")
			return "", err
		}
		tracer.Info("got digest from cache")
		return digest, nil
	}

	// Get digest
	digest, err = resolver.getDigest(imageReference, resourceCtx)
	if err != nil {
		resolver.setErrorInCache(imageReference, err)
		err = errors.Wrap(err, "Failed to get digest")
		tracer.Error(err, "")
		return "", err
	}

	// Save digest in cache
	err = resolver.tag2DigestResolverCacheClient.cacheClient.Set(imageReference.Original(), digest, resolver.tag2DigestResolverCacheClient.CacheExpirationTimeForResults)
	if err != nil{
		err = errors.Wrap(err, "Tag2DigestResolver.Resolve: Failed to set digest in cache")
		tracer.Error(err, "")
	}
	tracer.Info("Set digest in cache", "image", imageReference.Original(), "digest", digest)

	return digest, nil
}

// getDigestFromCache try to get digest from cache
// The cache mapping image to digest or to known errors.
// If the image exist in cache - return the value (digest or error) and a flag _gotResultsFromCache
// If the image don't exist in cache or any other unknown error occurred - return "", nil and _didntGetResultsFromCache
func (resolver *Tag2DigestResolver) getDigestFromCache(imageReference registry.IImageReference) (bool, string, error) {
	tracer := resolver.tracerProvider.GetTracer("getDigestFromCache")
	// First check if we can get digest from cache
	// Error as a result of key doesn't exist and error from the cache are treated the same (skip cache)
	digestFromCache, err := resolver.tag2DigestResolverCacheClient.cacheClient.Get(imageReference.Original())
	// If key dont exist in cache
	if err != nil {
		tracer.Info("Digest don't exist in cache", "image", imageReference.Original())
		return _didntGetResultsFromCache, "", nil
	}
	// If key exist
	err, isKnownError := registryerrors.TryParseStringToUnscannedWithReasonErr(digestFromCache)
	// A known error was found in cache as value
	if isKnownError{
		err = errors.Wrap(err, "got an error value instead of digest from cache")
		tracer.Error(err, "")
		return _gotResultsFromCache, "", err
	}
	// A valid digest found in cache
	tracer.Info("Digest exist in cache", "image", imageReference.Original(), "digest", digestFromCache)
	return _gotResultsFromCache, digestFromCache, nil
}

// getDigest receives an image refernce and the resource deployed context and returns image digest
// It tries to see if it's a digest based image reference - if so, extract its digest
// Then if it ACR based registry - tries to get digest using registry client's ACR attach auth method
// If above fails or not applicable - tries to get digest using registry client's k8s auth method.
func (resolver *Tag2DigestResolver) getDigest(imageReference registry.IImageReference, resourceCtx *ResourceContext) (string, error){
	tracer := resolver.tracerProvider.GetTracer("getDigest")
	tracer.Info("Received:", "imageReference", imageReference, "resourceCtx", resourceCtx)

	// First check if we can extract it from ref it self (digest based ref)
	digestReference, ok := imageReference.(*registry.Digest)
	if ok {
		tracer.Info("ImageReference is digestReference return it is digest", "digestReference", digestReference, "digest", digestReference.Digest())
		return digestReference.Digest(), nil
	}

	// ACR auth
	if registryutils.IsRegistryEndpointACR(imageReference.Registry()) {
		tracer.Info("ACR suffix so tries ACR auth", "imageRef", imageReference)

		digest, err := resolver.registryClient.GetDigestUsingACRAttachAuth(imageReference)
		if err != nil {
			// TODO Add tests that checks that we don't try another auth when we should stop.
			// 		Should be added once @tomerweinberger finished to merge his PR.
			if !resolver.shouldContinueOnError(err) {
				err = errors.Wrap(err, "Failed to get digest on ACRAttachAuth")
				tracer.Error(err, "")
				return "", err
			}

			// Failed to get digest using ACR attach auth method - continue and fall back to other methods
			tracer.Error(err, "Failed on ACR auth -> continue to other types of auth")
		} else {
			//TODO Check if digest is not empty
			return digest, nil
		}
	}

	tracer.Info("Tries K8S chain auth", "imageRef", imageReference)

	// Fallback to K8S auth
	// TODO Add fallback on missing pull secret
	digest, err := resolver.registryClient.GetDigestUsingK8SAuth(imageReference, resourceCtx.namespace, resourceCtx.imagePullSecrets, resourceCtx.serviceAccountName)
	if err != nil {
		if !resolver.shouldContinueOnError(err) {
			err = errors.Wrap(err, "Failed to get digest on K8SAuth")
			tracer.Error(err, "")
			return "", err
		}

		// Failed to get digest using K8S chain auth method - continue and fall back to other methods
		tracer.Error(err, "Failed on K8S Chain auth -> continue to other types of auth")
	} else {
		//TODO Check if digest is not empty
		return digest, nil
	}

	tracer.Info("Tries DefaultAuth", "imageRef", imageReference)

	// Fallback to DefaultAuth
	digest, err = resolver.registryClient.GetDigestUsingDefaultAuth(imageReference)
	if err != nil {
		err = errors.Wrap(err, "Failed to get digest on DefaultAuth")
		tracer.Error(err, "")
		return "", err
	}

	// Check if the digest is empty
	if digest == "" {
		err = errors.New("Empty digest received by registry client")
		tracer.Error(err, "")
		return "", err
	}

	return digest, nil
}

// ResourceContext represents deployed resource context to use for image digest extraction
type ResourceContext struct {
	namespace          string
	imagePullSecrets   []string
	serviceAccountName string
}

func NewResourceContext(namespace string, imagePullSecrets []string, serviceAccountName string) *ResourceContext {
	return &ResourceContext{
		namespace:          namespace,
		imagePullSecrets:   imagePullSecrets,
		serviceAccountName: serviceAccountName,
	}
}

// shouldContinueOnError is method that gets an error and returns true in case that the error is known
// error that thr resolve method should stop and don't try more authentications to resolve the digest.
func (resolver *Tag2DigestResolver) shouldContinueOnError(err error) bool {
	errorCause := errors.Cause(err)
	switch errorCause.(type) {
	case *registryerrors.ImageIsNotFoundErr,
		*registryerrors.RegistryIsNotFoundErr:
		return false
	default:
		return true
	}
}

// setErrorInCache map image to a known error. If the given error is a known error, set it in cache.
func (resolver *Tag2DigestResolver) setErrorInCache(imageReference registry.IImageReference, err error){
	tracer := resolver.tracerProvider.GetTracer("setErrorInCache")
	errorAsString, isErrParsedToUnscannedReason := registryerrors.TryParseErrToUnscannedWithReason(err)
	if !isErrParsedToUnscannedReason {
		err = errors.Wrap(err, "Unexpected error while trying to get results from ARGDataProvider - don't save in cache")
		tracer.Error(err, "")
		return
	}
	err = resolver.tag2DigestResolverCacheClient.cacheClient.Set(imageReference.Original(), string(*errorAsString), resolver.tag2DigestResolverCacheClient.CacheExpirationTimeForErrors)
	if err != nil{
		err = errors.Wrap(err, "Failed to set error in cache")
		tracer.Error(err, "")
	}
	tracer.Info("Set error in cache", "image", imageReference.Original(), "error", errorAsString)
}
