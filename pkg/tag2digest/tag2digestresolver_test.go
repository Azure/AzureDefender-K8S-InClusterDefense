package tag2digest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuite struct {
	suite.Suite
	cacheClientMock *mocks.ICacheClient
}

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	// Update deployment - is needed for set default namespace as empty.
	utils.UpdateDeploymentForTests(&utils.DeploymentConfiguration{Namespace: "kube-system"})
	// Mock
	suite.cacheClientMock = &mocks.ICacheClient{}
}

//func (suite *TestSuite) Test_Handle_ZeroContainerTwoInitContainer_ShouldPatchedTwo() {
//	// Setup
//	containers := []corev1.Container{_containers[0], _containers[1]}
//	pod := createPodForTests(nil, containers)
//	req := createRequestForTests(pod)
//
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo, _secondContainerVulnerabilityScanInfo}
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", &pod.Spec, &pod.ObjectMeta, &pod.TypeMeta).Return(expectedInfo, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//	resolver :=  NewTag2DigestResolver(instrumentation.NewNoOpInstrumentationProvider(), registryClient registry.IRegistryClient, redisCache cache.ICacheClient) *Tag2DigestResolver {
//
//	//Act
//	resp := handler.Handle(context.Background(), *req)
//
//	//Test
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//	suite.checkPatch(expectedInfo, patch)
//}


func TestTagToDigestResolver(t *testing.T) {
	suite.Run(t, new(TestSuite))
}