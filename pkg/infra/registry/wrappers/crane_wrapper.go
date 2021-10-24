package wrappers

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	craneerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers/errors"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

var (
	_emptyDigestErr = errors.New("crane returned empty digest")
)

// ICraneWrapper wraps crane operations
type ICraneWrapper interface {
	// Digest get image digest using image ref using crane Digest call
	Digest(ref string, opt ...crane.Option) (string, error)
	// DigestWithRetry calls Digest to get image's digest, with retries
	DigestWithRetry(ref string, opt ...crane.Option) (res string, err error)
}

// CraneWrapper wraps crane operations
type CraneWrapper struct {
	// tracerProvider is the tracer provider for the crane wrapper.
	tracerProvider trace.ITracerProvider
	// metricSubmitter is the metric submitter for the crane wrapper.
	metricSubmitter metric.IMetricSubmitter
	// retryPolicy is the manager of the retry policy of the crane wrapper.
	retryPolicy retrypolicy.IRetryPolicy
}

// NewCraneWrapper Cto'r for CraneWrapper
func NewCraneWrapper(instrumentationProvider instrumentation.IInstrumentationProvider, retryPolicy retrypolicy.IRetryPolicy) *CraneWrapper {
	return &CraneWrapper{
		tracerProvider:  instrumentationProvider.GetTracerProvider("CraneWrapper"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		retryPolicy:     retryPolicy,
	}
}

// Digest get image digest using image ref using crane Digest call
// Todo add auth options to pull secrets and ACR MSI based - currently only supports docker config auth
// K8s chain pull secrets ref: https://github.com/google/go-containerregistry/blob/main/pkg/authn/k8schain/k8schain.go
// ACR ref: // https://github.com/Azure/acr-docker-credential-helper/blob/master/src/docker-credential-acr/acr_login.go
func (craneWrapper *CraneWrapper) Digest(ref string, opt ...crane.Option) (string, error) {
	//(resolved digest of tomerwdevops.azurecr.io/imagescan:62 - https://ms.portal.azure.com/#@microsoft.onmicrosoft.com/resource/subscriptions/4009f3ee-43c4-4f19-97e4-32b6f2285a68/resourceGroups/tomerwdevops/providers/Microsoft.ContainerRegistry/registries/tomerwdevops/repository)
	tracer := craneWrapper.tracerProvider.GetTracer("Digest")
	digest, err := crane.Digest(ref, opt...)
	if err != nil {
		tracer.Error(err, "error encountered while trying to get digest with crane.")
		return "", craneerrors.ConvertErrToKnownErr(ref, err)
	} else if digest == "" {
		err = _emptyDigestErr
		tracer.Error(err, "")
		return "", err
	}
	tracer.Info("Crane Resolved digest", "image reference", ref, "options", opt, "digest", digest)
	return digest, nil
}

// DigestWithRetry re-executing Digest in case of a failure according to retryPolicy
func (craneWrapper *CraneWrapper) DigestWithRetry(imageReference string, opt ...crane.Option) (string, error) {
	tracer := craneWrapper.tracerProvider.GetTracer("DigestWithRetry")

	digest, err := craneWrapper.retryPolicy.RetryActionString(
		/*action ActionString*/
		func() (string, error) { return craneWrapper.Digest(imageReference, opt...) },

		/*handle ShouldRetryOnSpecificError*/
		func(err error) bool {
			if utils.IsErrorIsTypeOf(err, craneerrors.GetImageIsNotFoundErrType()) ||
				utils.IsErrorIsTypeOf(err, craneerrors.GetUnauthorizedErrType()) {
				return false
			}
			return true
		},
	)

	if err != nil {
		err = errors.Wrapf(err, "failed to extract digest of image %v", imageReference)
		tracer.Error(err, "")
		return "", err
	}

	tracer.Info("Managed to extract digest", "Image ref", imageReference, "digest", digest)
	return digest, nil
}
