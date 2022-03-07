package admisionrequest

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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
	//from yaml.ConventionalContainersPaths, without containers(the last string in paths).
	conventionalPodSpecPaths = [][]string{
		{"spec", "jobTemplate", "spec", "template", "spec"},
		{"spec", "template", "spec"},
		{"template", "spec"},
		{"spec"}}
	kubernetesWorkloadResources = []string{"Pod", "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet",
		"Job", "CronJob", "ReplicationController"} //https://kubernetes.io/docs/concepts/workloads/
	_errInvalidAdmission   = errors.New(_errMsgInvalidAdmission)
	_errObjectNotFound     = errors.New(_errMsgObjectNotFound)
	_errUnexpectedResource = errors.New(_errMsgUnexpectedResource)
)

// GoToDestNode returns the *Rnode of the given path.
func (extractor *Extractor) GoToDestNode(yamlFile *yaml.RNode, path ...string) (destNode *yaml.RNode, err error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractWorkloadResourceFromAdmissionRequest")
	DestNode, err := yamlFile.Pipe(yaml.Lookup(path...))
	if err != nil {
		tracer.Error(err, _errMsgInvalidPath)
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}
	return DestNode, err
}

func getContainers(specNode *yaml.RNode) (containers []Container, initContainers []Container, err error) {
	allContainers := make([][]Container,2)
	for pathIndex, containerPath := range []string{containersConst, initContainersConst} {
		containersInterface, err := specNode.GetSlice(containerPath)
		// if err != nil it means that containerType is an empty field in admission request
		if err != nil {
			allContainers[pathIndex] = nil
			break
		}
		allContainers[pathIndex] = make([]Container, len(containersInterface))
		for i, containerObj := range containersInterface {
			v, ok := containerObj.(map[string]interface{})
			if ok == false {
				return nil, nil, err
			}
			allContainers[pathIndex][i].Image, ok = (v["image"]).(string)
			if ok == false {
				return nil, nil, err
			}
			allContainers[pathIndex][i].Name, ok = (v["name"]).(string)
			if ok == false {
				return nil, nil, err
			}
		}
	}
	return allContainers[0], allContainers[1], nil
}

// getImagePullSecrets returns workload kubernetes resource's image pull secrets.
func getImagePullSecrets(specNode *yaml.RNode) (secrets []corev1.LocalObjectReference, err error) {
	imagePullSecretsInterface, pathErr := specNode.GetSlice(imagePullSecretsConst)
	// if pathErr != nil it means that "imagePullSecrets" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	secrets = make([]corev1.LocalObjectReference, len(imagePullSecretsInterface))
	for i, sercet := range imagePullSecretsInterface {
		v, ok := sercet.(map[string]interface{})
		if ok == false {
			return nil, err
		}
		secrets[i].Name, ok = (v["name"]).(string)
		if ok == false {
			return nil, err
		}
	}
	return secrets, nil
}

// GetOwnerReference returns workload kubernetes resource's owner reference.
func (extractor *Extractor) getOwnerReference(yamlNode *yaml.RNode) (ownerReferences []OwnerReference, err error) {
	metaNode, pathErr := extractor.GoToDestNode(yamlNode, metadataConst)
	// if err != nil it means that "ownerReferences" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	sliceOwnerReferences, err := metaNode.GetSlice(ownerReferencesConst)
	if err != nil {
		return nil, nil
	}
	ownerReferences = make([]OwnerReference,len(sliceOwnerReferences))
	for i,reference := range sliceOwnerReferences{
		mapReference,ok := reference.(map[string]interface{})
		if ok == false {
			return nil, err
		}
		ownerReferences[i].APIVersion,ok = mapReference["apiVersion"].(string)
		if ok == false {
			return nil, err
		}
		ownerReferences[i].Kind,ok = mapReference["kind"].(string)
		if ok == false {
			return nil, err
		}
		ownerReferences[i].Name,ok = mapReference["name"].(string)
		if ok == false {
			return nil, err
		}

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

//Basics checks of the application admission request.
func reqBasicChecks(req *admission.Request) (err error) {
	if req == nil {
		return _errInvalidAdmission
	}
	if len(req.Object.Raw) == 0 {
		return _errObjectNotFound
	}
	if !stringInSlice(req.Kind.Kind, kubernetesWorkloadResources) {
		return _errUnexpectedResource
	}
	return nil
}

// ExtractMetadataFromAdmissionRequest return *ObjectMetadata object according
//// to the information in yamlFile.
func (extractor *Extractor) ExtractMetadataFromAdmissionRequest(yamlFile *yaml.RNode) (*ObjectMetadata, error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractMetadataFromAdmissionRequest")
	name := yamlFile.GetName()
	namespace := yamlFile.GetNamespace()
	annotation := yamlFile.GetAnnotations()
	ownerReferences, err := extractor.getOwnerReference(yamlFile)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}
	meta := newObjectMetadata(name, namespace, annotation, ownerReferences)
	return &meta, nil
}

// ExtractSpecFromAdmissionRequest return *PodSpec object according
//// to the information in yamlFile.
func (extractor *Extractor) ExtractSpecFromAdmissionRequest(yamlFile *yaml.RNode) (*PodSpec, error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractSpecFromAdmissionRequest")
	// return podspec yaml rNode.
	specNode, err := yaml.LookupFirstMatch(conventionalPodSpecPaths).Filter(yamlFile)
	if err != nil {
		tracer.Error(errors.Wrap(err, _errMsgInvalidPath), "")
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}
	containerList, initContainerList, err := getContainers(specNode)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}
	imagePullSecrets, err := getImagePullSecrets(specNode)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}

	// err ignore because serviceAccountName may not exist. in that case ot will
	// be assigned to empty string.
	serviceAccountName, err := specNode.GetString(serviceAccountNameConst)

	spec := newSpec(containerList, initContainerList, imagePullSecrets, serviceAccountName)
	return &spec, nil
}

// ExtractWorkloadResourceFromAdmissionRequest return WorkloadResource object according
// to the information in admission.Request.
func (extractor *Extractor) ExtractWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractWorkloadResourceFromAdmissionRequest")
	tracer.Info("ExtractWorkloadResourceFromAdmissionRequest Enter", "admission request", req)

	err = reqBasicChecks(req)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}
	yamlFile, err := yaml.ConvertJSONToYamlNode(string(req.Object.Raw))
	if err != nil {
		tracer.Error(errors.Wrap(err, _errMsgJsonToYamlConversionFail), "")
		return nil, errors.Wrap(err, _errMsgJsonToYamlConversionFail)
	}

	metadata, err := extractor.ExtractMetadataFromAdmissionRequest(yamlFile)
	if err != nil {
		return nil, err
	}

	spec, err := extractor.ExtractSpecFromAdmissionRequest(yamlFile)
	if err != nil {
		return nil, err
	}

	return newWorkLoadResource(*metadata, *spec), nil
}
