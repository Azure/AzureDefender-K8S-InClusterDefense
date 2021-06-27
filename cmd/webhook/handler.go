// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"github.com/go-logr/logr"
	"gomodules.xyz/jsonpatch/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// The reasons of patching
const (
	reasonPatched = "Patched"
)

// Handler is handle with all admission requests according to the MutatingWebhookConfiguration.
// It implements the Handle interface that each webhook have to implement.
type Handler struct {
	Logger logr.Logger // Logger is the handler logger - gets it from the server.
	DryRun bool        // dryRun if its true, handle request but don't mutate the pod spec.
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	h.Logger.Info("received request",
		"name", req.Name,
		"namespace", req.Namespace,
		"operation", req.Operation,
		"object", req.Object)
	var patches []jsonpatch.JsonPatchOperation
	if h.DryRun { // In case of dryrun=true:  reset all patch operations
		h.Logger.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}
	//Patch all patches operations
	return admission.Patched(reasonPatched, patches...)
}
