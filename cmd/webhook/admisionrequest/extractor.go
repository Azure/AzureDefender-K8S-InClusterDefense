package admisionrequest

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	_errInvalidMetadata = errors.New("extractor.GetWorkloadResourceFromAdmissionRequest: request did not include metadata in the correct format")
	_errInvalidSpec = errors.New("extractor.GetWorkloadResourceFromAdmissionRequest: request did not include spec in the correct format")
)

// UnmarshalPod unmarshals the raw object in the AdmissionRequest into a Pod.
func UnmarshalPod(r *admission.Request) (*corev1.Pod, error) {
	if r == nil {
		return nil, _errInvalidAdmission
	}
	if len(r.Object.Raw) == 0 {
		return nil, _errObjectNotFound
	}
	//if r.Kind.Kind != PodKind {
	//	// If the MutatingWebhookConfiguration was given additional resource scopes.
	//	return nil, _errUnexpectedResource
	//}

	pod := new(corev1.Pod)

	err := json.Unmarshal(r.Object.Raw, &pod)
	if err != nil {
		fmt.Print(err)
		return nil, errors.Wrap(err, "extractor.UnmarshalPod: failed in json.Unmarshal")
	}
	return pod, nil
}

func UnmarshalResource(r *admission.Request) (*unstructured.Unstructured, error) {
	if r == nil {
		return nil, _errInvalidAdmission
	}
	if len(r.Object.Raw) == 0 {
		return nil, _errObjectNotFound
	}
	resource := &unstructured.Unstructured{}
	err := json.Unmarshal(r.Object.Raw, resource)
	if err != nil {
		fmt.Print(err)
		return nil, errors.Wrap(err, "extractor.UnmarshalResource: failed in json.Unmarshal")
	}
	return resource, nil
}

type WorkloadResource struct {
	PodSpec          corev1.PodSpec
	ResourceMetadata metav1.ObjectMeta
	//resourceKind metav1.TypeMeta
}
func GetWorkloadResourceFromAdmissionRequest(req admission.Request) (resource *WorkloadResource, err error){
	pod, err := UnmarshalResource(&req)
	if err != nil {
		return nil,err
	}
	workloadResource := WorkloadResource{}
	byt1,_ := json.Marshal(pod.Object["metadata"])
	_ = json.Unmarshal(byt1, &workloadResource.ResourceMetadata)
	//workloadResource.ResourceMetadata, _ = (pod.Object["metadata"]).(metav1.ObjectMeta)
	//if ok==false{
	//	return nil,_errInvalidMetadata
	//}
	var byt2 []byte
	if req.Kind.Kind == PodKind{
		byt2,_ = json.Marshal(pod.Object["spec"])
		_ = json.Unmarshal(byt2, &workloadResource.PodSpec)
		//workloadResource.PodSpec, _ = pod.Object["spec"].(corev1.PodSpec)
		//if ok==false{
		//	return nil,_errInvalidSpec
		//}
	} else{
		byt2,_ = json.Marshal(pod.Object["spec"].(map[string]interface {})["spec"])
		_ = json.Unmarshal(byt2, &workloadResource.PodSpec)
		//workloadResource.PodSpec, _ = pod.Object["spec"].(map[string]interface {})["spec"].(corev1.PodSpec)
		//if ok==false{
		//	return nil,_errInvalidSpec
		//}
	}
	_ = json.Unmarshal(byt2, &workloadResource.PodSpec)
	return &workloadResource,nil
}

