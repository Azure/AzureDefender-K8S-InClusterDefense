package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type App struct {
}

func (app *App) HandleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func (app *App) HandleMutate(w http.ResponseWriter, r *http.Request) {
	admissionReview := &admissionv1.AdmissionReview{}

	// read the AdmissionReview from the request json body
	err := readJSON(r, admissionReview)
	if err != nil {
		app.HandleError(w, r, err)
		return
	}

	// unmarshal the pod from the AdmissionRequest
	pod := &corev1.Pod{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, pod); err != nil {
		app.HandleError(w, r, fmt.Errorf("unmarshal to pod: %v", err))
		return
	}

	scanMap := []ImageScanEntry{}

	for i := 0; i < len(pod.Spec.Containers); i++ {
		scanMap = append(scanMap, ImageScanEntry{
			Image: pod.Spec.Containers[i].Image,
			Severity: SeverityEntry{
				High:   2,
				Medium: 5,
				Low:    10},
			Status: "Scanned",
		})
	}

	scanMapSer, err := json.Marshal(&scanMap)
	if err != nil {
		app.HandleError(w, r, fmt.Errorf("marshall jsonpatch: %v", err))
		return
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
		return
	}

	patchType := admissionv1.PatchTypeJSONPatch

	// build admission response
	admissionResponse := &admissionv1.AdmissionResponse{
		UID:       admissionReview.Request.UID,
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}

	respAdmissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: admissionResponse,
	}

	jsonOk(w, &respAdmissionReview)
}

type JSONPatchAnnotationsEntry struct {
	OP    string            `json:"op"`
	Path  string            `json:"path"`
	Value map[string]string `json:"value,omitempty"`
}

type ImageScanEntry struct {
	Image    string        `json:"image"`
	Severity SeverityEntry `json:"severity"`
	Status   string        `json:"status"`
}

type SeverityEntry struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
}
