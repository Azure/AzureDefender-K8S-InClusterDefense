// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"log"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/annotations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// patchReason enum status reason of patching
type patchReason string

const (
	// _patched in case that the handler patched to the webhook.
	_patched    patchReason = "Patched"
	_notPathced patchReason = "NotPatched"
)

// Handler implements the admission.Handle interface that each webhook have to implement.
// Handler handles with all admission requests according to the MutatingWebhookConfiguration.
type Handler struct {
	// tracerProvider of the handler
	tracerProvider trace.ITracerProvider
	// MetricSubmitter
	metricSubmitter metric.IMetricSubmitter
	// AzdSecInfoProvider provides azure defender security information
	azdSecInfoProvider azdsecinfo.IAzdSecInfoProvider
	// Configurations handler's config.
	configuration *HandlerConfiguration
}

// HandlerConfiguration configuration for handler
type HandlerConfiguration struct {
	// DryRun is flag that if it's true, it handles request but doesn't mutate the pod spec.
	DryRun bool
}

// NewHandler Constructor for Handler
func NewHandler(azdSecInfoProvider azdsecinfo.IAzdSecInfoProvider, configuration *HandlerConfiguration, instrumentationProvider instrumentation.IInstrumentationProvider) (handler *Handler) {

	return &Handler{
		tracerProvider:     instrumentationProvider.GetTracerProvider("Handler"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		azdSecInfoProvider: azdSecInfoProvider,
		configuration:      configuration,
	}
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (handler *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	tracer := handler.tracerProvider.GetTracer("Handle")
	if ctx == nil {
		tracer.Error(errors.New("ctx received is nil"), "Handler.Handle")
		// Exit with panic in case that the context is nil
		log.Fatal("Can't handle requests when the context (ctx) is nil")
	}

	// Logs
	tracer.Info("received request", "name", req.Name, "namespace", req.Namespace, "operation", req.Operation, "reqKind", req.Kind, "uid", req.UID)

	patches := []jsonpatch.JsonPatchOperation{}
	patchReason := _notPathced

	if req.Kind.Kind == admisionrequest.PodKind {

		pod, err := admisionrequest.UnmarshalPod(&req)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handle handler failed to admisionrequest.UnmarshalPod req")
			tracer.Error(wrappedError, "")
			log.Fatal(wrappedError)
		}

		vulnerabilitySecAnnotationsPatch, err := handler.getPodContainersVulnerabilityScanInfoAnnotationsOperation(pod)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler.Handle Failed to getPodContainersVulnerabilityScanInfoAnnotationsOperation for Pod")
			tracer.Error(wrappedError, "")
			log.Fatal(wrappedError)
		}

		// Add to response patches
		patches = append(patches, *vulnerabilitySecAnnotationsPatch)

		// update patch reason
		patchReason = _patched
	}

	// In case of dryrun=true:  reset all patch operations
	if handler.configuration.DryRun {
		tracer.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}

	// Patch all patches operations
	response := admission.Patched(string(patchReason), patches...)
	tracer.Info("Responded", "response", response)
	return response
}

func (handler *Handler) getPodContainersVulnerabilityScanInfoAnnotationsOperation(pod *corev1.Pod) (*jsonpatch.JsonPatchOperation, error) {
	tracer := handler.tracerProvider.GetTracer("getPodContainersVulnerabilityScanInfoAnnotationsOperation")
	vulnSecInfoContainers := []*contracts.ContainerVulnerabilityScanInfo{}

	for _, container := range pod.Spec.InitContainers {

		// Get container vulnerability scan information for init containers
		vulnerabilitySecInfo, err := handler.azdSecInfoProvider.GetContainerVulnerabilityScanInfo(&container)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo Init containers")
			tracer.Error(wrappedError, "")
			return nil, wrappedError
		}

		// Add it to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}

	for _, container := range pod.Spec.Containers {

		// Get container vulnerability scan information for containers
		vulnerabilitySecInfo, err := handler.azdSecInfoProvider.GetContainerVulnerabilityScanInfo(&container)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo Containers")
			tracer.Error(wrappedError, "")
			return nil, wrappedError
		}

		// Add it to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}
	// Create the annotations add json patch operation
	vulnerabilitySecAnnotationsPatch, err := annotations.CreateContainersVulnerabilityScanAnnotationPatchAdd(vulnSecInfoContainers)
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo")
		tracer.Error(wrappedError, "Handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		return nil, wrappedError
	}

	return vulnerabilitySecAnnotationsPatch, nil
}
