// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/annotations"
	webhookmetric "github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// responseReason enum status reason of admission response
type responseReason string

const (
	// _patchedReason in case that the handler patched to the webhook.
	_patchedReason responseReason = "Patched"
	//_patchedDeleteStaleContainersVulnerabilityScanInfoAnnotationsReason in case that error occurred and old annotations exist.
	_patchedDeleteStaleContainersVulnerabilityScanInfoAnnotationsReason = "PatchedDeleteStaleContainersVulnerabilityScanInfoAnnotations"
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
	// Extractor extracts workload resource from admission request.
	extractor admisionrequest.IExtractor
}

// HandlerConfiguration configuration for handler
type HandlerConfiguration struct {
	// DryRun is flag that if it's true, it handles request but doesn't mutate the workLoadResource podSpec.
	DryRun                               bool
	SupportedKubernetesWorkloadResources []string
}

// NewHandler Constructor for Handler
func NewHandler(azdSecInfoProvider azdsecinfo.IAzdSecInfoProvider, configuration *HandlerConfiguration, instrumentationProvider instrumentation.IInstrumentationProvider, extractor admisionrequest.IExtractor) *Handler {

	return &Handler{
		tracerProvider:     instrumentationProvider.GetTracerProvider("Handler"),
		metricSubmitter:    instrumentationProvider.GetMetricSubmitter(),
		azdSecInfoProvider: azdSecInfoProvider,
		configuration:      configuration,
		extractor:          extractor,
	}
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (handler *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	startTime := time.Now().UTC()
	tracer := handler.tracerProvider.GetTracer("Handle")
	response := admission.Response{}
	reason := _notPatchedReason
	workLoadResourceName := ""
	var workLoadResourceOwnerRefrences []*admisionrequest.OwnerReference

	var err error
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = errors.New(fmt.Sprint(r))
			}
			tracer.Error(err, "Handler handle Panic error", "resource:", req.Resource, "namespace:", req.Namespace, "Name:", req.Name, "workLoadResourceOwnerRefrences:", workLoadResourceOwnerRefrences, "workLoadResourceName:", workLoadResourceName, "operation:", req.Operation, "reqKind:", req.Kind)
			handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handler.Handle.Panic"))
			// Re throw panic
			panic(r)
		}
		// Repost response latency
		tracer.Info("HandleLatency", "resource", req.Resource, "namespace:", req.Namespace, "Name:", req.Name, "workLoadResourceOwnerRefrences:", workLoadResourceOwnerRefrences, "workLoadResourceName:", workLoadResourceName, "latencyinMS", util.GetDurationMilliseconds(startTime))

		// Extract response status
		var responseCode int32
		var responseResultReasonStr string
		var patchCount int = len(response.Patches)
		if response.Result != nil {
			responseCode = response.Result.Code
			responseResultReasonStr = string(response.Result.Reason)
		}
		tracer.Info("Handle.Response.Result", "resource", req.Resource, "namespace:", req.Namespace, "Name:", req.Name, "Allowed", response.Allowed, "ResultReason", responseResultReasonStr, "code", responseCode, "patchCount", patchCount)
		handler.metricSubmitter.SendMetric(util.GetDurationMilliseconds(startTime), webhookmetric.NewHandlerHandleLatencyMetric(req.Kind.Kind, response.Allowed, responseResultReasonStr, responseCode, patchCount))
	}()
	// Logs
	tracer.Info("received request", "resource:", req.Resource, "namespace:", req.Namespace, "Name:", req.Name, "operation:", req.Operation, "reqKind:", req.Kind)

	handler.metricSubmitter.SendMetric(1, webhookmetric.NewHandlerNewRequestMetric(req.Kind.Kind, req.Operation))

	// Check if the request should be filtered.
	shouldBeFiltered, reason := handler.shouldRequestBeFiltered(req)
	if shouldBeFiltered {
		response = admission.Allowed(string(reason))
		return response
	}
	workloadResource, err := handler.extractor.ExtractWorkloadResourceFromAdmissionRequest(&req)
	if err != nil {
		err = errors.Wrap(err, "Handler.Handle received error on handleRequest")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handle.handleRequest"))
		reason = _notPatchedErrorReason
		response = handler.admissionErrorResponse(errors.Wrap(err, string(reason)))
		return response
	}
	tracer.Info("WorkLoadResource request unmarshall", "resource:", req.Resource, "namespace:", req.Namespace, "WorkLoadResourceOwnerRefrences:", workLoadResourceOwnerRefrences, "operation:", req.Operation, "reqKind:", req.Kind)
	workLoadResourceName = workloadResource.Metadata.Name
	workLoadResourceOwnerRefrences = workloadResource.Metadata.OwnerReferences
	response, err = handler.handleWorkLoadResourceRequest(workloadResource)
	if err != nil {
		err = errors.Wrap(err, "Handler.Handle received error on handleWorkLoadResourceRequest")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handle.handleWorkLoadResourceRequest"))
		response := handler.getResponseWhenErrorEncountered(workloadResource, err)
		tracer.Info("Handler Responded", "resource:", req.Resource, "namespace:", req.Namespace, "Name:", req.Name, "operation:", req.Operation, "reqKind:", req.Kind, "response:", response)
		return response
	}

	// In case of dryrun=true:  reset all patch operations
	if handler.configuration.DryRun {
		tracer.Info("Handler.handleWorkLoadResourceRequest not mutating resource, because handler is on dryrun mode.", "ResponseInCaseOfNotDryRun", response)
		reason = _notPatchedHandlerDryRunReason
		// Override response with clean response.
		response = admission.Allowed(string(reason))
		return response
	}

	reason = _patchedReason
	tracer.Info("Handler Responded", "resource:", req.Resource, "namespace:", req.Namespace, "Name:", req.Name, "operation:", req.Operation, "reqKind:", req.Kind, "response:", response)
	return response
}

// handleWorkLoadResourceRequest gets request that should be handled and returned the response with the relevant patches.
func (handler *Handler) handleWorkLoadResourceRequest(workloadResource *admisionrequest.WorkloadResource) (admission.Response, error) {
	tracer := handler.tracerProvider.GetTracer("handleWorkloadResourceRequest")
	patches := []jsonpatch.JsonPatchOperation{}
	vulnerabilitySecAnnotationsPatch, err := handler.getWorkLoadResourceContainersVulnerabilityScanInfoAnnotationsOperation(workloadResource)
	if err != nil {
		err = errors.Wrap(err, "Handler.handleWorkLoadResourceRequest Failed to getWorkLoadResourceContainersVulnerabilityScanInfoAnnotationsOperation for WorkLoadResource")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "handleWorkLoadResourceRequest.getWorkLoadResourceContainersVulnerabilityScanInfoAnnotationsOperation"))
		return admission.Response{}, err
	}

	// Add to response patches
	patches = append(patches, *vulnerabilitySecAnnotationsPatch)

	// Patch all patches operations
	return admission.Patched(string(_patchedReason), patches...), nil
}

// getResponseWhenErrorEncountered returns a response in which it deletes previous ContainersVulnerabilityScan annotations.
// If no such annotations exist it returns handler.admissionErrorResponse with the original error.
func (handler *Handler) getResponseWhenErrorEncountered(workloadResource *admisionrequest.WorkloadResource, originalError error) admission.Response {
	tracer := handler.tracerProvider.GetTracer("getResponseWhenErrorEncountered")

	patches := []jsonpatch.JsonPatchOperation{}

	// returns nil if no deletion is needed.
	patch, err := annotations.CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotationIfNeeded(workloadResource)

	// if error encountered during CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotationIfNeeded - response with the original error
	if err != nil {
		err = errors.Wrap(err, "Handler.getResponseWhenErrorEncountered Failed to CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotationIfNeeded for workLoadResource")
		tracer.Error(err, "")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "Handler.getResponseWhenErrorEncountered"))
		reason := _notPatchedErrorReason
		response := handler.admissionErrorResponse(errors.Wrap(originalError, string(reason)))
		return response
	}

	// patch is nil when the workLoadResource's annotations doesn't contain the webhook annotations so there is no need
	//to delete them. response with the original error
	if patch == nil {
		tracer.Info("ContainersVulnerabilityScanAnnotation dont exist - no need to delete them ")
		reason := _notPatchedErrorReason
		response := handler.admissionErrorResponse(errors.Wrap(originalError, string(reason)))
		return response
	}

	// Add to response patches
	patches = append(patches, *patch)

	// Patch all patches operations
	return handler.admissionErrorResponseWithAnnotationsDelete(originalError, patches)
}

// getWorkLoadResourceContainersVulnerabilityScanInfoAnnotationsOperation receives a workLoadResource to generate a vuln scan annotation add operation
// Get vuln scan infor from azdSecInfo provider, then create a json annotation for it on workLoadResources custom annotations of azd vuln scan info
func (handler *Handler) getWorkLoadResourceContainersVulnerabilityScanInfoAnnotationsOperation(workloadResource *admisionrequest.WorkloadResource) (*jsonpatch.JsonPatchOperation, error) {
	tracer := handler.tracerProvider.GetTracer("getWorkLoadResourceContainersVulnerabilityScanInfoAnnotationsOperation")
	handler.metricSubmitter.SendMetric(len(workloadResource.Spec.Containers)+len(workloadResource.Spec.InitContainers), webhookmetric.NewHandlerNumOfContainersPerworkLoadResourceMetric())

	// Get workLoadResource's containers vulnerability scan info
	vulnSecInfoContainers, err := handler.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo")
		tracer.Error(wrappedError, "Handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		handler.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(wrappedError, "Handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo"))
		return nil, wrappedError
	}

	// Log result
	tracer.Info("vulnSecInfoContainers", "vulnSecInfoContainers", vulnSecInfoContainers)

	// Create the annotations add json patch operation
	vulnerabilitySecAnnotationsPatch, err := annotations.CreateContainersVulnerabilityScanAnnotationPatchAdd(vulnSecInfoContainers, workloadResource)
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

// admissionErrorResponseWithAnnotationsDelete generates an admission response error in case of handler failing to process request and delete ContainersVulnerabilityScanInfo annotation.
func (handler *Handler) admissionErrorResponseWithAnnotationsDelete(err error, patches []jsonpatch.JsonPatchOperation) admission.Response {
	tracer := handler.tracerProvider.GetTracer("admissionErrorResponseWithAnnotationsDelete")
	tracer.Error(err, "")
	// If status code is not 200-OK (500 for example), but there is still non nil patch, the mutation applied successfully on api-server as status admission review is not verified on mutating webhooks.
	response := handler.admissionErrorResponse(errors.Wrap(err, string(_patchedDeleteStaleContainersVulnerabilityScanInfoAnnotationsReason)))
	response.Patches = patches
	return response
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

	// Filter if the kind is not workload resource
	if !utils.StringInSlice(req.Kind.Kind, handler.configuration.SupportedKubernetesWorkloadResources) {
		tracer.Info("Request filtered out due to the request is not supported kind.", "ReqKind", req.Kind.Kind)
		return true, _noMutationForKindReason
	}

	// Filter if the operation is not Create or update
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
