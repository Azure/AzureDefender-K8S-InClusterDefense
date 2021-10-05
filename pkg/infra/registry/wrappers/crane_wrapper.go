package wrappers

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/pkg/errors"
)

// ICraneWrapper wraps crane operations
type ICraneWrapper interface {
	// Digest get image digest using image ref using crane Digest call
	Digest(ref string, opt ...crane.Option) (string, error)
	// DigestWithRetry calls Digest to get image's digest, with retries
	DigestWithRetry(ref string, tracerProvider trace.ITracerProvider, metricSubmitter metric.IMetricSubmitter, opt ...crane.Option) (res string, err error)
}

// CraneWrapper wraps crane operations
type CraneWrapper struct {
	// retryPolicy is the manager of the retry policy of the crane wrapper.
	retryPolicy retrypolicy.IRetryPolicy
}

// NewCraneWrapper Cto'r for CraneWrapper
func NewCraneWrapper(retryPolicy retrypolicy.IRetryPolicy) *CraneWrapper {
	return &CraneWrapper{
		retryPolicy: retryPolicy,
	}
}

// Digest get image digest using image ref using crane Digest call
// Todo add auth options to pull secrets and ACR MSI based - currently only supports docker config auth
// K8s chain pull secrets ref: https://github.com/google/go-containerregistry/blob/main/pkg/authn/k8schain/k8schain.go
// ACR ref: // https://github.com/Azure/acr-docker-credential-helper/blob/master/src/docker-credential-acr/acr_login.go
func (*CraneWrapper) Digest(ref string, opt ...crane.Option) (string, error) {
	//(resolved digest of tomerwdevops.azurecr.io/imagescan:62 - https://ms.portal.azure.com/#@microsoft.onmicrosoft.com/resource/subscriptions/4009f3ee-43c4-4f19-97e4-32b6f2285a68/resourceGroups/tomerwdevops/providers/Microsoft.ContainerRegistry/registries/tomerwdevops/repository)
	digest, err := crane.Digest(ref, opt...)
	// Crane is not checking cases that digest is empty //TODO why does it happen?
	if digest == "" {
		err = errors.Wrapf(err, "failed to extract digest of image ref <%s>. crane returned empty digest", ref)
		return "", err
	}

	return digest, err
}

// DigestWithRetry re-executing Digest in case of a failure according to retryPolicy
func (craneWrapper *CraneWrapper) DigestWithRetry(imageReference string, tracerProvider trace.ITracerProvider, metricSubmitter metric.IMetricSubmitter, opt ...crane.Option) (string, error) {
	tracer := tracerProvider.GetTracer("DigestWithRetry")

	var action retrypolicy.ActionString = func() (string, error) {
		return craneWrapper.Digest(imageReference, opt...)
	}

	var handle retrypolicy.Handle = func(error) bool {
		// TODO deal with cases for which we do not want to retry after method as been implemented
		return false
	}

	digest, err := craneWrapper.retryPolicy.RetryActionString(action, handle)

	if err != nil {
		err := errors.Wrapf(err, "failed to extract digest of image %v", imageReference)
		tracer.Error(err, "")
		return "", err
	}

	tracer.Info("Managed to extract digest", "Image ref", imageReference, "digest", digest)
	return digest, nil
}
