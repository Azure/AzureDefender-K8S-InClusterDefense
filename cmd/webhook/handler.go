// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
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
	// Instrumentation is the handler instrumentation - gets it from the server.
	Instrumentation *instrumentation.Instrumentation
}

// NewHandler Constructor for Handler
func NewHandler(runOnDryRun bool, instrumentation *instrumentation.Instrumentation) (handler *Handler) {
	return &Handler{
		Instrumentation: instrumentation,
		DryRun:          runOnDryRun,
	}
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if ctx == nil {
		// Exit with panic in case that the context is nil
		panic("Can't handle requests when the context (ctx) is nil")
	}

	h.Instrumentation.Tracer.Info("received request",
		"name", req.Name,
		"namespace", req.Namespace,
		"operation", req.Operation,
		"request", req)

	var patches []jsonpatch.JsonPatchOperation
	//TODO invoke AzDSecInfo and patch the result.

	// In case of dryrun=true:  reset all patch operations
	if h.DryRun {
		h.Instrumentation.Tracer.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}
	//Patch all patches operations
	return admission.Patched(string(_patched), patches...)
}
