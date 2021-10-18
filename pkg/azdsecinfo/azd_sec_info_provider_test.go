package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	argDataProviderMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	tag2DigestResolverMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"testing"
	"time"
)

var (

	_imageRegistry = "playground.azurecr.io"
	_imageRepo= "testrepo"
	_imageTagTest1 = "1.0"
	_imageTagTest2 = "2.0"
	_imageOriginalTest1 = _imageRegistry + "/" + _imageRepo + ":" + _imageTagTest1
	_imageOriginalTest2 = _imageRegistry + "/" + _imageRepo + ":" + _imageTagTest2

	_containers = []corev1.Container{
		{
			Name:  "containerTest1",
			Image: _imageOriginalTest1,
		},
		{
			Name:  "containerTest2",
			Image: _imageOriginalTest2,
		},
	}

	imageRedTest1 = registry.NewTag(_imageOriginalTest1, _imageRegistry, _imageRepo, _imageTagTest1)
	resourceCtxTest1 = tag2digest.NewResourceContext("default", []string{}, "")
	digestTest1 = "sha256:9f9ed5fe24766b31bcb64aabba73e96cc5b7c2da578f9cd2fca20846cf5d7557"

	imageRedTest2 = registry.NewTag(_imageOriginalTest2, _imageRegistry, _imageRepo, _imageTagTest2)
	resourceCtxTest2 = tag2digest.NewResourceContext("default", []string{}, "")
	digestTest2 = "sha256:86a80e680602c613519a5af190219346230a3b02d98606727b9c8d47d8dc88ed"
)

type TestSuite struct {
	suite.Suite
	tag2DigestResolverMock *tag2DigestResolverMocks.ITag2DigestResolver
	argDataProviderMock *argDataProviderMocks.IARGDataProvider
}

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	// Update deployment - is needed for set default namespace as empty.
	utils.UpdateDeploymentForTests(&utils.DeploymentConfiguration{Namespace: "kube-system"})
	// Mock
	suite.tag2DigestResolverMock = &tag2DigestResolverMocks.ITag2DigestResolver{}
	suite.argDataProviderMock = &argDataProviderMocks.IARGDataProvider{}
}

func (suite *TestSuite) Test_getContainersVulnerabilityScanInfo_Run_In_Parallel() {
	suite.Fail("fail")
	run1 := measureTime(suite.getContainersVulnerabilityScanInfoTest, 2, 2)
	run2 := measureTime(suite.getContainersVulnerabilityScanInfoTest, 2, 0)
	suite.True(math.Abs(run1.Seconds()-run2.Seconds()) < 0.3)
}

func TestUpdateVulnSecInfoContainers(t *testing.T) {
	suite.Run(t, new(TestSuite))
}


func (suite *TestSuite)getContainersVulnerabilityScanInfoTest(waitFirstContainer time.Duration, waitSecondContainer time.Duration){
	containers := []corev1.Container{_containers[0], _containers[1]}
	pod := createPodForTests(containers, nil)
	suite.tag2DigestResolverMock.On("Resolve", imageRedTest1, resourceCtxTest1).Return(digestTest1, nil).Once().Run(func(args mock.Arguments) {
		time.Sleep(waitFirstContainer * time.Second)
	})
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", imageRedTest1.Registry(), imageRedTest1.Repository(), digestTest1).Return(contracts.Unscanned, nil, nil)
	suite.tag2DigestResolverMock.On("Resolve", imageRedTest2, resourceCtxTest2).Return(digestTest2, nil).Once().Run(func(args mock.Arguments) {
		time.Sleep(waitSecondContainer * time.Second)
	})
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", imageRedTest2.Registry(), imageRedTest2.Repository(), digestTest2).Return(contracts.Unscanned, nil, nil)
	azdSecInfoProvider := NewAzdSecInfoProvider(instrumentation.NewNoOpInstrumentationProvider(), suite.argDataProviderMock, suite.tag2DigestResolverMock)
	// Act
	res, _ := azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
	// Test
	suite.Equal(res[0].ScanStatus, contracts.Unscanned)
	suite.Equal(res[1].ScanStatus, contracts.Unscanned)
}

// measureTime runs the given function f and returns its run time
func measureTime(f func(time.Duration, time.Duration), waitFirstContainer time.Duration, waitSecondContainer time.Duration) time.Duration{
	start := time.Now()
	f(waitFirstContainer, waitSecondContainer)
	return time.Since(start)
}

func createPodForTests(containers []corev1.Container, initContainers []corev1.Container) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podTest",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{},
		Spec: corev1.PodSpec{
			Containers:     containers,
			InitContainers: initContainers,
		},
	}
}