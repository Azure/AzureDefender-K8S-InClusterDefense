package admisionrequest

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
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
	_errMsgMarshal                  = "admisionrequest.extractor: failed in json.Marshal"
	_errMsgJsonToYamlConversionFail = "admisionrequest.extractor: failed to convert json to yaml node"
	_errMsgInvalidPath              = "admisionrequest.extractor: failed to access the given path"
	_errMsgUnexpectedResource       = "admisionrequest.extractor: expected workload resource"
	imagePullSecretsConst           = "imagePullSecrets"
	metadataConst                   = "metadata"
	ownerReferencesConst            = "ownerReferences"
	containersConst                 = "containers"
	initContainersConst             = "initContainers"
	specConst                       = "spec"
	serviceAccountNameConst         = "serviceAccountName"
)

var (
	conventionalPodSpecPaths = [][]string{
		{"spec", "jobTemplate", "spec", "template", "spec"},
		{"spec", "template", "spec"},
		{"template", "spec"},
		{"spec"}} //from yaml.ConventionalContainersPaths, without containers(the last string in paths).
	kubernetesWorkloadResources =[]string {"Pod", "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet",
		"Job", "CronJob", "ReplicationController"} //https://kubernetes.io/docs/concepts/workloads/
	_errInvalidAdmission = errors.New(_errMsgInvalidAdmission)
	_errObjectNotFound = errors.New(_errMsgObjectNotFound)
	_errUnexpectedResource = errors.New(_errMsgUnexpectedResource)
)


// getContainers returns workload kubernetes resource's containers or initContainers(according to
// ContainersType).
func getContainers(specNode *yaml.RNode, containerType containersType) (containers []Container, err error) {
	containersInterface, pathErr := specNode.GetSlice(string(containerType))
	// if pathErr != nil it means that containerType is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	var list []Container
	bytes, err := json.Marshal(containersInterface)
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
	imagePullSecretsInterface, pathErr := specNode.GetSlice(imagePullSecretsConst)
	// if pathErr != nil it means that "imagePullSecrets" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	bytes, err := json.Marshal(imagePullSecretsInterface)
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
	metaNode, pathErr := utils.GoToDestNode(yamlNode, metadataConst)
	// if err != nil it means that "ownerReferences" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	sliceOwnerReferences, err := metaNode.GetSlice(ownerReferencesConst)
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

// stringInSlice return true if list contain str,false otherwise.
func stringInSlice(str string, list []string) bool {
	for _, listValue := range list {
		if listValue == str {
			return true
		}
	}
	return false
}

// ExtractWorkloadResourceFromAdmissionRequest return WorkloadResource object according
// to the information in admission.Request.
func (extractor *Extractor) ExtractWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error) {
	//Sanity checks of the application admission request
	if req == nil {
		return nil, _errInvalidAdmission
	}
	if len(req.Object.Raw) == 0 {
		return nil, _errObjectNotFound
	}
	if !stringInSlice(req.Kind.Kind, kubernetesWorkloadResources){
		return nil, _errUnexpectedResource
	}

	yamlFile, err := yaml.ConvertJSONToYamlNode(string(req.Object.Raw))
	if err != nil {
		return nil, errors.Wrap(err, _errMsgJsonToYamlConversionFail)
	}
	// return podspec yaml rNode.
	specNode, err := yaml.LookupFirstMatch(conventionalPodSpecPaths).Filter(yamlFile)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}

	// extract fields from admission request for WorkloadResource
	namespace := yamlFile.GetNamespace()
	annotation := yamlFile.GetAnnotations()
	containers, err := getContainers(specNode, containersConst)
	if err != nil {
		return nil, err
	}

	initContainers, err := getContainers(specNode, initContainersConst)
	if err != nil {
		return nil, err
	}

	imagePullSecrets, err := getImagePullSecrets(specNode)
	if err != nil {
		return nil, err
	}

	serviceAccountName, err := utils.GetValue(yamlFile, specConst, serviceAccountNameConst)
	if err != nil {
		return nil, err
	}

	ownerReferences, err := getOwnerReference(yamlFile)
	if err != nil {
		return nil, err
	}
	return newWorkLoadResource(newObjectMetadata(namespace,annotation,ownerReferences),
		newSpec(containers,initContainers,imagePullSecrets,serviceAccountName)), nil
}
