package wrappers

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	registrymetric "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/metric"
	"github.com/pkg/errors"
	"time"
	"github.com/google/go-containerregistry/pkg/crane"
)

// ICraneWrapper wraps crane operations
type ICraneWrapper interface {
	// Digest get image digest using image ref using crane Digest call
	Digest(ref string) (string, error)
	// DigestWithRetry calls Digest to get image's digest, with retries
	DigestWithRetry(ref string, tracerProvider trace.ITracerProvider, metricSubmitter metric.IMetricSubmitter) (res string, err error)
}

// CraneWrapper wraps crane operations
type CraneWrapper struct {
	// Number of attempts to retrieve digest from arg
	RetryAttempts int
	// time duration between each retry
	RetryDuration time.Duration
}

/*
// SetCraneWrapperInstrumentation sets the crane wrapper object's tracer and metric
func (craneWrapper *CraneWrapper) SetCraneWrapperInstrumentation(instrumentationProvider instrumentation.IInstrumentationProvider) {
	craneWrapper.tracerProvider = instrumentationProvider.GetTracerProvider("SetCraneWrapperInstrumentation")
	craneWrapper.metricSubmitter = instrumentationProvider.GetMetricSubmitter()
}
*/

// Digest get image digest using image ref using crane Digest call
// Todo add auth options to pull secrets and ACR MSI based - currently only supports docker config auth
// K8s chain pull secrets ref: https://github.com/google/go-containerregistry/blob/main/pkg/authn/k8schain/k8schain.go
// ACR ref: // https://github.com/Azure/acr-docker-credential-helper/blob/master/src/docker-credential-acr/acr_login.go
func (*CraneWrapper) Digest(ref string, opt ...crane.Option) (string, error) {
	//TODO implement this method using the libraries that were mentioned above. in the meantime, return static digest
	//(resolved digest of tomerwdevops.azurecr.io/imagescan:62 - https://ms.portal.azure.com/#@microsoft.onmicrosoft.com/resource/subscriptions/4009f3ee-43c4-4f19-97e4-32b6f2285a68/resourceGroups/tomerwdevops/providers/Microsoft.ContainerRegistry/registries/tomerwdevops/repository)
	//return "sha256:f6a835256950f699175eecb9fd82e4a84684c9bab6ffb641b6fc23ff7b23e4b3", nil
	return crane.Digest(ref, opt...)
}

// DigestWithRetry re-executing Digest in case of a failure according to retryPolicy
func (craneWrapper *CraneWrapper) DigestWithRetry(ref string, tracerProvider trace.ITracerProvider, metricSubmitter metric.IMetricSubmitter) (res string, err error) {
	tracer := tracerProvider.GetTracer("GetDigestWithRetries")
	retryCount := 1
	for retryCount < craneWrapper.RetryAttempts + 1{
		// TODO change Digest call after Digest method has been implemented correctly
		// TODO deal with cases for which we do not want to retry after method as been implemented
		// Execute Digest and check if an error occurred. We want to retry if err is not nil
		if res, err = craneWrapper.Digest(ref); err != nil {
			tracer.Info("Managed to extract digest", "attempts:", retryCount)
			return res, nil
		} else {
			tracer.Error(err, "failed extracting digest from ARC", "attempts:", retryCount)
			retryCount += 1
			// wait (retryCount * craneWrapper.retryDuration) milliseconds between retries
			time.Sleep(time.Duration(retryCount) * craneWrapper.RetryDuration * time.Millisecond)
		}
	}
	// Send metrics
	metricSubmitter.SendMetric(retryCount, registrymetric.NewCraneWrapperNumOfRetryAttempts())
	return res, errors.Wrapf(err, "failed after %d retries due to error", retryCount)
}
