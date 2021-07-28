// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/pkg/infra/instrumentation"
	"gomodules.xyz/jsonpatch/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// patchStatus status of patching
type patchStatus string

// The reasons of patching
const (
	patched patchStatus = "Patched"
)

// Handler is handle with all admission requests according to the MutatingWebhookConfiguration.
// It implements the Handle interface that each webhook have to implement.
type Handler struct {
	Instrumentation *instrumentation.Instrumentation // Instrumentation is the handler instrumentation - gets it from the server.
	dryRun          bool                             // dryRun is flag that if its true, it handles request but doesn't mutate the pod spec.
}

// NewHandler creates new handler
func NewHandler(instrumentation *instrumentation.Instrumentation, dryRun bool) *Handler {
	return &Handler{
		Instrumentation: instrumentation,
		dryRun:          dryRun,
	}
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	// Exit with panic in case that the context is nil
	if ctx == nil {
		panic("Can't handle requests when the context (ctx) is nil")
	}

	h.Instrumentation.Tracer.Info("received request",
		"name", req.Name,
		"namespace", req.Namespace,
		"operation", req.Operation)

	var patches []jsonpatch.JsonPatchOperation
	//TODO invoke AzDSecInfo and patch the result.

	// In case of dryrun=true:  reset all patch operations
	if h.dryRun {
		h.Instrumentation.Tracer.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}
	//Patch all patches operations
	return admission.Patched(string(patched), patches...)
}
