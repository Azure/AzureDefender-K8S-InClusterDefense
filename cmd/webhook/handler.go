// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"gomodules.xyz/jsonpatch/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// patchStatus status of patching
type patchStatus string

// The reasons of patching
const (
	// _patched in case that the handler patched to the webhook.
	_patched patchStatus = "Patched"
)

// Handler implements the admission.Handle interface that each webhook have to implement.
// Handler handles with all admission requests according to the MutatingWebhookConfiguration.
type Handler struct {
	// DryRun is flag that if it's true, it handles request but doesn't mutate the pod spec.
	DryRun bool
	// TracerProvider of the handler
	TracerProvider trace.ITracerProvider
	// MetricSubmitter
	MetricSubmitter metric.IMetricSubmitter
}

// NewHandler Constructor for Handler
func NewHandler(runOnDryRun bool, provider instrumentation.IInstrumentationProvider) (handler *Handler) {
	tracerProvider := provider.GetTracerProvider("Handler")
	metricSubmitter := provider.GetMetricSubmitter()
	return &Handler{
		TracerProvider:  tracerProvider,
		MetricSubmitter: metricSubmitter,
		DryRun:          runOnDryRun,
	}
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (handler *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	tracer := handler.TracerProvider.GetTracer("Handle")
	if ctx == nil {
		// Exit with panic in case that the context is nil
		panic("Can't handle requests when the context (ctx) is nil")
	}

	tracer.Info("received request",
		"name", req.Name,
		"namespace", req.Namespace,
		"operation", req.Operation,
		"request", req)

	var patches []jsonpatch.JsonPatchOperation
	//TODO invoke AzDSecInfo and patch the result.

	// In case of dryrun=true:  reset all patch operations
	if handler.DryRun {
		tracer.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}
	//Patch all patches operations
	return admission.Patched(string(_patched), patches...)
}
