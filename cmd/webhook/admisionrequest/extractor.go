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
// ContainersPath Declare ContainersPath enum.
type ContainersPath string

const (
	_errMsgJsonToYamlConversionFail = "Failed to convert json to yaml node"
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
	containersPath     ContainersPath = _containersConst
	initContainersPath ContainersPath = _initContainersConst
)

var (
	// conventional pod spec paths for all kubernetes workload resources.
	//based on yaml.ConventionalContainersPaths var.
	_conventionalPodSpecPaths = [][]string{
		{"spec", "jobTemplate", "spec", "template", "spec"}, // CronJob
		{"spec", "template", "spec"},                        // Deployment, ReplicaSet, StatefulSet, DaemonSet,Job, ReplicationController
		{"spec"}}                                            // Pod
	_errInvalidAdmission      = errors.New("Admission request was nil")
	_errWorkloadResourceEmpty = errors.New("Request did not include workload resource")
	_errTypeConversionFailed  = errors.New("Type conversion failed")
	_errWrongContainersPath   = errors.New("Wrong containers path")
)

// ExtractorConfiguration configuration for extractor
type ExtractorConfiguration struct {
	SupportedKubernetesWorkloadResources []string
}

// IExtractor represents interface for admission request extractor.
type IExtractor interface {
	// ExtractWorkloadResourceFromAdmissionRequest return WorkloadResource object according
	// to the information in admission.Request.
	ExtractWorkloadResourceFromAdmissionRequest(req *admission.Request) (*WorkloadResource, error)
}

// Extractor implements IExtractor interface
var _ IExtractor = (*Extractor)(nil)

// Extractor implements extractor from admission request to workload resource.
type Extractor struct {
	// tracerProvider of the handler
	tracerProvider trace.ITracerProvider
	// MetricSubmitter
	metricSubmitter metric.IMetricSubmitter

	configuration *ExtractorConfiguration
}

// NewExtractor Constructor for Extractor
func NewExtractor(instrumentationProvider instrumentation.IInstrumentationProvider, configuration *ExtractorConfiguration) *Extractor {
	return &Extractor{
		tracerProvider:  instrumentationProvider.GetTracerProvider("Extractor"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
		configuration: configuration,
	}
}

// ExtractWorkloadResourceFromAdmissionRequest return WorkloadResource object according
// to the information in admission.Request.
func (extractor *Extractor) ExtractWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractWorkloadResourceFromAdmissionRequest")
	tracer.Info("ExtractWorkloadResourceFromAdmissionRequest Enter", "admission request", req)

	err = extractor.isRequestValid(req)
	if err != nil {
		err = errors.Wrap(err,"request isn't valid")
		tracer.Error(err,"")
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
		err = errors.Wrap(err,"failed to extract metadata from admission request")
		tracer.Error(err, "")
		return nil, err
	}

	spec, err := extractor.extractSpecFromAdmissionRequest(objectRequestYaml)
	if err != nil {
		return nil, err
	}

	workloadResource := newWorkLoadResource(metadata, spec)
	tracer.Info("workload resource :",workloadResource)
	return workloadResource, nil
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
	if !utils.StringInSlice(req.Kind.Kind, extractor.configuration.SupportedKubernetesWorkloadResources) {
		err = errors.New(req.Kind.Kind+" is unsupported kind of workload resource")
		tracer.Error(err ,"")
		return err
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
	metadata = newObjectMetadata(name, namespace, annotations, ownerReferences)
	return metadata, nil
}

// extractSpecFromAdmissionRequest extracts *PodSpec from admission request.
func (extractor *Extractor) extractSpecFromAdmissionRequest(root *yaml.RNode) (spec *PodSpec, err error) {
	tracer := extractor.tracerProvider.GetTracer("extractSpecFromAdmissionRequest")
	// Gets the matching pod spec path of the given root.
	podSpecPathFilter := yaml.LookupFirstMatch(_conventionalPodSpecPaths)
	// Go to podSpec Node according to given podSpecPathFilter.
	specNode, err := podSpecPathFilter.Filter(root)
	if err != nil {
		tracer.Error(err,"")
		return nil, err
	}

	if specNode == nil{
		// spec.containers is mandatory, therefore the api server will block the request.
		// we don't want that our webhook will report error in this case.
		tracer.Info("spec field is missing. Api server should have blocked the request")
		return newEmptySpec(), err
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
	if serviceAccountName == "" {
		tracer.Info("serviceAccountName is empty field")
	}
	spec = newSpec(containerList, initContainerList, imagePullSecrets, serviceAccountName)
	return spec, nil
}

// getImagePullSecrets returns workload kubernetes resource's image pull secrets.
func (extractor *Extractor) getImagePullSecrets(specRoot *yaml.RNode) (secrets []*corev1.LocalObjectReference, err error) {
	tracer := extractor.tracerProvider.GetTracer("getImagePullSecrets")
	abstractImagePullSecrets, err := specRoot.GetSlice(_imagePullSecretsConst)
	// if err != nil it means that "imagePullSecretsResource" is an empty field in admission request
	if err != nil {
		tracer.Info("imagePullSecrets field  is missing")
		return nil, nil
	}
	secrets = make([]*corev1.LocalObjectReference, len(abstractImagePullSecrets))
	for i, abstractImagePullSecret := range abstractImagePullSecrets {
		temporaryUnmarshalledSecret, ok := abstractImagePullSecret.(map[string]interface{})
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert secret interface to map")
			return nil, _errTypeConversionFailed
		}
		secretName, ok := (temporaryUnmarshalledSecret[_nameConst]).(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert secret name interface to string")
			return nil, _errTypeConversionFailed
		}
		secret := &corev1.LocalObjectReference{}
		secret.Name = secretName
		secrets[i] = secret
	}
	tracer.Info("secrets array: ", secrets)
	return secrets, nil
}

// getOwnerReference returns workload kubernetes resource's owner reference.
func (extractor *Extractor) getOwnerReference(root *yaml.RNode) (ownerReferences []*OwnerReference, err error) {
	tracer := extractor.tracerProvider.GetTracer("getOwnerReference")
	metadata, err := utils.GoToDestNode(root, _metadataConst)
	// if err != nil it means that "metadata" is an empty field in admission request
	if err != nil {
		tracer.Info("metadata field is missing")
		return nil, nil
	}
	// if err != nil it means that "ownerReferences" is an empty field in admission request
	sliceOwnerReferences, err := metadata.GetSlice(_ownerReferencesConst)
	if err != nil {
		tracer.Info("ownerReferences field is missing")
		return nil, nil
	}

	ownerReferences = make([]*OwnerReference, len(sliceOwnerReferences))
	for i, abstractOwnerReference := range sliceOwnerReferences {
		mapReference, ok := abstractOwnerReference.(map[string]interface{})
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert owner reference interface to map")
			return nil, _errTypeConversionFailed
		}
		apiVersion, ok := mapReference[_apiVersionConst].(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert owner reference api version interface to string")
			return nil, _errTypeConversionFailed
		}
		kind, ok := mapReference[_kindConst].(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert owner reference kind interface to string")
			return nil, _errTypeConversionFailed
		}
		name, ok := mapReference[_nameConst].(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert owner reference name interface to string")
			return nil, _errTypeConversionFailed
		}
		ownerReference := &OwnerReference{}
		ownerReference.APIVersion = apiVersion
		ownerReference.Kind = kind
		ownerReference.Name = name
		ownerReferences[i] = ownerReference
	}
	tracer.Info("ownerReferences array: ", ownerReferences)
	return ownerReferences, nil
}

// getContainers returns workload kubernetes resource's containers and initContainers.
func (extractor *Extractor) getContainers(specRoot *yaml.RNode) (containers []*Container, initContainers []*Container, err error) {
	containers, err = extractor.getContainersFromPath(specRoot, containersPath)
	if err != nil {
		return nil, nil, err
	}
	initContainers, err = extractor.getContainersFromPath(specRoot, initContainersPath)
	if err != nil {
		return nil, nil, err
	}
	return containers, initContainers, nil
}

func (extractor *Extractor) getContainersFromPath(specRoot *yaml.RNode, path ContainersPath) (containers []*Container, err error) {
	tracer := extractor.tracerProvider.GetTracer("getContainersFromPath")
	if !(path == containersPath || path == initContainersPath) {
		tracer.Error(_errWrongContainersPath, "")
		return nil, _errWrongContainersPath
	}
	containersInterface, err := specRoot.GetSlice(string(path))
	if err != nil {
		tracer.Info(string(path) + " field is missing")
		return nil, nil
	}
	containers = make([]*Container, len(containersInterface))
	for i, containerObj := range containersInterface {
		tempContainer, ok := containerObj.(map[string]interface{})
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert container interface to map")
			return nil, _errTypeConversionFailed
		}
		imageName, ok := (tempContainer[_imageConst]).(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert image name interface to string")
			return nil, _errTypeConversionFailed
		}
		containerName, ok := (tempContainer[_nameConst]).(string)
		if ok == false {
			tracer.Error(_errTypeConversionFailed, "Failed to convert container name interface to string")
			return nil, _errTypeConversionFailed
		}
		container := &Container{}
		container.Image, container.Name = imageName,  containerName
		containers[i] = container
	}
	tracer.Info(string(path)+" array: ",containers)
	return containers, nil
}

