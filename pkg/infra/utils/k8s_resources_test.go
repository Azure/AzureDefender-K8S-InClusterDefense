package utils

import (
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

var ()

type K8sTestSuite struct {
	suite.Suite
	firstContainer  corev1.Container
	secondContainer corev1.Container
}

func (suite *K8sTestSuite) SetupTest() {
	suite.firstContainer = corev1.Container{Image: "first"}
	suite.secondContainer = corev1.Container{Image: "second"}

}

func (suite *K8sTestSuite) Test_ExtractImagesFromPodSpec_get_nil_pod_spec() {
	images := ExtractImagesFromPodSpec(nil)
	suite.Equal(0, len(images))
}

func (suite *K8sTestSuite) Test_ExtractImagesFromPodSpec_NilInitContainer() {
	podSpec := createPodSpec(nil, []corev1.Container{suite.firstContainer})
	images := ExtractImagesFromPodSpec(podSpec)
	suite.Equal(1, len(images))
	suite.Equal(suite.firstContainer.Image, images[0])
}

func (suite *K8sTestSuite) Test_ExtractImagesFromPodSpec_NilContainer() {
	podSpec := createPodSpec([]corev1.Container{suite.firstContainer}, nil)
	images := ExtractImagesFromPodSpec(podSpec)
	suite.Equal(1, len(images))
	suite.Equal(suite.firstContainer.Image, images[0])
}

func (suite *K8sTestSuite) Test_ExtractImagesFromPodSpec_NotNil() {
	podSpec := createPodSpec([]corev1.Container{suite.firstContainer}, []corev1.Container{suite.secondContainer})
	images := ExtractImagesFromPodSpec(podSpec)
	suite.Equal(2, len(images))
	suite.Equal(suite.firstContainer.Image, images[0])
	suite.Equal(suite.secondContainer.Image, images[1])
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestK8sUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(K8sTestSuite))
}

func createPodSpec(initContainers []corev1.Container, containers []corev1.Container) *corev1.PodSpec {
	return &corev1.PodSpec{
		InitContainers: initContainers,
		Containers:     containers,
	}
}
