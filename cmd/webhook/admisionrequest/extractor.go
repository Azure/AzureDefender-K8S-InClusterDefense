package admisionrequest

import (
	"encoding/json"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	_errMsgObjectNotFound           = "admisionrequest.extractor: request did not include object"
	_errMsgInvalidAdmission         = "admisionrequest.extractor: admission request was nil"
	_errMsgUnmarshal                = "admisionrequest.extractor: failed in json.Unmarshal"
	_errMsgMarshal                  = "admisionrequest.extractor: failed in json.Unmarshal"
	_errMsgJsonToYamlConversionFail = "admisionrequest.extractor: failed to convert json to yaml node."
	_errMsgInvalidPath              = "admisionrequest.extractor: failed to access the given path ."
)

var (
	ConventionalPodSpecPaths = [][]string{
		{"spec", "jobTemplate", "spec", "template", "spec"},
		{"spec", "template", "spec"},
		{"template", "spec"},
		{"spec"}}
)

// goToDestNode returns the *Rnode of the given path.
func goToDestNode(yamlFile *yaml.RNode, path ...string) (destNode *yaml.RNode, err error) {
	DestNode, err := yamlFile.Pipe(yaml.Lookup(path...))
	if err != nil {
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}
	return DestNode, err
}

// getValue returns a string value that the given path contains, can be empty.
func getValue(yamlFile *yaml.RNode, path ...string) (value string, err error) {
	DestNode, pathErr := goToDestNode(yamlFile, path...)
	if err != nil {
		return "", pathErr
	}
	val := yaml.GetValue(DestNode)
	return val, nil
}

// getContainers returns workload kubernetes resource's containers or initContainers(according to
// ContainersType).
func getContainers(specNode *yaml.RNode, ContainersType string) (containers []Container, err error) {
	con, err := specNode.GetSlice(ContainersType)
	if err != nil {
		return nil, nil
	}
	var list []Container
	bytes, err := json.Marshal(con)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgMarshal)
	}
	err = json.Unmarshal(bytes, &list)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgUnmarshal)
	}
	return list, nil
}

// getImagePullSecrets returns workload kubernetes resource's image pull secrets.
func getImagePullSecrets(specNode *yaml.RNode) (secrets []corev1.LocalObjectReference, err error) {
	sliceImagePullSecrets, err := specNode.GetSlice("imagePullSecrets")
	if err != nil {
		return nil, nil
	}
	bytes, err := json.Marshal(sliceImagePullSecrets)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgMarshal)
	}
	err = json.Unmarshal(bytes, &secrets)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgUnmarshal)
	}
	return secrets, nil
}

// GetOwnerReference returns workload kubernetes resource's owner reference.
func getOwnerReference(yamlNode *yaml.RNode) (ownerReferences []metav1.OwnerReference, err error) {
	metaNode, pathErr := goToDestNode(yamlNode, "metadata")
	if err != nil {
		return nil, pathErr
	}
	sliceOwnerReferences, err := metaNode.GetSlice("ownerReferences")
	if err != nil {
		return nil, nil
	}
	bytes, err := json.Marshal(sliceOwnerReferences)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgMarshal)
	}
	err = json.Unmarshal(bytes, &ownerReferences)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgUnmarshal)
	}
	return ownerReferences, nil
}

// NewWorkLoadResource initialize WorkLoadResource object.
func newWorkLoadResource(namespace string, annotation map[string]string, ownerReferences []metav1.OwnerReference,
	containers []Container, initContainers []Container,imagePullSecrets []corev1.LocalObjectReference,
	serviceAccountName string) (workLoadResource *WorkLoadResource){
	spec := PodSpec{Containers: containers, InitContainers: initContainers,
		ImagePullSecrets: imagePullSecrets, ServiceAccountName: serviceAccountName}
	metadata := ObjectMetadata{Namespace: namespace, Annotation: annotation, OwnerReferences: ownerReferences}
	resource := WorkLoadResource{Metadata:metadata, Spec:spec}
	return &resource
}

// GetWorkloadResourceFromAdmissionRequest return WorkLoadResource object according
// to the information in admission.Request.
func GetWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkLoadResource, err error) {
	if req == nil {
		return nil, errors.New(_errMsgInvalidAdmission)
	}
	if len(req.Object.Raw) == 0 {
		return nil, errors.New(_errMsgObjectNotFound)
	}

	yamlFile, err := yaml.ConvertJSONToYamlNode(string(req.Object.Raw))
	if err != nil {
		return nil, errors.Wrap(err, _errMsgJsonToYamlConversionFail)
	}

	namespace := yamlFile.GetNamespace()
	annotation := yamlFile.GetAnnotations()
	// return the node of the Rnode's podspec.
	specNode, err := yaml.LookupFirstMatch(ConventionalPodSpecPaths).Filter(yamlFile)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}

	containers, err := getContainers(specNode, "containers")
	if err != nil {
		return nil, err
	}

	initContainers, err := getContainers(specNode, "initContainers")
	if err != nil {
		return nil, err
	}

	imagePullSecrets, err := getImagePullSecrets(specNode)
	if err != nil {
		return nil, err
	}

	serviceAccountName, err := getValue(yamlFile, "spec", "serviceAccountName")
	if err != nil {
		return nil, err
	}

	ownerReferences, err := getOwnerReference(yamlFile)
	if err != nil {
		return nil, err
	}
	return newWorkLoadResource(namespace, annotation, ownerReferences, containers, initContainers, imagePullSecrets, serviceAccountName), nil
}
