package azdsecinfo
//
//import (
//	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
//	argDataProviderMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/mocks"
//	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
//	tag2DigestResolverMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest/mocks"
//	"github.com/stretchr/testify/suite"
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//var (
//	_containers = []corev1.Container{
//		{
//			Name:  "containerTest1",
//			Image: "image1.com",
//		},
//		{
//			Name:  "containerTest2",
//			Image: "image2.com",
//		},
//		{
//			Name:  "containerTest3",
//			Image: "image3.com",
//		},
//	}
//
//	_firstContainerVulnerabilityScanInfo  = &contracts.ContainerVulnerabilityScanInfo{Name: "Lior"}
//	_secondContainerVulnerabilityScanInfo = &contracts.ContainerVulnerabilityScanInfo{Name: "Or"}
//)
//
//type TestSuite struct {
//	suite.Suite
//	tag2DigestResolverMock *tag2DigestResolverMocks.ITag2DigestResolver
//	argDataProviderMock *argDataProviderMocks.IARGDataProvider
//}
//
//func (suite *TestSuite) Test_Handle_DryRunTrue_ShouldNotPatched() {
//	// Setup
//	containers := []corev1.Container{_containers[0]}
//	pod := createPodForTests(containers, nil)
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo}
//	suite.tag2DigestResolverMock.On("Resolve", &pod.Spec, &pod.ObjectMeta, &pod.TypeMeta).Return(expectedInfo, nil).Once()
//	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", &pod.Spec, &pod.ObjectMeta, &pod.TypeMeta).Return(expectedInfo, nil).Once()
//
//	azdSecInfoProvider := NewAzdSecInfoProvider(instrumentation.NewNoOpInstrumentationProvider(), suite.argDataProviderMock, suite.tag2DigestResolverMock)
//	// Act
//	res, _ := azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
//	// Test
//	suite.Equal(metav1.StatusReason(_notPatchedHandlerDryRunReason), resp.Result.Reason)
//	suite.Emptyf(resp.Patches, "response.Patches should be empty on dryrun mode")
//}
//
//func createPodForTests(containers []corev1.Container, initContainers []corev1.Container) *corev1.Pod {
//	return &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "podTest",
//			Namespace: "default",
//		},
//		TypeMeta: metav1.TypeMeta{},
//		Spec: corev1.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}