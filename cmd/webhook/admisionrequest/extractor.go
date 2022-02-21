package admisionrequest

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// PodKind admission pod kind of the pod request in admission review
	PodKind = "Pod"
	CronJobKind = "CronJob"
	Spec = "spec"
	Metadata = "metadata"
	Template = "template"
	JobTemplate = "jobTemplate"

)

var (
	_errObjectNotFound     = errors.New("admisionrequest.extractor: request did not include object")
	_errUnexpectedResource = errors.New("admisionrequest.extractor: expected pod resource")
	_errInvalidAdmission   = errors.New("admisionrequest.extractor: admission request was nil")
	_errUnmarshal = "extractor.GetWorkloadResourceFromAdmissionRequest: failed in json.Unmarshal"
	_errMarshal = "extractor.GetWorkloadResourceFromAdmissionRequest: failed in json.Unmarshal"
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

type PodNew struct{
	ResourceMetadata metav1.ObjectMeta `json:"metadata"`
	PodSpec          corev1.PodSpec `json:"spec"`
}
type OuterSpec  struct {
	Template  struct {
		PodSpec corev1.PodSpec `json:"spec"`
	} `json:"template"`
}

type GeneralWorkloadResource struct {
	ResourceMetadata metav1.ObjectMeta `json:"metadata"`
	OuterSpec OuterSpec `json:"spec"`
}

type CronJob struct {
	ResourceMetadata metav1.ObjectMeta `json:"metadata"`
	CronJobSpec struct{
		JobTemplate  struct {
			OuterSpec OuterSpec `json:"spec"`
		} `json:"jobTemplate"`
	}`json:"spec"`
}


type WorkloadResource struct {
	PodSpec          corev1.PodSpec
	ResourceMetadata metav1.ObjectMeta

}

func GetWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error){
	if req == nil {
		return nil, _errInvalidAdmission
	}
	if len(req.Object.Raw) == 0 {
		return nil, _errObjectNotFound
	}
	obj := req.Object.Raw
	workResource := WorkloadResource{}
	if req.Kind.Kind == PodKind{
		resource := PodNew{}
		 err = json.Unmarshal(obj, &resource)
		 if err != nil{
			 return nil, errors.Wrap(err, "failed")
		 }
		 workResource.ResourceMetadata = resource.ResourceMetadata
		 workResource.PodSpec = resource.PodSpec
	}else if req.Kind.Kind == CronJobKind {
		resource := CronJob{}
		err = json.Unmarshal(obj, &resource)
		if err != nil{
			return nil, errors.Wrap(err, "failed")
		}
		workResource.ResourceMetadata = resource.ResourceMetadata
		workResource.PodSpec = resource.CronJobSpec.JobTemplate.OuterSpec.Template.PodSpec
	}else{
		resource := GeneralWorkloadResource{}
		err = json.Unmarshal(obj, &resource)
		if err != nil{
			return nil, errors.Wrap(err, "failed")
		}
		workResource.ResourceMetadata = resource.ResourceMetadata
		workResource.PodSpec = resource.OuterSpec.Template.PodSpec
	}
	return &workResource,nil
}


//func GetWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error) {
//	if req == nil {
//		return nil, _errInvalidAdmission
//	}
//	if len(req.Object.Raw) == 0 {
//		return nil, _errObjectNotFound
//	}
//	unstructuredObject := &unstructured.Unstructured{}
//	er := json.Unmarshal(req.Object.Raw, unstructuredObject)
//	if er != nil {
//		return nil, errors.Wrap(er, _errUnmarshal)
//	}
//	metadataByt,er1 := json.Marshal(unstructuredObject.Object[Metadata])
//	if er1 != nil{
//		return nil, errors.Wrap(er1, _errMarshal)
//	}
//	outerSpecByt,er2 := json.Marshal(unstructuredObject.Object[Spec])
//	if er2 != nil{
//		return nil, errors.Wrap(er2, _errMarshal)
//	}
//	workloadResource := WorkloadResource{}
//	errUnmarshalMetadata := json.Unmarshal(metadataByt, &workloadResource.ResourceMetadata)
//	if errUnmarshalMetadata != nil{
//		return nil, errors.Wrap(errUnmarshalMetadata, _errUnmarshal)
//	}
//	var errInnerSpecMarshal error
//	var innerSpecByt []byte
//	if req.Kind.Kind == PodKind{
//		innerSpecByt = outerSpecByt
//	}else if req.Kind.Kind == CronJobKind {
//		var outerSpec = map[string]map[string]map[string]map[string]interface{}{}
//		_ = json.Unmarshal(outerSpecByt, &outerSpec)
//		innerSpecByt, errInnerSpecMarshal = json.Marshal(outerSpec[JobTemplate][Spec][Template][Spec])
//	}else{
//		var outerSpec = map[string]map[string]interface{}{}
//		_ = json.Unmarshal(outerSpecByt, &outerSpec)
//		innerSpecByt, errInnerSpecMarshal = json.Marshal(outerSpec[Template][Spec])
//	}
//	if errInnerSpecMarshal!=nil{
//		return nil, errors.Wrap(errInnerSpecMarshal, _errMarshal)
//	}
//	errInnerSpecUnmarshal := json.Unmarshal(innerSpecByt, &workloadResource.PodSpec)
//	if errInnerSpecUnmarshal!=nil{
//		return nil, errors.Wrap(errInnerSpecUnmarshal, _errUnmarshal)
//	}
//	return &workloadResource,nil
//}



