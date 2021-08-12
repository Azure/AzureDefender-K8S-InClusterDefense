// Package webhook is setting up the webhook service and it's own dependencies (e.g. cert controller, logger, metrics, etc.).
package webhook

import (
	"context"
	"log"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/annotations"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// patchReason enum status reason of patching
type patchReason string

const (
	// _patched in case that the handler patched to the webhook.
	_patched    patchReason = "Patched"
	_notPathced patchReason = "NotPatched"
)

// Constants
const (
	_podKind = "Pod"
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
	if logger == nil {
		logger = ctrl.Log.WithName("handler")
	}

	return &Handler{
		// TODO Update on real instrumentation
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

	// TODO: Debug
	handler.Logger.Info("req", "req", req)

	// Logs
	handler.Logger.Info("received request", "name", req.Name, "namespace", req.Namespace, "operation", req.Operation, "reqKind", req.Kind, "uid", req.UID)

	patches := []jsonpatch.JsonPatchOperation{}
	patchReason := _notPathced

	if req.Kind.Kind == _podKind {

		pod, err := admisionrequest.UnmarshalPod(&req)
		if err != nil {
			wrappedError := errors.Wrap(err, "Failed to admisionrequest.UnmarshalPod req")
			handler.Logger.Error(wrappedError, "")
			log.Fatal(err)
		}

		vulnerabilitySecAnnotationsPatch, err := handler.getContainersVulnerabilityScanInfoAnnotationsOperation(pod)
		if err != nil {
			wrappedError := errors.Wrap(err, "Failed to getContainersVulnerabilityScanInfoAnnotationsOperation for Pod")
			handler.Logger.Error(wrappedError, "")
			log.Fatal(err)
		}

		// Add to response patches
		patches = append(patches, *vulnerabilitySecAnnotationsPatch)

		// update patch reason
		patchReason = _patched
	}

	// In case of dryrun=true:  reset all patch operations
	if handler.Configuration.DryRun {
		handler.Logger.Info("not mutating resource, because dry-run=true")
		patches = []jsonpatch.JsonPatchOperation{}
	}

	// Patch all patches operations
	response := admission.Patched(string(patchReason), patches...)
	handler.Logger.Info("Responded", "response", response)
	return response
}

func (handler *Handler) getContainersVulnerabilityScanInfoAnnotationsOperation(pod *corev1.Pod) (*jsonpatch.JsonPatchOperation, error) {
	vulnSecInfoContainers := []*contracts.ContainerVulnerabilityScanInfo{}
	for _, container := range pod.Spec.InitContainers {

		// Get container vulnerability scan information for congainers
		vulnerabilitySecInfo, err := handler.AzdSecInfoProvider.GetContainerVulnerabilityScanInfo(&container)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo Init containers")
			handler.Logger.Error(wrappedError, "")
			return nil, wrappedError
		}

		// Add it to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}

	for _, container := range pod.Spec.Containers {

		// Get container vulnerability scan information for congainers
		vulnerabilitySecInfo, err := handler.AzdSecInfoProvider.GetContainerVulnerabilityScanInfo(&container)
		if err != nil {
			wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo Init containers")
			handler.Logger.Error(wrappedError, "")
			return nil, wrappedError
		}

		// Add it to slice
		vulnSecInfoContainers = append(vulnSecInfoContainers, vulnerabilitySecInfo)
	}

	// Create the annotations add json patch operation
	vulnerabilitySecAnnotationsPatch, err := annotations.CreateContainersVulnerabilityScanAnnotationPatchAdd(vulnSecInfoContainers)
	if err != nil {
		wrappedError := errors.Wrap(err, "Handler failed to GetContainersVulnerabilityScanInfo")
		handler.Logger.Error(wrappedError, "Handler.AzdSecInfoProvider.GetContainersVulnerabilityScanInfo")
		return nil, wrappedError
	}

	return vulnerabilitySecAnnotationsPatch, nil
}
