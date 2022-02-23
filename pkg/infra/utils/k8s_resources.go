package utils

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	corev1 "k8s.io/api/core/v1"
)




// ExtractImagesFromPodSpec gets pod spec and returns all images used by the pod.
func ExtractImagesFromPodSpec(podSpec *corev1.PodSpec) []string {
	images := []string{}
	if podSpec == nil {
		return images
	}
	for _, initContainer := range podSpec.InitContainers {
		images = append(images, initContainer.Image)
	}
	for _, container := range podSpec.Containers {
		images = append(images, container.Image)
	}
	return images
}

// ExtractContainersFromPodSpecAsString gets pod spec and returns all containers as containerName:image used by the pod as String.
// For example appContainer:alpine
func ExtractContainersFromPodSpecAsString(podSpec *admisionrequest.SpecRes) []string {
	containers := []string{}
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
