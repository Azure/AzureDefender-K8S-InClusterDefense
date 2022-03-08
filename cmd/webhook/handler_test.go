/////*
////TODO  Add tests that the patching format is match to our policy.
////*/
package webhook
//
//import (
//	"context"
//	"encoding/json"
//	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
//	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
//	azdsecinfoMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/mocks"
//	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
//	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
//	"github.com/pkg/errors"
//	"github.com/stretchr/testify/suite"
//	"gomodules.xyz/jsonpatch/v2"
//	admissionv1 "k8s.io/api/admission/v1"
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"log"
//	"net/http"
//	"reflect"
//	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
//	"testing"
//	"time"
//)
//
//const (
//	_expectedTestAddPatchOperation   = "add"
//	_expectedTestAnnotationPatchPath = "/metadata/annotations"
//	_annotationTestKeyOne = "cluster-autoscaler.kubernetes.io/safe-to-evict"
//	_annotationTestValueOne = "true"
//	_annotationTestKeyTwo = "container.seccomp.security.alpha.kubernetes.io/manager"
//	_annotationTestValueTwo = "runtime/default"
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
//	_containersAdmision = []admisionrequest.Container{
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
//
//	_firstContainerVulnerabilityScanInfo  = &contracts.ContainerVulnerabilityScanInfo{Name: "Lior"}
//	_secondContainerVulnerabilityScanInfo = &contracts.ContainerVulnerabilityScanInfo{Name: "Or"}
//
//	_expectedPatchForErrorEncountered1 = jsonpatch.Operation{
//		Operation: "add",
//		Path: _expectedTestAnnotationPatchPath,
//		Value: map[string]string{
//			_annotationTestKeyOne : _annotationTestValueOne,
//			_annotationTestKeyTwo : _annotationTestValueTwo,
//		},
//	}
//
//	_expectedPatchForErrorEncountered2 = jsonpatch.Operation{
//		Operation: "add",
//		Path: _expectedTestAnnotationPatchPath,
//		Value: map[string]string{},
//	}
//)
//
//type TestSuite struct {
//	suite.Suite
//	azdSecProviderMock *azdsecinfoMocks.IAzdSecInfoProvider
//}
//
//// This will run before each test in the suite
//func (suite *TestSuite) SetupTest() {
//	// Update deployment - is needed for set default namespace as empty.
//	utils.UpdateDeploymentForTests(&utils.DeploymentConfiguration{Namespace: "kube-system"})
//	// Mock
//	suite.azdSecProviderMock = &azdsecinfoMocks.IAzdSecInfoProvider{}
//}
//
//func (suite *TestSuite) Test_Handle_DryRunTrue_ShouldNotPatched() {
//	// Setup
//	pod := createPodForTests([]corev1.Container{_containers[0]}, nil)
//	resource := createWorkloadResourceForTests([]admisionrequest.Container{_containersAdmision[0]}, nil)
//	req := createRequestForTests(pod)
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo}
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expectedInfo, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: true}, instrumentation.NewNoOpInstrumentationProvider())
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//	suite.Equal(admission.Allowed(string(_notPatchedHandlerDryRunReason)), resp)
//	suite.Equal(metav1.StatusReason(_notPatchedHandlerDryRunReason), resp.Result.Reason)
//	suite.Emptyf(resp.Patches, "response.Patches should be empty on dryrun mode")
//}
//
//func (suite *TestSuite) Test_Handle_DryRunFalse_ShouldPatched() {
//	// Setup
//	pod := createPodForTests([]corev1.Container{_containers[0]}, nil)
//	resource := createWorkloadResourceForTests([]admisionrequest.Container{_containersAdmision[0]}, nil)
//	req := createRequestForTests(pod)
//	expected := []*contracts.ContainerVulnerabilityScanInfo{
//		_firstContainerVulnerabilityScanInfo,
//	}
//
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expected, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//	suite.Equal(admission.Allowed(string(_patchedReason)).AdmissionResponse, resp.AdmissionResponse)
//	suite.Equal(metav1.StatusReason(_patchedReason), resp.Result.Reason)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//
//	suite.checkPatch(expected, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_RequestKindIsNotPod_ShouldNotPatched() {
//// Setup
//containers := []corev1.Container{_containers[0]}
//pod := createPodForTests(containers, nil)
//req := createRequestForTests(pod)
//req.Kind.Kind = "NotPodKind"
//
//handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//// Act
//resp := handler.Handle(context.Background(), *req)
//// Test
//suite.Equal(admission.Allowed(string(_noMutationForKindReason)), resp)
//suite.Equal(metav1.StatusReason(_noMutationForKindReason), resp.Result.Reason)
//suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_RequestDeleteOperation_ShouldNotPatched() {
//	// Setup
//	containers := []corev1.Container{_containers[0]}
//	pod := createPodForTests(containers, nil)
//	req := createRequestForTests(pod)
//	req.Operation = admissionv1.Delete
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//	suite.Equal(admission.Allowed(string(_noMutationForOperationReason)), resp)
//	suite.Equal(metav1.StatusReason(_noMutationForOperationReason), resp.Result.Reason)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_RequestConnectOperation_ShouldNotPatched() {
//	// Setup
//	containers := []corev1.Container{_containers[0]}
//	pod := createPodForTests(containers, nil)
//	req := createRequestForTests(pod)
//	req.Operation = admissionv1.Connect
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//	suite.Equal(admission.Allowed(string(_noMutationForOperationReason)), resp)
//	suite.Equal(metav1.StatusReason(_noMutationForOperationReason), resp.Result.Reason)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_RequestUpdateOperation_ShouldPatched() {
//	// Setup
//	pod := createPodForTests([]corev1.Container{_containers[0]}, nil)
//	resource := createWorkloadResourceForTests([]admisionrequest.Container{_containersAdmision[0]}, nil)
//	req := createRequestForTests(pod)
//	req.Operation = admissionv1.Update
//	expected := []*contracts.ContainerVulnerabilityScanInfo{
//		_firstContainerVulnerabilityScanInfo,
//	}
//
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expected, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//	suite.Equal(admission.Allowed(string(_patchedReason)).AdmissionResponse, resp.AdmissionResponse)
//	suite.Equal(metav1.StatusReason(_patchedReason), resp.Result.Reason)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//
//	suite.checkPatch(expected, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_OneContainerZeroInitContainer_ShouldPatchedOne() {
//	// Setup
//	pod := createPodForTests([]corev1.Container{_containers[0]}, nil)
//	resource := createWorkloadResourceForTests([]admisionrequest.Container{_containersAdmision[0]},nil)
//	req := createRequestForTests(pod)
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo}
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expectedInfo, nil).Once()
//
//	expected := []*contracts.ContainerVulnerabilityScanInfo{
//		_firstContainerVulnerabilityScanInfo,
//	}
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//
//	// Test
//	suite.Equal(admission.Allowed(string(_patchedReason)).AdmissionResponse, resp.AdmissionResponse)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//
//	suite.checkPatch(expected, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_TwoContainerZeroInitContainer_ShouldPatchedTwo() {
//	// Setup
//	pod := createPodForTests([]corev1.Container{_containers[0], _containers[1]}, nil)
//	resource := createWorkloadResourceForTests([]admisionrequest.Container{_containersAdmision[0], _containersAdmision[1]}, nil)
//	req := createRequestForTests(pod)
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo, _secondContainerVulnerabilityScanInfo}
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expectedInfo, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//	suite.Equal(admission.Allowed(string(_patchedReason)).AdmissionResponse, resp.AdmissionResponse)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//
//	suite.checkPatch(expectedInfo, patch)
//}
//
//func (suite *TestSuite) Test_Handle_ZeroContainerOneInitContainer_ShouldPatchedOne() {
//	// Setup
//	pod := createPodForTests(nil, []corev1.Container{_containers[0]})
//	resource := createWorkloadResourceForTests(nil, []admisionrequest.Container{_containersAdmision[0]})
//	req := createRequestForTests(pod)
//
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo}
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expectedInfo, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	//Act
//	resp := handler.Handle(context.Background(), *req)
//
//	//Test
//	suite.Equal(admission.Allowed(string(_patchedReason)).AdmissionResponse, resp.AdmissionResponse)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//	suite.checkPatch(expectedInfo, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_ZeroContainerTwoInitContainer_ShouldPatchedTwo() {
//	// Setup
//	pod := createPodForTests(nil, []corev1.Container{_containers[0], _containers[1]})
//	resource := createWorkloadResourceForTests(nil, []admisionrequest.Container{_containersAdmision[0], _containersAdmision[1]})
//	req := createRequestForTests(pod)
//
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo, _secondContainerVulnerabilityScanInfo}
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expectedInfo, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	//Act
//	resp := handler.Handle(context.Background(), *req)
//
//	//Test
//	suite.Equal(admission.Allowed(string(_patchedReason)).AdmissionResponse, resp.AdmissionResponse)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//	suite.checkPatch(expectedInfo, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//
//}
//
//func (suite *TestSuite) Test_Handle_OneContainerOneInitContainer_ShouldPatchedTwo() {
//	// Setup
//	pod := createPodForTests(nil, []corev1.Container{_containers[0], _containers[1]})
//	resource := createWorkloadResourceForTests(nil,[]admisionrequest.Container{_containersAdmision[0], _containersAdmision[1]})
//	req := createRequestForTests(pod)
//
//	expectedInfo := []*contracts.ContainerVulnerabilityScanInfo{_firstContainerVulnerabilityScanInfo, _secondContainerVulnerabilityScanInfo}
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(expectedInfo, nil).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//	suite.Equal(admission.Allowed(string(_patchedReason)).AdmissionResponse, resp.AdmissionResponse)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//	suite.checkPatch(expectedInfo, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_Error_AllowedTrueWithError_PodNoAnnotations() {
//	// Setup
//	pod := createPodForTests(nil, []corev1.Container{_containers[0]})
//	resource := createWorkloadResourceForTests(nil, []admisionrequest.Container{_containersAdmision[0]})
//	req := createRequestForTests(pod)
//
//	err := errors.New("MockError!!")
//
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(nil, err).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//
//
//
//	suite.Equal(int32(http.StatusInternalServerError), resp.Result.Code)
//	suite.NotEmpty(resp.Result.Message)
//	suite.Nil(resp.Patches)
//	// Super important
//	suite.True(resp.Allowed)
//	suite.Equal(0, len(resp.Patches))
//	suite.Nil( resp.Patches)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_Error_AllowedTrueWithError_PodWithAnnotations() {
//	// Setup
//	pod := createPodForTestsWithAnnotations(nil, []corev1.Container{_containers[0]})
//	resource := createWorkloadResourceForTestsWithAnnotations(nil, []admisionrequest.Container{_containersAdmision[0]})
//	req := createRequestForTests(pod)
//
//	err := errors.New("MockError!!")
//
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(nil, err).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//
//
//
//	suite.Equal(int32(http.StatusInternalServerError), resp.Result.Code)
//	suite.NotEmpty(resp.Result.Message)
//	suite.Nil(resp.Patches)
//	// Super important
//	suite.True(resp.Allowed)
//	suite.Equal(0, len(resp.Patches))
//	suite.Nil( resp.Patches)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_Error_AllowedTrueWithError_PodWithAzdAnnotations() {
//	// Setup
//	pod := createPodForTestsWithAzdAnnotations(nil, []corev1.Container{_containers[0]})
//	resource := createWorkloadResourceForTestsWithAzdAnnotations(nil, []admisionrequest.Container{_containersAdmision[0]})
//	req := createRequestForTests(pod)
//
//	err := errors.New("MockError!!")
//
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(nil, err).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//
//
//
//	suite.Equal(int32(http.StatusInternalServerError), resp.Result.Code)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//	suite.Equal(_expectedPatchForErrorEncountered1, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) Test_Handle_Error_AllowedTrueWithError_PodWithOnlyAzdAnnotations() {
//	// Setup
//	pod := createPodForTestsWithOnlyAzdAnnotations(nil, []corev1.Container{_containers[0]})
//	resource := createWorkloadResourceForTestsWithOnlyAzdAnnotations(nil, []admisionrequest.Container{_containersAdmision[0]})
//	req := createRequestForTests(pod)
//
//	err := errors.New("MockError!!")
//
//	suite.azdSecProviderMock.On("GetContainersVulnerabilityScanInfo", resource).Return(nil, err).Once()
//
//	handler := NewHandler(suite.azdSecProviderMock, &HandlerConfiguration{DryRun: false}, instrumentation.NewNoOpInstrumentationProvider())
//
//	// Act
//	resp := handler.Handle(context.Background(), *req)
//	// Test
//
//
//	suite.Equal(int32(http.StatusInternalServerError), resp.Result.Code)
//	suite.Equal(1, len(resp.Patches))
//	patch := resp.Patches[0]
//	suite.Equal(_expectedPatchForErrorEncountered2, patch)
//	suite.azdSecProviderMock.AssertExpectations(suite.T())
//}
//
//func (suite *TestSuite) checkPatch(expected []*contracts.ContainerVulnerabilityScanInfo, patch jsonpatch.JsonPatchOperation) {
//	// Verify the operation and the patch
//	suite.Equal(_expectedTestAddPatchOperation, patch.Operation)
//	suite.Equal(_expectedTestAnnotationPatchPath, patch.Path)
//
//	// Get data string
//	mapAnnotations, ok := patch.Value.(map[string]string)
//	suite.True(ok)
//
//	suite.Equal(1, len(mapAnnotations))
//	strValue, ok := mapAnnotations[contracts.ContainersVulnerabilityScanInfoAnnotationName]
//	suite.True(ok)
//
//	// Unmarshal
//	scanInfoList := new(contracts.ContainerVulnerabilityScanInfoList)
//	err := json.Unmarshal([]byte(strValue), scanInfoList)
//	suite.Nil(err)
//
//	// Verify timestamp
//	diff := time.Now().UTC().Sub(scanInfoList.GeneratedTimestamp)
//	suite.True((diff >= 0 && diff < time.Second))
//	suite.Equal(time.UTC, scanInfoList.GeneratedTimestamp.Location())
//
//	// Verify containers slice deep equal verification
//	suite.True(reflect.DeepEqual(expected, scanInfoList.Containers))
//}
//
//func TestCreateContainersVulnerabilityScanAnnotationPatchAdd(t *testing.T) {
//	suite.Run(t, new(TestSuite))
//}
//
//func createRequestForTests(pod *corev1.Pod) *admission.Request {
//	raw, err := json.Marshal(pod)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return &admission.Request{
//		AdmissionRequest: admissionv1.AdmissionRequest{
//			Name:      "podTest",
//			Namespace: "default",
//			Operation: admissionv1.Create,
//			Kind: metav1.GroupVersionKind{
//				Kind:    "Pod",
//				Group:   "",
//				Version: "v1",
//			},
//			Object: runtime.RawExtension{
//				Raw: raw,
//			},
//		},
//	}
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
//
//func createWorkloadResourceForTests(containers []admisionrequest.Container, initContainers []admisionrequest.Container) *admisionrequest.WorkloadResource {
//	return &admisionrequest.WorkloadResource{
//		Metadata: admisionrequest.ObjectMetadata{
//			Name:      "podTest",
//			Namespace: "default",
//		},
//		Spec: admisionrequest.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}
//
//func createPodForTestsWithAnnotations(containers []corev1.Container, initContainers []corev1.Container) *corev1.Pod {
//	return &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "podTest",
//			Namespace: "default",
//			Annotations: map[string]string{
//				_annotationTestKeyOne : _annotationTestValueOne,
//				_annotationTestKeyTwo : _annotationTestValueTwo,
//			},
//		},
//		TypeMeta: metav1.TypeMeta{},
//		Spec: corev1.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}
//
//func createWorkloadResourceForTestsWithAnnotations(containers []admisionrequest.Container, initContainers []admisionrequest.Container) *admisionrequest.WorkloadResource {
//	return &admisionrequest.WorkloadResource{
//		Metadata: admisionrequest.ObjectMetadata{
//			Name:      "podTest",
//			Namespace: "default",
//			Annotations: map[string]string{
//				_annotationTestKeyOne : _annotationTestValueOne,
//				_annotationTestKeyTwo : _annotationTestValueTwo,
//			},
//		},
//		Spec: admisionrequest.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}
//
//func createPodForTestsWithAzdAnnotations(containers []corev1.Container, initContainers []corev1.Container) *corev1.Pod {
//	return &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "podTest",
//			Namespace: "default",
//			Annotations: map[string]string{
//				_annotationTestKeyOne : _annotationTestValueOne,
//				_annotationTestKeyTwo : _annotationTestValueTwo,
//				contracts.ContainersVulnerabilityScanInfoAnnotationName: "some value",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{},
//		Spec: corev1.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}
//
//func createWorkloadResourceForTestsWithAzdAnnotations(containers []admisionrequest.Container, initContainers []admisionrequest.Container) *admisionrequest.WorkloadResource {
//	return &admisionrequest.WorkloadResource{
//		Metadata: admisionrequest.ObjectMetadata{
//			Name:      "podTest",
//			Namespace: "default",
//			Annotations: map[string]string{
//				_annotationTestKeyOne : _annotationTestValueOne,
//				_annotationTestKeyTwo : _annotationTestValueTwo,
//				contracts.ContainersVulnerabilityScanInfoAnnotationName: "some value",
//			},
//		},
//		Spec: admisionrequest.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}
//
//func createPodForTestsWithOnlyAzdAnnotations(containers []corev1.Container, initContainers []corev1.Container) *corev1.Pod {
//	return &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "podTest",
//			Namespace: "default",
//			Annotations: map[string]string{
//				contracts.ContainersVulnerabilityScanInfoAnnotationName: "some value",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{},
//		Spec: corev1.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}
//
//func createWorkloadResourceForTestsWithOnlyAzdAnnotations(containers []admisionrequest.Container, initContainers []admisionrequest.Container) *admisionrequest.WorkloadResource {
//	return &admisionrequest.WorkloadResource{
//		Metadata: admisionrequest.ObjectMetadata{
//			Name:      "podTest",
//			Namespace: "default",
//			Annotations: map[string]string{
//				contracts.ContainersVulnerabilityScanInfoAnnotationName: "some value",
//			},
//		},
//		Spec: admisionrequest.PodSpec{
//			Containers:     containers,
//			InitContainers: initContainers,
//		},
//	}
//}
