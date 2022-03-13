package admisionrequest

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	_errMsgJsonToYamlConversionFail = "admisionrequest.extractor: failed to convert json to yaml node"
	_errMsgInvalidPath              = "admisionrequest.extractor: failed to access the given path"
	_imagePullSecretsConst          = "imagePullSecrets"
	_metadataConst                  = "metadata"
	_ownerReferencesConst           = "ownerReferences"
	_containersConst                = "containers"
	_initContainersConst            = "initContainers"
	_serviceAccountNameConst        = "serviceAccountName"
	_imageConst                     = "image"
	_nameConst                      = "name"
	_kindConst                      = "kind"
    _apiVersionConst                = "apiVersion"
)

var (
	KubernetesWorkloadResources = []string{"Pod", "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet",
		"Job", "CronJob", "ReplicationController"} //https://kubernetes.io/docs/concepts/workloads/
	// conventional pod spec paths for all kubernetes workload resources.
	//based on yaml.ConventionalContainersPaths var.
	conventionalPodSpecPaths = [][]string{
		{"spec", "jobTemplate", "spec", "template", "spec"}, // CronJob
		{"spec", "template", "spec"},// Deployment, ReplicaSet, StatefulSet, DaemonSet,Job, ReplicationController
		{"spec"}} // Pod
	_errInvalidAdmission   = errors.New("admisionrequest.extractor: admission request was nil")
	_errObjectNotFound     = errors.New("admisionrequest.extractor: request did not include object")
	_errUnexpectedResource = errors.New("admisionrequest.extractor: expected workload resource")
)

// ExtractWorkloadResourceFromAdmissionRequest return WorkloadResource object according
// to the information in admission.Request.
func (extractor *Extractor) ExtractWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractWorkloadResourceFromAdmissionRequest")
	tracer.Info("ExtractWorkloadResourceFromAdmissionRequest Enter", "admission request", req)

	err = extractor.isRequestValid(req)
	if err != nil {
		return nil, err
	}

	objectRequest := string(req.Object.Raw)
	tracer.Info("kubernetes resource in admission request: ","object: ", objectRequest)
	objectRequestYaml, err := yaml.ConvertJSONToYamlNode(objectRequest)
	if err != nil {
		err = errors.Wrap(err, _errMsgJsonToYamlConversionFail)
		tracer.Error(err, "")
		return nil, err
	}

	metadata, err := extractor.extractMetadataFromAdmissionRequest(objectRequestYaml)
	if err != nil {
		return nil, err
	}

	spec, err := extractor.extractSpecFromAdmissionRequest(objectRequestYaml)
	if err != nil {
		return nil, err
	}

	return newWorkLoadResource(*metadata, *spec), nil
}

// getContainers returns workload kubernetes resource's containers and initContainers.
func getContainers(specRoot *yaml.RNode) (containers []Container, initContainers []Container, err error) {
	allContainers := make([][]Container,2)
	for pathIndex, containerPath := range []string{_containersConst, _initContainersConst} {
		containersInterface, err := specRoot.GetSlice(containerPath)
		// if err != nil it means that containerType is an empty field in admission request
		if err != nil {
			allContainers[pathIndex] = nil
			continue
		}
		allContainers[pathIndex] = make([]Container, len(containersInterface))
		for i, containerObj := range containersInterface {
			v, ok := containerObj.(map[string]interface{})
			if ok == false {
				return nil, nil, err
			}
			allContainers[pathIndex][i].Image, ok = (v[_imageConst]).(string)
			if ok == false {
				return nil, nil, err
			}
			allContainers[pathIndex][i].Name, ok = (v[_nameConst]).(string)
			if ok == false {
				return nil, nil, err
			}
		}
	}
	return allContainers[0], allContainers[1], nil
}

// getImagePullSecrets returns workload kubernetes resource's image pull secrets.
func getImagePullSecrets(specRoot *yaml.RNode) (secrets []corev1.LocalObjectReference, err error) {
	imagePullSecretsInterface, pathErr := specRoot.GetSlice(_imagePullSecretsConst)
	// if pathErr != nil it means that "imagePullSecrets" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	secrets = make([]corev1.LocalObjectReference, len(imagePullSecretsInterface))
	for i, secret := range imagePullSecretsInterface {
		v, ok := secret.(map[string]interface{})
		if ok == false {
			return nil, err
		}
		secrets[i].Name, ok = (v[_nameConst]).(string)
		if ok == false {
			return nil, err
		}
	}
	return secrets, nil
}

// getOwnerReference returns workload kubernetes resource's owner reference.
func (extractor *Extractor) getOwnerReference(root *yaml.RNode) (ownerReferences []OwnerReference, err error) {
	metadata, pathErr := goToDestNode(root, _metadataConst)
	// if err != nil it means that "metadata" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	// if err != nil it means that "ownerReferences" is an empty field in admission request
	sliceOwnerReferences, err := metadata.GetSlice(_ownerReferencesConst)
	if err != nil {
		return nil, nil
	}
	ownerReferences = make([]OwnerReference,len(sliceOwnerReferences))
	for i,reference := range sliceOwnerReferences{
		mapReference,ok := reference.(map[string]interface{})
		if ok == false {
			return nil, err
		}
		ownerReferences[i].APIVersion,ok = mapReference[_apiVersionConst].(string)
		if ok == false {
			return nil, err
		}
		ownerReferences[i].Kind,ok = mapReference[_kindConst].(string)
		if ok == false {
			return nil, err
		}
		ownerReferences[i].Name,ok = mapReference[_nameConst].(string)
		if ok == false {
			return nil, err
		}

	}
	return ownerReferences, nil
}


//isRequestValid return error is request isn't valid, else returns nil.
func (extractor *Extractor) isRequestValid(req *admission.Request) (err error) {
	tracer := extractor.tracerProvider.GetTracer("isRequestValid")
	if req == nil {
		tracer.Error(_errInvalidAdmission, "")
		return _errInvalidAdmission
	}
	if len(req.Object.Raw) == 0 {
		tracer.Error(_errObjectNotFound, "")
		return _errObjectNotFound
	}
	if !StringInSlice(req.Kind.Kind, KubernetesWorkloadResources) {
		tracer.Error(_errUnexpectedResource, "")
		return _errUnexpectedResource
	}
	return nil
}

// extractMetadataFromAdmissionRequest return *ObjectMetadata object according
// to the information in yamlFile.
func (extractor *Extractor) extractMetadataFromAdmissionRequest(root *yaml.RNode) (metadata *ObjectMetadata,err error) {
	tracer := extractor.tracerProvider.GetTracer("extractMetadataFromAdmissionRequest")
	name := root.GetName()
	namespace := root.GetNamespace()
	annotations := root.GetAnnotations()
	if len(annotations) == 0{
		annotations = nil
	}
	ownerReferences, err := extractor.getOwnerReference(root)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}
	metadataObj := newObjectMetadata(name, namespace, annotations, ownerReferences)
	metadata = &metadataObj
	return metadata, nil
}

// extractSpecFromAdmissionRequest return *PodSpec object according
// to the information in yamlFile.
func (extractor *Extractor) extractSpecFromAdmissionRequest(root *yaml.RNode) (spec *PodSpec,err error) {
	tracer := extractor.tracerProvider.GetTracer("extractSpecFromAdmissionRequest")
	// Gets the matching pod spec path of the given root.
	podSpecPathFilter := yaml.LookupFirstMatch(conventionalPodSpecPaths)
	// Go to podSpec Node according to given podSpecPathFilter.
	specNode, err := podSpecPathFilter.Filter(root)
	if err != nil{
		//if err is not nil then it means that specNode wasn't found. spec.containers is mandatory, therefore the api
		//server will block the request - we don't want that our webhook will report error in this case.
		// Todo [t-ngandelman] log and metric to deal with possible error with the spec path.
		podSpec := newEmptySpec()
		return &podSpec,nil
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
	serviceAccountName, err := specNode.GetString(_serviceAccountNameConst)

	podSpec := newSpec(containerList, initContainerList, imagePullSecrets, serviceAccountName)
	return &podSpec, nil
}

