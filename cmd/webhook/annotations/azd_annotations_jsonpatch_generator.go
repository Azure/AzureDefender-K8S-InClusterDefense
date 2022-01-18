package annotations

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	corev1 "k8s.io/api/core/v1"
	"time"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
)

const (
	// _addPatchOperation operation type in json patch to add annotations
	_addPatchOperation = "add"
	// _annotationPatchPath operation path in json patch to add object annotations
	_annotationPatchPath = "/metadata/annotations"
)

// CreateContainersVulnerabilityScanAnnotationPatchAdd returns an add type json patch in order to add to annotations map a new key value of ContainersVulnerabilityScanInfoAnnotationName.
// It does so by adding to the exiting map the new key value and setting the updated map as the json patch value.
// The function creates a scanInfoList from the provided containers scan info  slice of type contracts.ContainerVulnerabilityScanInfoList serialize/marshal it and set it as a value string to the new key annotation
// Contracts.ContainersVulnerabilityScanInfoAnnotationName (azuredefender.io/containers.vulnerability.scan.info)
// If the annotations map doesn't exist, it creates a new map and add the key value before setting it as the json patch value.
// As a result, the annotations are updated with no override of the existing values.
func CreateContainersVulnerabilityScanAnnotationPatchAdd(containersScanInfoList []*contracts.ContainerVulnerabilityScanInfo, pod *corev1.Pod) (*jsonpatch.JsonPatchOperation, error) {
	scanInfoList := &contracts.ContainerVulnerabilityScanInfoList{
		GeneratedTimestamp: time.Now().UTC(),
		Containers:         containersScanInfoList,
	}

	// Marshal the scan info list (annotations can only be strings)
	serVulnerabilitySecInfo, err := marshalAnnotationInnerObject(scanInfoList)
	if err != nil {
		return nil, errors.Wrap(err, "AzdAnnotationsPatchGenerator failed marshaling scanInfoList during CreateContainersVulnerabilityScanAnnotationPatchAdd")
	}

	// Create annotations map and add to the map serVulnerabilitySecInfo. If the pod's annotations is nil create a new map
	annotations, err := updateAnnotations(pod, contracts.ContainersVulnerabilityScanInfoAnnotationName, serVulnerabilitySecInfo)
	if err != nil {
		return nil, errors.Wrap(err, "AzdAnnotationsPatchGenerator failed updating annotations because pod is nil during CreateContainersVulnerabilityScanAnnotationPatchAdd")
	}

	// Create an add operation to annotations to add or create if annotations are empty
	// **important note** any future changes to the pod's annotation map will result in changing the json patch because annotations is a map reference.
	patch := jsonpatch.NewOperation(_addPatchOperation, _annotationPatchPath, annotations)
	return &patch, nil
}

// CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotation create a patch to delete ContainersVulnerabilityScanAnnotation if the pod's annotations contain ContainersVulnerabilityScanAnnotation.
// Otherwise, return nil
func CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotation(pod *corev1.Pod) (*jsonpatch.JsonPatchOperation, error) {
	// Create annotations map after deleting contracts.ContainersVulnerabilityScanInfoAnnotationName
	annotations, err := deleteAzdAnnotations(pod, contracts.ContainersVulnerabilityScanInfoAnnotationName)
	if err != nil {
		return nil, errors.Wrap(err, "AzdAnnotationsPatchGenerator failed to delete annotations")
	}
	// if annotations is nil, there is no need to delete ContainersVulnerabilityScanInfoAnnotation (pod's annotations are nil or contracts.ContainersVulnerabilityScanInfoAnnotationName don't exist)
	if annotations == nil {
		return nil, nil
	}
	// Create an add operation to annotations to add
	patch := jsonpatch.NewOperation(_addPatchOperation, _annotationPatchPath, annotations)
	return &patch, nil
}

// marshalAnnotationInnerObject marshaling provided object needed to be set as string in annotations to json represented string
func marshalAnnotationInnerObject(object interface{}) (string, error) {
	// Marshal object
	marshaledVulnerabilitySecInfo, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	// Cast to string
	ser := string(marshaledVulnerabilitySecInfo)
	return ser, nil
}

// updateAnnotations update the annotations of a given pod with the given key and value.
// If annotations map not exist - create a new map and add the key.
// Return the annotations map
func updateAnnotations(pod *corev1.Pod, key string, value string) (map[string]string, error){
	if pod == nil {
		return nil, errors.Wrap(utils.NilArgumentError, "updateAnnotations got nil pod")
	}
	annotations := pod.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		pod.SetAnnotations(annotations)
	}
	annotations[key]=value
	return annotations, nil
}

// deleteAzdAnnotations return the pod's annotations after deleting contracts.ContainersVulnerabilityScanInfoAnnotationName if exists otherwise returns nil.
func deleteAzdAnnotations(pod *corev1.Pod, key string) (map[string]string, error){
	if pod == nil {
		return nil, errors.Wrap(utils.NilArgumentError, "updateAnnotations got nil pod")
	}
	annotations := pod.GetAnnotations()
	// no annotations - no need to delete
	if annotations == nil {
		return nil, nil
	}
	// key don't exist - no need to delete
	if _, ok := annotations[key]; !ok{
		return nil, nil
	}
	// key exist - delete the key
	delete(annotations, key)
	return annotations, nil
}