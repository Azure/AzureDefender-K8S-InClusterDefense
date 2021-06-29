package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"github.com/prometheus/common/log"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Application manager
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

	// Unmarshal the pod from the AdmissionRequest
	pod := &corev1.Pod{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, pod); err != nil {
		app.HandleError(w, r, fmt.Errorf("unmarshal to pod: %v", err))
		return
	}

	scanMap := []ImageSecInfo{} // TODO Move this scan map into the for loop.
	// Add
	for i := 0; i < len(pod.Spec.Containers); i++ {
		imageAsString := pod.Spec.Containers[i].Image
		// Get scan result:
		imageSecInfo, err := GetImageSecInfo(imageAsString)
		if err != nil {
			log.Errorf("Error: %s", err)
			return
		}
		scanMap = append(scanMap, *imageSecInfo)
		admissionResponse := app.pathScanResultToPodSpec(w, r, scanMap, pod, admissionReview)
		respAdmissionReview := &admissionv1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
			Response: admissionResponse,
		}
		jsonOk(w, &respAdmissionReview)
	}
}
