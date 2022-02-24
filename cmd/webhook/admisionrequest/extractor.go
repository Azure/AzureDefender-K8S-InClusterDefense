package admisionrequest

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	yaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	// PodKind admission pod kind of the pod request in admission review
	PodKind = "Pod"
	CronJobKind = "CronJob"
	DeploymentKind = "Deployment"
	ReplicasetKind = "Replicaset"
	StatefulSetKind = "StatefulSet"
	ReplicationControllerKind = "ReplicationController"
	JobKind = "Job"
	DaemonSetKind = "DaemonSet"
	Spec = "spec"
	Metadata = "metadata"
	Template = "template"
	JobTemplate = "jobTemplate"

)

var (
	_errObjectNotFound     = errors.New("admisionrequest.extractor: request did not include object")
	_errUnexpectedResource = errors.New("admisionrequest.extractor: expected workload resource.")
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

type WorkloadResource struct{
	ResourceMetadata metav1.ObjectMeta `json:"metadata"`
	PodSpec          corev1.PodSpec `json:"spec"`
}

type OuterSpec  struct {
	Template corev1.PodTemplateSpec `json:"template"`
}

type TemporaryWorkloadResource struct {
	ResourceMetadata metav1.ObjectMeta `json:"metadata"`
	OuterSpec OuterSpec `json:"spec"`
}

type TemporaryCronJob struct {
	ResourceMetadata metav1.ObjectMeta `json:"metadata"`
	CronJobSpec struct{
		JobTemplate  struct {
			OuterSpec OuterSpec `json:"spec"`
		} `json:"jobTemplate"`
	}`json:"spec"`
}

type MetadataRes struct{
	Namespace string
	Annotation map[string]string
}

type Container struct{
	Name string
	Image string
}

type SpecRes struct{
	Containers []Container
	InitContainers []Container
	ImagePullSecrets []corev1.LocalObjectReference
	ServiceAccountName string
}


type ResourceWorkLoad struct{
	Metadata MetadataRes
	Spec SpecRes
}


func GetContainers(yamlFile *yaml.RNode,ContainersType string) (containers []Container,err error) {
	conParent, err := yamlFile.Pipe(yaml.Lookup("spec", ContainersType))
	if err!=nil {
		return nil,errors.Wrap(err, "fail")
	}
	if conParent==nil{
		if ContainersType=="containers"{
			return nil, errors.New("fail")
		}else{
			return nil,nil
		}
	}
	if conParent.Content()==nil || len(conParent.Content())==0{
		return nil, errors.Wrap(err, "fail")
	}
	containersList := make([]Container, len(conParent.Content()))
	for i := 0; i < len(conParent.Content()); i += 1 {
		containersList[i] = Container{}
		inner := conParent.Content()[i].Content
		if len(inner)!=4{
			return nil,errors.Wrap(err, "fail")
		}
		if inner[0].Value == "name" && inner[2].Value == "image" {
			containersList[i].Name = inner[1].Value
			containersList[i].Image = inner[3].Value
		} else if inner[2].Value == "name" && inner[0].Value == "image" {
			containersList[i].Name = inner[3].Value
			containersList[i].Image = inner[1].Value
		} else {
			return nil,errors.Wrap(err, "fail")
		}
	}
	return containersList,nil
}

func GetAnnotation(yamlFile *yaml.RNode) (annotation map[string]string, err error){
	AnnotationParent , err := yamlFile.Pipe(yaml.Lookup("metadata","annotations"))
	if err!=nil{
		return nil,errors.Wrap(err, "fail")
	}
	if AnnotationParent==nil{
		return nil,nil
	}
	if len(AnnotationParent.Content())==0 || len(AnnotationParent.Content())/2!=0{
		return nil,errors.New("failed")
	}
	annotationMap := make(map[string] string)
	for i := 0; i < len(AnnotationParent.Content()); i +=2 {
		annotationMap[AnnotationParent.Content()[i].Value] = AnnotationParent.Content()[i+1].Value
	}
	return annotationMap,nil
}
func GetValue(yamlFile *yaml.RNode,path ...string) (value string,err error){
	valueParent , err := yamlFile.Pipe(yaml.Lookup(path ...))
	if err!=nil{
		return "",errors.Wrap(err, "fail")
	}
	val := yaml.GetValue(valueParent)
	return val,nil
}

func GetWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *ResourceWorkLoad, err error){
	if req == nil {
		return nil, _errInvalidAdmission
	}
	if len(req.Object.Raw) == 0 {
		return nil, _errObjectNotFound
	}
	yamlFile, _ := yaml.ConvertJSONToYamlNode(string(req.Object.Raw))
	workResource := ResourceWorkLoad{}
	//imagePullSecrets , _ := yamlFile.Pipe(yaml.Lookup("spec","imagePullSecrets"))
	//fmt.Print(imagePullSecrets)
	annotations  := yamlFile.GetAnnotations()
	if len(annotations)==0{
		workResource.Metadata.Annotation = nil
	}
	workResource.Metadata.Annotation = a
	if err!=nil{
		return nil,errors.Wrap(err, "fail")
	}
	workResource.Spec.Containers, err =GetContainers(yamlFile,"containers")
	if err!=nil{
		return nil,errors.Wrap(err, "fail")
	}
	workResource.Spec.InitContainers, err =GetContainers(yamlFile,"initContainers")
	if err!=nil{
		return nil,errors.Wrap(err, "fail")
	}
	workResource.Spec.ServiceAccountName , err = GetValue(yamlFile,"spec","serviceAccountName")
	if err!=nil{
		return nil,errors.Wrap(err, "fail")
	}
	workResource.Metadata.Namespace,err = GetValue(yamlFile,"metadata","namespace")
	if err!=nil{
		return nil,errors.Wrap(err, "fail")
	}
	return &workResource,nil
}


//func GetWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *ResourceWorkLoad, err error){
//	if req == nil {
//		return nil, _errInvalidAdmission
//	}
//	if len(req.Object.Raw) == 0 {
//		return nil, _errObjectNotFound
//	}
//	obj := req.Object.Raw
//	workResource := ResourceWorkLoad{}
//	if req.Kind.Kind == PodKind{
//		 err = json.Unmarshal(obj, &workResource)
//		 if err != nil{
//			 return nil, errors.Wrap(err, "failed")
//		 }
//	}else if req.Kind.Kind == CronJobKind {
//		resource := TemporaryCronJob{}
//		err = json.Unmarshal(obj, &resource)
//		if err != nil{
//			return nil, errors.Wrap(err, "failed")
//		}
//		workResource.ResourceMetadata = resource.ResourceMetadata
//		workResource.PodSpec = resource.CronJobSpec.JobTemplate.OuterSpec.Template.Spec
//	}else if req.Kind.Kind==DeploymentKind || req.Kind.Kind==ReplicasetKind || req.Kind.Kind==StatefulSetKind ||
//		req.Kind.Kind==ReplicationControllerKind || req.Kind.Kind==JobKind || req.Kind.Kind==DaemonSetKind {
//		resource := TemporaryWorkloadResource{}
//		err = json.Unmarshal(obj, &resource)
//		if err != nil{
//			return nil, errors.Wrap(err, "failed")
//		}
//		workResource.ResourceMetadata = resource.ResourceMetadata
//		workResource.PodSpec = resource.OuterSpec.Template.Spec
//	}else{
//		return nil, _errUnexpectedResource
//	}
//	return &workResource,nil
//}


//func GetWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error) {
//	if req == nil {
//		return nil, _errInvalidAdmission
//	}
//	if len(req.Object.Raw) == 0 {
//		return nil, _errObjectNotFound
//	}
//	unstructuredObbject := &unstructured.Unstructured{}
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



