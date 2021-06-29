package webhook

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	reasonNoMutationForOperation = "NoMutationForOperation"
	reasonNoSelfManagement       = "NoSelfManagement"
	reasonNotPatched             = "NotPatched"
	reasonPatched                = "Patched"
)

type Handler struct {
	Log    logr.Logger
	DryRun bool
	Config *rest.Config
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	h.Log.Info("received request", "name", req.Name, "namespace", req.Namespace, "gvk", req.Kind)
	if req.Operation != admissionv1.Create && req.Operation != admissionv1.Update {
		return admission.Allowed(reasonNoMutationForOperation)
	}
	r, err := yaml.Parse(string(req.Object.Raw))
	if err != nil {
		return h.admissionError(err)
	}
	h.Log.Info("parsed resource", "resource", r)
	// if req.Namespace == util.GetNamespace() {
	// 	return admission.Allowed(reasonNoSelfManagement)
	// }
	// The raw resource may not contain the namespace, but the admission review
	// request will. That's why the next step copies the namespace from the
	// request to the resource. This handles situations such as when the
	// ReplicaSetController creates a new pod.
	if err := r.SetNamespace(req.Namespace); err != nil {
		h.admissionError(err)
	}
	before, err := r.MarshalJSON()
	if err != nil {
		return h.admissionError(err)
	}

	// if err = resolveImageTags(ctx, h.Log, h.Config, r); err != nil {
	// 	return h.admissionError(err)
	// }

	after, err := r.MarshalJSON()
	if err != nil {
		return h.admissionError(err)
	}
	patches, err := jsonpatch.CreatePatch(before, after)
	if err != nil {
		return h.admissionError(err)
	}
	h.Log.Info("patched resource", "patches", patches)
	if h.DryRun {
		h.Log.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}
	reason := reasonPatched
	if len(patches) == 0 {
		reason = reasonNotPatched
	}
	h.Log.Info("Patched Successfully")
	return admission.Patched(reason, patches...)
}

func (h *Handler) admissionError(err error) admission.Response {
	h.Log.Error(err, "admission error")
	return admission.Errored(int32(http.StatusInternalServerError), err)
}
