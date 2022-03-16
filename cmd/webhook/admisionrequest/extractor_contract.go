package admisionrequest

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

// ContainersPath Declare ContainersPath enum.
type ContainersPath string

const (
	containersPath     ContainersPath = _containersConst
	initContainersPath ContainersPath = _initContainersConst
)


// OwnerReference contains information to let you identify an owning
// object. An owning object must be in the same namespace as the dependent, or
// be cluster-scoped, so there is no namespace field.
type OwnerReference struct {
	// API version of the referent.
	APIVersion string
	// Kind of the referent.
	Kind string
	// Name of the referent.
	Name string
}

// ObjectMetadata represents the metadata of WorkloadResource object.
type ObjectMetadata struct {
	// Name is the resource's name.
	Name string
	// Namespace defines the space within which each name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	Namespace string
	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	Annotations map[string]string
	// List of objects depended by this object. If ALL objects in the list have
	// been deleted, this object will be garbage collected. If this object is managed by a controller,
	// then an entry in this list will point to this controller, with the controller field set to true.
	// There cannot be more than one managing controller.
	OwnerReferences []*OwnerReference
}

// newObjectMetadata initialize ObjectMetadata object.
func newObjectMetadata(name string, namespace string, annotation map[string]string, ownerReferences []*OwnerReference) (metadata *ObjectMetadata) {
	return &ObjectMetadata{Name: name, Namespace: namespace, Annotations: annotation, OwnerReferences: ownerReferences}
}

// Container represents container object.
type Container struct {
	// Container name
	Name string
	// Image name
	Image string
}

// PodSpec represents a specification of the desired behavior of the WorkloadResource.
type PodSpec struct {
	// Containers is a list of Container objects.
	Containers []*Container
	// InitContainers is a list of Container objects.
	InitContainers []*Container
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.  For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	ImagePullSecrets []*corev1.LocalObjectReference
	// ServiceAccountName is the name of the ServiceAccount to use to run this WorkloadResource
	// The WorkloadResource will be allowed to use secrets referenced by the ServiceAccount
	ServiceAccountName string
}

// newSpec initialize PodSpec object.
func newSpec(containers []*Container, initContainers []*Container, imagePullSecrets []*corev1.LocalObjectReference,
	serviceAccountName string) (spec *PodSpec) {
	return &PodSpec{Containers: containers,
		InitContainers: initContainers,
		ImagePullSecrets: imagePullSecrets,
		ServiceAccountName: serviceAccountName}
}

// newEmptySpec initialize empty PodSpec object.
func newEmptySpec() (spec *PodSpec) {
	return newSpec(nil, nil, nil, "")
}

// ExtractContainersFromPodSpecAsString gets pod spec and returns all containers as containerName:image used by the pod as String.
// For example appContainer:alpine
func (podSpec *PodSpec) ExtractContainersFromPodSpecAsString() []string {
	var containers []string
	if podSpec == nil {
		return containers
	}
	for _, initContainer := range podSpec.InitContainers {
		containers = append(containers, fmt.Sprintf("%s:%s", initContainer.Name, initContainer.Image))
	}
	for _, container := range podSpec.Containers {
		containers = append(containers, fmt.Sprintf("%s:%s", container.Name, container.Image))
	}
	return containers
}

// WorkloadResource represents an abstraction of a kubernetes workload resources such as:
// Pod, Deployments, ReplicaSet, StatefulSets, DaemonSet, Jobs, CronJob and ReplicationController.
type WorkloadResource struct {
	// WorkloadResource metadata
	Metadata *ObjectMetadata
	// Spec defines the behavior of a WorkloadResource.
	Spec *PodSpec
}

// newWorkLoadResource initialize WorkloadResource object.
func newWorkLoadResource(metadata *ObjectMetadata, spec *PodSpec) (workLoadResource *WorkloadResource) {
	return &WorkloadResource{Metadata: metadata, Spec: spec}
}

