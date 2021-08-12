package admisionrequest

import (
	"encoding/json"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	_podKind = "Pod"
)

var (
	errObjectNotFound     = errors.New("admisionrequest.extractor: request did not include object")
	errUnexpectedResource = errors.New("admisionrequest.extractor: expected pod resource")
	errInvalidAdmission   = errors.New("admisionrequest.extractor: admission request was nil")
)

// UnmarshalPod unmarshals the raw object in the AdmissionRequest into a Pod.
func UnmarshalPod(r *admission.Request) (*corev1.Pod, error) {
	if r == nil {
		return nil, errInvalidAdmission
	}
	if len(r.Object.Raw) == 0 {
		return nil, errObjectNotFound
	}
	if r.Kind.Kind != _podKind {
		// If the ValidatingWebhookConfiguration was given additional resource scopes.
		return nil, errUnexpectedResource
	}

	var pod corev1.Pod
	if err := json.Unmarshal(r.Object.Raw, &pod); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal failed in unmarshaling pod")
	}
	return &pod, nil
}
