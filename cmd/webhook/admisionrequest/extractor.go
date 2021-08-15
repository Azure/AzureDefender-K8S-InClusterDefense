package admisionrequest

import (
	"encoding/json"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// _podKind kind of the pod request in admission review
	_podKind = "Pod"
)

var (
	_errObjectNotFound     = errors.New("admisionrequest.extractor: request did not include object")
	_errUnexpectedResource = errors.New("admisionrequest.extractor: expected pod resource")
	_errInvalidAdmission   = errors.New("admisionrequest.extractor: admission request was nil")
)

// UnmarshalPod unmarshals the raw object in the AdmissionRequest into a Pod.
func UnmarshalPod(r *admission.Request) (pod *corev1.Pod, err error) {
	if r == nil {
		return nil, _errInvalidAdmission
	}
	if len(r.Object.Raw) == 0 {
		return nil, _errObjectNotFound
	}
	if r.Kind.Kind != _podKind {
		// If the MutatingWebhookConfiguration was given additional resource scopes.
		return nil, _errUnexpectedResource
	}

	if err := json.Unmarshal(r.Object.Raw, pod); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal failed in unmarshaling pod")
	}
	return pod, nil
}
