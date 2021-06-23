package webhook

import (
	"context"
	"github.com/go-logr/logr"
	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	reasonPatched = "Patched"
)

type Handler struct {
	Log    logr.Logger
	DryRun bool
	Config *rest.Config
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	h.Log.Info("received request",
		"name", req.Name,
		"namespace", req.Namespace,
		"operation", req.Operation,
		"object", req.Object)
	var patches []jsonpatch.JsonPatchOperation
	if h.DryRun { // In case of dryrun=true:  reset all patch operations
		h.Log.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}
	//Patch all patches operations
	return admission.Patched(reasonPatched, patches...)
}
