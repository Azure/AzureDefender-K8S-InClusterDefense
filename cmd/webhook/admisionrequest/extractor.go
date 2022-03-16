package admisionrequest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	_errMsgJsonToYamlConversionFail = "admisionrequest.extractor: failed to convert json to yaml node"
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
	KubernetesWorkloadResourcesSupported = []string{"Pod", "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet",
		"Job", "CronJob", "ReplicationController"} //https://kubernetes.io/docs/concepts/workloads/
	// conventional pod spec paths for all kubernetes workload resources.
	//based on yaml.ConventionalContainersPaths var.
	conventionalPodSpecPaths = [][]string{
		{"spec", "jobTemplate", "spec", "template", "spec"}, // CronJob
		{"spec", "template", "spec"},                        // Deployment, ReplicaSet, StatefulSet, DaemonSet,Job, ReplicationController
		{"spec"}} // Pod
	_errInvalidAdmission      = errors.New("admisionrequest.extractor: admission request was nil")
	_errWorkloadResourceEmpty = errors.New("admisionrequest.extractor: request did not include workload resource")
	_errUnexpectedResource    = errors.New("admisionrequest.extractor: expected workload resource")
	_errTypeConversionFailed = errors.New("admisionrequest.extractor: type conversion failed")
	_errWrongContainersPath  = errors.New("admisionrequest.extractor: wrong containers path")
)

// Extractor implements extractor from admission request to workload resource.
type Extractor struct {
	// tracerProvider of the handler
	tracerProvider trace.ITracerProvider
	// MetricSubmitter
	metricSubmitter metric.IMetricSubmitter
}

// NewExtractor Constructor for Extractor
func NewExtractor(instrumentationProvider instrumentation.IInstrumentationProvider) *Extractor {
	return &Extractor{
		tracerProvider:  instrumentationProvider.GetTracerProvider("Extractor"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}
}

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
	tracer.Info("kubernetes resource in admission request: ", "object: ", objectRequest)
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

	return newWorkLoadResource(metadata, spec), nil
}


func (extractor *Extractor) getContainersFromPath(specRoot *yaml.RNode, path ContainersPath) (containers []*Container, err error){
	tracer := extractor.tracerProvider.GetTracer("getContainersFromPath")
	if !(path == containersPath|| path == initContainersPath){
		tracer.Error(_errWrongContainersPath,"")
		return nil, _errWrongContainersPath
	}
	containersInterface, err := specRoot.GetSlice(string(path))
	if err != nil {
		tracer.Info(string(path)+" field is missing")
		return nil,nil
	}
	containers = make([]*Container, len(containersInterface))
	for i, containerObj := range containersInterface {
		v, ok := containerObj.(map[string]interface{})
		if ok == false {
			tracer.Error(_errTypeConversionFailed,"")
			return nil, _errTypeConversionFailed
		}
		containers[i] = &Container{}
		containers[i].Image, ok = (v[_imageConst]).(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed,"")
			return nil, _errTypeConversionFailed
		}
		containers[i].Name, ok = (v[_nameConst]).(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed,"")
			return nil, _errTypeConversionFailed
		}
	}
	return containers, nil
}

// getContainers returns workload kubernetes resource's containers and initContainers.
func (extractor *Extractor) getContainers(specRoot *yaml.RNode) (containers []*Container, initContainers []*Container, err error) {
	containers, err = extractor.getContainersFromPath(specRoot,containersPath)
	if err != nil {
		return nil, nil, err
	}
	initContainers, err = extractor.getContainersFromPath(specRoot,initContainersPath)
	if err != nil {
		return nil, nil, err
	}
	return containers, initContainers, nil
}

// getImagePullSecrets returns workload kubernetes resource's image pull secrets.
func (extractor *Extractor) getImagePullSecrets(specRoot *yaml.RNode) (secrets []*corev1.LocalObjectReference, err error) {
	tracer := extractor.tracerProvider.GetTracer("getImagePullSecrets")
	abstractImagePullSecrets, err := specRoot.GetSlice(_imagePullSecretsConst)
	// if pathErr != nil it means that "imagePullSecretsResource" is an empty field in admission request
	if err != nil {
		tracer.Info("imagePullSecrets field  is missing")
		return nil, nil
	}
	secrets = make([]*corev1.LocalObjectReference, len(abstractImagePullSecrets))
	for i, abstractImagePullSecret := range abstractImagePullSecrets {
		temporaryUnmarshalledSecret, ok := abstractImagePullSecret.(map[string]interface{})
		if ok == false {
			tracer.Error(_errTypeConversionFailed,"")
			return nil, _errTypeConversionFailed
		}
		secrets[i] = &corev1.LocalObjectReference{}
		secrets[i].Name, ok = (temporaryUnmarshalledSecret[_nameConst]).(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed,"")
			return nil, _errTypeConversionFailed
		}
	}
	return secrets, nil
}

// getOwnerReference returns workload kubernetes resource's owner reference.
func (extractor *Extractor) getOwnerReference(root *yaml.RNode) (abstractOwnerReferences []*OwnerReference, err error) {
	metadata, pathErr := utils.GoToDestNode(root, _metadataConst)
	// if err != nil it means that "metadata" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	// if err != nil it means that "abstractOwnerReferences" is an empty field in admission request
	sliceOwnerReferences, err := metadata.GetSlice(_ownerReferencesConst)
	if err != nil {
		return nil, nil
	}
	abstractOwnerReferences = make([]*OwnerReference, len(sliceOwnerReferences))
	for i, abstractOwnerReference := range sliceOwnerReferences {
		mapReference, ok := abstractOwnerReference.(map[string]interface{})
		if ok == false {
			return nil, _errTypeConversionFailed
		}
		abstractOwnerReferences[i] = &OwnerReference{}
		(*abstractOwnerReferences[i]).APIVersion, ok = mapReference[_apiVersionConst].(string)
		if ok == false {
			return nil, _errTypeConversionFailed
		}
		(*abstractOwnerReferences[i]).Kind, ok = mapReference[_kindConst].(string)
		if ok == false {
			return nil, _errTypeConversionFailed
		}
		(*abstractOwnerReferences[i]).Name, ok = mapReference[_nameConst].(string)
		if ok == false {
			return nil, _errTypeConversionFailed
		}
	}
	return abstractOwnerReferences, nil
}

//isRequestValid return error is request isn't valid, else returns nil.
func (extractor *Extractor) isRequestValid(req *admission.Request) (err error) {
	tracer := extractor.tracerProvider.GetTracer("isRequestValid")
	if req == nil {
		tracer.Error(_errInvalidAdmission, "")
		return _errInvalidAdmission
	}
	if len(req.Object.Raw) == 0 {
		tracer.Error(_errWorkloadResourceEmpty, "")
		return _errWorkloadResourceEmpty
	}
	if !utils.StringInSlice(req.Kind.Kind, KubernetesWorkloadResourcesSupported) {
		tracer.Error(_errUnexpectedResource, "")
		return _errUnexpectedResource
	}
	return nil
}

// extractMetadataFromAdmissionRequest extracts *ObjectMetadata from admission request.
func (extractor *Extractor) extractMetadataFromAdmissionRequest(root *yaml.RNode) (metadata *ObjectMetadata, err error) {
	tracer := extractor.tracerProvider.GetTracer("extractMetadataFromAdmissionRequest")
	name := root.GetName()
	// name is mandatory therefore if it is empty the api server will block the request -
	// we don't want that our webhook will report error in this case.
	if name == "" {
		tracer.Info("name is empty")
	}
	namespace := root.GetNamespace()
	annotations := root.GetAnnotations()
	// If annotations field missing from yaml, annotations is nil by default but GetAnnotations returns empty map.
	if len(annotations) == 0 {
		annotations = nil
	}
	ownerReferences, err := extractor.getOwnerReference(root)
	if err != nil {
		err = errors.Wrap(err, "Couldn't get owner references from metadata: error encountered")
		tracer.Error(err, "")
		return nil, err
	}

	return newObjectMetadata(name, namespace, annotations, ownerReferences), nil
}

// extractSpecFromAdmissionRequest extracts *PodSpec from admission request.
func (extractor *Extractor) extractSpecFromAdmissionRequest(root *yaml.RNode) (spec *PodSpec, err error) {
	tracer := extractor.tracerProvider.GetTracer("extractSpecFromAdmissionRequest")
	// Gets the matching pod spec path of the given root.
	podSpecPathFilter := yaml.LookupFirstMatch(conventionalPodSpecPaths)
	// Go to podSpec Node according to given podSpecPathFilter.
	specNode, err := podSpecPathFilter.Filter(root)
	if err != nil {
		// if err is not nil then it means that specNode wasn't found. spec.containers is mandatory, therefore the api
		// server will block the request - we don't want that our webhook will report error in this case.
		// Todo [t-ngandelman] log and metric to deal with possible error with the spec path.
		podSpec := newEmptySpec()
		return podSpec, nil
	}
	containerList, initContainerList, err := extractor.getContainers(specNode)
	if err != nil {
		err = errors.Wrap(err, "Couldn't get containers/init containers from spec: error encountered")
		tracer.Error(err, "")
		return nil, err
	}
	imagePullSecrets, err := extractor.getImagePullSecrets(specNode)
	if err != nil {
		err = errors.Wrap(err, "Couldn't get  image pull secrets from spec: error encountered")
		tracer.Error(err, "")
		return nil, err
	}

	serviceAccountName, err := specNode.GetString(_serviceAccountNameConst)
	if serviceAccountName == ""{
		tracer.Info("serviceAccountName is empty field")
	}
	tracer.Info("spec: ", "containers", containerList,"initContainers",initContainerList,
		"initContainer", imagePullSecrets, "serviceAccountName", serviceAccountName)
	return newSpec(containerList, initContainerList, imagePullSecrets, serviceAccountName), nil
}
