package utils

import (
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

