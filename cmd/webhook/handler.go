// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/annotations"
	webhookmetric "github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// responseReason enum status reason of admission response
type responseReason string

const (
	// _patchedReason in case that the handler patched to the webhook.
	_patchedReason responseReason = "Patched"
	// _notPatchedReason not patched response reason.
	_notPatchedReason responseReason = "NotPatched"
	// _notPatchedErrorReason not patched due to error response reason.
	_notPatchedErrorReason responseReason = "NotPatchedError"
	// _notPatchedDryRunReason in case that DryRun of Handler is True.
	_notPatchedHandlerDryRunReason responseReason = "NotPatchedHandlerDryRun"
	// _noMutationForOperationOrKindReason in case that the resource kind of the request is not supported kind
	_noMutationForKindReason responseReason = "NotPatchedNotSupportedKind"
	// _noMutationForOperationOrKindReason in case that the resource kind of the request is not supported kind
	_noMutationForOperationReason responseReason = "NotPatchedNotSupportedOperation"
	// _noSelfManagementReason in case of resource in same namespace
	_noSelfManagementReason responseReason = "NotPatchedResourceInTheSameNsOfHandler"
)

// Handler implements admission.Handler interface
var _ admission.Handler = (*Handler)(nil)

// Handler implements the admission.Handler interface that each webhook have to implement.
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
func NewHandler(azdSecInfoProvider azdsecinfo.IAzdSecInfoProvider, configuration *HandlerConfiguration, instrumentationProvider instrumentation.IInstrumentationProvider) *Handler {

	return &Handler{
		tracerProvider:     instrumentationProvider.GetTracerProvider("Handler"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		azdSecInfoProvider: azdSecInfoProvider,
		configuration:      configuration,
	}
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (handler *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	startTime := time.Now().UTC()
	tracer := handler.tracerProvider.GetTracer("Handle")
	response := admission.Response{}
	reason := _notPatchedReason
	podName  := ""
	var podOwnerRefrences[] metav1.OwnerReference

	var err error
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = errors.New(fmt.Sprint(r))
			}
			tracer.Error(err, "Handler handle Panic error","resource:", req.Resource,"namespace:", req.Namespace,"Name:", req.Name,"podOwnerRefrences:",podOwnerRefrences, "PodName:", podName, "operation:", req.Operation, "reqKind:", req.Kind)
			handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handler.Handle.Panic"))
			// Re throw panic
			panic(r)
		}
		// Repost response latency
		tracer.Info("HandleLatency", "resource", req.Resource, "namespace:", req.Namespace,"Name:", req.Name,"podOwnerRefrences:",podOwnerRefrences, "PodName:", podName, "latencyinMS", util.GetDurationMilliseconds(startTime))

		// Extract response status
		var responseCode int32
		var responseResultReasonStr string
		var patchCount int = len(response.Patches)
		if response.Result != nil {
			responseCode = response.Result.Code
			responseResultReasonStr = string(response.Result.Reason)
		}
		tracer.Info("Handle.Response.Result", "resource", req.Resource, "namespace:", req.Namespace,"Name:", req.Name, "Allowed", response.Allowed, "ResultReason", responseResultReasonStr, "code", responseCode, "patchCount", patchCount)
		handler.metricSubmitter.SendMetric(util.GetDurationMilliseconds(startTime), webhookmetric.NewHandlerHandleLatencyMetric(req.Kind.Kind, response.Allowed, responseResultReasonStr, responseCode, patchCount))
	}()

	// Logs
	tracer.Info("received request", "resource:", req.Resource,"namespace:", req.Namespace,"Name:", req.Name, "operation:", req.Operation, "reqKind:", req.Kind)

	handler.metricSubmitter.SendMetric(1, webhookmetric.NewHandlerNewRequestMetric(req.Kind.Kind, req.Operation))

	// Check if the request should be filtered.
	shouldBeFiltered, reason := handler.shouldRequestBeFiltered(req)
	if shouldBeFiltered {
		response = admission.Allowed(string(reason))
		return response
	}
	pod, err := admisionrequest.UnmarshalPod(&req)
	if err != nil {
		err = errors.Wrap(err, "Handler.Handle received error on handlePodRequest")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handle.handlePodRequest"))
		reason = _notPatchedErrorReason
		response = handler.admissionErrorResponse(errors.Wrap(err, string(reason)))
		return response
	}
	tracer.Info("Pod request unmarshall","resource:", req.Resource,"namespace:", req.Namespace,"podName:", pod.Name, "podOwnerRefrences:", podOwnerRefrences, "podOwners:", pod.OwnerReferences, "operation:", req.Operation, "reqKind:", req.Kind)
	podName = pod.Name
	podOwnerRefrences = pod.OwnerReferences

	response, err = handler.handlePodRequest(pod)
	if err != nil {
		err = errors.Wrap(err, "Handler.Handle received error on handlePodRequest")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handle.handlePodRequest"))
		response := handler.getResponseWhenErrorEncountered(pod, err)
		return response
	}

	// In case of dryrun=true:  reset all patch operations
	if handler.configuration.DryRun {
		tracer.Info("Handler.handlePodRequest not mutating resource, because handler is on dryrun mode.", "ResponseInCaseOfNotDryRun", response)
		reason = _notPatchedHandlerDryRunReason
		// Override response with clean response.
		response = admission.Allowed(string(reason))
		return response
	}

	reason = _patchedReason
	tracer.Info("Handler Responded","resource:", req.Resource,"namespace:", req.Namespace,"Name:", req.Name, "operation:", req.Operation, "reqKind:", req.Kind, "response:", response)
	return response
}

// handlePodRequest gets request that should be handled and returned the response with the relevant patches.
func (handler *Handler) handlePodRequest(pod *corev1.Pod) (admission.Response, error) {
	tracer := handler.tracerProvider.GetTracer("handlePodRequest")

	patches := []jsonpatch.JsonPatchOperation{}

	vulnerabilitySecAnnotationsPatch, err := handler.getPodContainersVulnerabilityScanInfoAnnotationsOperation(pod)
	if err != nil {
		err = errors.Wrap(err, "Handler.handlePodRequest Failed to getPodContainersVulnerabilityScanInfoAnnotationsOperation for Pod")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "handlePodRequest.getPodContainersVulnerabilityScanInfoAnnotationsOperation"))
		return admission.Response{}, err
	}

	// Add to response patches
	patches = append(patches, *vulnerabilitySecAnnotationsPatch)

	// Patch all patches operations
	return admission.Patched(string(_patchedReason), patches...), nil
}
// getResponseWhenErrorEncountered returns a response in which it deletes previous ContainersVulnerabilityScan annotations.
// If no such annotations exist it returns handler.admissionErrorResponse with the original error.
func (handler *Handler) getResponseWhenErrorEncountered(pod *corev1.Pod, originalError error) admission.Response {
	tracer := handler.tracerProvider.GetTracer("getResponseWhenErrorEncountered")

	patches := []jsonpatch.JsonPatchOperation{}

	patch, err := annotations.CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotation(pod)

	// if error encountered during CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotation - response with the original error
	if err != nil {
		err = errors.Wrap(err, "Handler.getResponseWhenErrorEncountered Failed to CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotation for Pod")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handler.getResponseWhenErrorEncountered"))
		reason := _notPatchedErrorReason
		response := handler.admissionErrorResponse(errors.Wrap(originalError, string(reason)))
		return response
	}

	// patch is nil when the pod's annotations doesn't contain the webhook annotations so there is no need to delete them.
	// response with the original error
	if patch == nil{
		tracer.Info("ContainersVulnerabilityScanAnnotation dont exist - no need to delete them ")
		reason := _notPatchedErrorReason
		response := handler.admissionErrorResponse(errors.Wrap(originalError, string(reason)))
		return response
	}

	// Add to response patches
	patches = append(patches, *patch)

	// Patch all patches operations
	return admission.Patched(string(_patchedReason), patches...)
}

// getPodContainersVulnerabilityScanInfoAnnotationsOperation receives a pod to generate a vuln scan annotation add operation
// Get vuln scan infor from azdSecInfo provider, then create a json annotation for it on pods custom annotations of azd vuln scan info
func (handler *Handler) getPodContainersVulnerabilityScanInfoAnnotationsOperation(pod *corev1.Pod) (*jsonpatch.JsonPatchOperation, error) {
	tracer := handler.tracerProvider.GetTracer("getPodContainersVulnerabilityScanInfoAnnotationsOperation")
	handler.metricSubmitter.SendMetric(len(pod.Spec.Containers)+len(pod.Spec.InitContainers), webhookmetric.NewHandlerNumOfContainersPerPodMetric())

	// Get pod's containers vulnerability scan info
	vulnSecInfoContainers, err := handler.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo")
		tracer.Error(wrappedError, "Handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(wrappedError, "Handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo"))
		return nil, wrappedError
	}

	// Log result
	tracer.Info("vulnSecInfoContainers", "vulnSecInfoContainers", vulnSecInfoContainers)

	// Create the annotations add json patch operation
	vulnerabilitySecAnnotationsPatch, err := annotations.CreateContainersVulnerabilityScanAnnotationPatchAdd(vulnSecInfoContainers, pod)
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to CreateContainersVulnerabilityScanAnnotationPatchAdd")
		tracer.Error(wrappedError, "Handler.annotations.CreateContainersVulnerabilityScanAnnotationPatchAdd")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(wrappedError, "Handler.annotations.CreateContainersVulnerabilityScanAnnotationPatchAdd"))
		return nil, wrappedError
	}

	return vulnerabilitySecAnnotationsPatch, nil
}

// admissionErrorResponse generates an admission response error in case of handler failing to process request
func (handler *Handler) admissionErrorResponse(err error) admission.Response {
	tracer := handler.tracerProvider.GetTracer("admissionErrorResponse")
	tracer.Error(err, "")
	return admission.Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			// mutating webhook error should not block deployment
			Allowed: true,
			Result: &metav1.Status{
				Code:    int32(http.StatusInternalServerError),
				Message: err.Error(),
			},
		},
	}
}

// shouldRequestBeFiltered checks if the request should be filtered.
// In case that it should be filtered, it returns true and the admission.Response.
// In case that it shouldn't be filtered, it returns false and nil
func (handler *Handler) shouldRequestBeFiltered(req admission.Request) (bool, responseReason) {
	// TODO add filter on blacklist namespaces.
	tracer := handler.tracerProvider.GetTracer("shouldRequestBeFiltered")
	// If it's the same namespace of the mutation webhook
	if req.Namespace == utils.GetDeploymentInstance().GetNamespace() {
		tracer.Info("Request filtered out due to it is in the same namespace as the handler.", "Namespace", req.Namespace)
		return true, _noSelfManagementReason
	}

	// Filter if the kind is not pod.
	if req.Kind.Kind != admisionrequest.PodKind {
		tracer.Info("Request filtered out due to the request is not supported kind.", "ReqKind", req.Kind.Kind)
		return true, _noMutationForKindReason
	}

	// Filter if the operation is not Create
	if !isOperationAllowed(&req.Operation) {
		tracer.Info("Request filtered out due to the request is not supported operation.", "ReqOperation", req.Operation)
		return true, _noMutationForOperationReason
	}

	tracer.Info("Request shouldn't be filtered out.")
	// Request shouldn't be filtered out.
	return false, _patchedReason
}

// isOperationAllowed returns boolean if the operation is allowed.
func isOperationAllowed(operation *admissionv1.Operation) bool {
	return *operation == admissionv1.Create || *operation == admissionv1.Update
}
