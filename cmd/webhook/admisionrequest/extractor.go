package admisionrequest

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// PodKind admission pod kind of the pod request in admission review
	PodKind = "Pod"
)

var (
	_errObjectNotFound     = errors.New("admisionrequest.extractor: request did not include object")
	_errUnexpectedResource = errors.New("admisionrequest.extractor: expected pod resource")
	_errInvalidAdmission   = errors.New("admisionrequest.extractor: admission request was nil")
)

// UnmarshalPod unmarshals the raw object in the AdmissionRequest into a Pod.
func UnmarshalPod(r *admission.Request) (*corev1.Pod, error) {
	if r == nil {
		return nil, _errInvalidAdmission
	}
	if len(r.Object.Raw) == 0 {
		return nil, _errObjectNotFound
	}
	if r.Kind.Kind != PodKind {
		// If the MutatingWebhookConfiguration was given additional resource scopes.
		return nil, _errUnexpectedResource
	}

	pod := new(corev1.Pod)

	err := json.Unmarshal(r.Object.Raw, &pod)
	if err != nil {
		fmt.Print(err)
		return nil, errors.Wrap(err, "extractor.UnmarshalPod: failed in json.Unmarshal")
	}
	return pod, nil
}
