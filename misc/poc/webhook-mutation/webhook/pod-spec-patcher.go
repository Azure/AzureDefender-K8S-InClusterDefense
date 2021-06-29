package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

// Json path annotations entry
type JSONPatchAnnotationsEntry struct {
	OP    string            `json:"op"`
	Path  string            `json:"path"`
	Value map[string]string `json:"value,omitempty"`
}

// Path scan result to annotations of pod spec.
func (app *App) pathScanResultToPodSpec(
	w http.ResponseWriter,
	r *http.Request,
	scanMap []ImageSecInfo,
	pod *corev1.Pod,
	admissionReview *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	scanMapSer, err := json.Marshal(&scanMap)
	if err != nil {
		app.HandleError(w, r, fmt.Errorf("marshall jsonpatch: %v", err))
		return nil
	}
	podAnnotations := pod.Annotations
	if podAnnotations == nil {
		podAnnotations = make(map[string]string)
	}
	podAnnotations["azure-denfder.io/scanInfo"] = string(scanMapSer)
	// build json patch
	patch := []JSONPatchAnnotationsEntry{
		{
			OP:    "add",
			Path:  "/metadata/annotations",
			Value: podAnnotations,
		},
	}
	patchBytes, err := json.Marshal(&patch)
	if err != nil {
		app.HandleError(w, r, fmt.Errorf("marshall jsonpatch: %v", err))
		return nil
	}
	patchType := admissionv1.PatchTypeJSONPatch
	// build admission response
	admissionResponse := &admissionv1.AdmissionResponse{
		UID:       admissionReview.Request.UID,
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
	return admissionResponse
}
