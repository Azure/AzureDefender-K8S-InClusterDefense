// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"encoding/json"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
	// Logger is the handler logger - gets it from the server.
	Logger logr.Logger
	// AzdSecInfoProvider provides azure defender security information
	AzdSecInfoProvider azdsecinfo.IAzdSecInfoProvider
	// Configurations handler's config.
	Configuration *HandlerConfiguration
}

type HandlerConfiguration struct {
	// DryRun is flag that if it's true, it handles request but doesn't mutate the pod spec.
	DryRun bool
}

// NewHandler Constructor for Handler
func NewHandler(azdSecInfoProvider azdsecinfo.IAzdSecInfoProvider, configuration *HandlerConfiguration, logger logr.Logger) (handler *Handler) {
	return &Handler{
		Logger:             logger,
		AzdSecInfoProvider: azdSecInfoProvider,
		Configuration:      configuration,
	}
}

// Handle processes the AdmissionRequest by invoking the underlying function.
func (handler *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if ctx == nil {
		// Exit with panic in case that the context is nil
		panic("Can't handle requests when the context (ctx) is nil")
	}

	handler.Logger.Info("received request",
		"name", req.Name,
		"namespace", req.Namespace,
		"operation", req.Operation,
		"object", req.Object)

	vulnerabilitySecInfo, err := handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo()
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo")
		handler.Logger.Error(wrappedError, "Handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		panic(err)
	}

	marshaledVulnerabilitySecInfo, err := json.Marshal(vulnerabilitySecInfo)
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to marshaling GetContainersVulnerabilityScanInfo result")
		handler.Logger.Error(wrappedError, "Handler Marshal marshaling GetContainersVulnerabilityScanInfo result")
		panic(err)
	}

	patches := []jsonpatch.JsonPatchOperation{
		jsonpatch.NewOperation("add", "/metadata/annotations/kubernetes.io~1ingress.class", marshaledVulnerabilitySecInfo),
	}

	// In case of dryrun=true:  reset all patch operations
	if handler.Configuration.DryRun {
		handler.Logger.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}
	//Patch all patches operations
	return admission.Patched(string(_patched), patches...)
}
