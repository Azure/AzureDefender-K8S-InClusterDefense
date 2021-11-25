package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	argDataProviderMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
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

const (
	_TimeDurationGetContainersVulnerabilityScanInfo = 0
	_waitTwoSeconds                                 = 2
	_waitZeroSeconds                                = 0
	_maxAllowedDifferenceBetweenRuns                = 0.3
	_cacheExpirationTimeTimeout                     = 1
)

var (
	_imageRegistry      = "playground.azurecr.io"
	_imageRepo          = "testrepo"
	_imageTagTest1      = "1.0"
	_imageTagTest2      = "2.0"
	_imageOriginalTest1 = _imageRegistry + "/" + _imageRepo + ":" + _imageTagTest1
	_imageOriginalTest2 = _imageRegistry + "/" + _imageRepo + ":" + _imageTagTest2
	_containerVulnerabilityScanInfoForCacheKey = "ContainerVulnerabilityScanInfo" + _imageOriginalTest1
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

	_imageRedTest1    = registry.NewTag(_imageOriginalTest1, _imageRegistry, _imageRepo, _imageTagTest1)
	_resourceCtxTest1 = tag2digest.NewResourceContext("default", []string{}, "")
	_digestTest1      = "sha256:9f9ed5fe24766b31bcb64aabba73e96cc5b7c2da578f9cd2fca20846cf5d7557"

	_imageRedTest2    = registry.NewTag(_imageOriginalTest2, _imageRegistry, _imageRepo, _imageTagTest2)
	_resourceCtxTest2 = tag2digest.NewResourceContext("default", []string{}, "")
	_digestTest2      = "sha256:86a80e680602c613519a5af190219346230a3b02d98606727b9c8d47d8dc88ed"

	_scanStatus = contracts.UnhealthyScan
	_scanFindings = []*contracts.ScanFinding{{Patchable: true, Id: "1", Severity: "High"}}
	_containerVulnerabilityScanInfo = &contracts.ContainerVulnerabilityScanInfo{
		Name: _containers[0].Name,
		Image: &contracts.Image{Name: _imageOriginalTest1, Digest: _digestTest1},
		ScanStatus: _scanStatus,
		ScanFindings: _scanFindings,
	}
	_expectedResults = []*contracts.ContainerVulnerabilityScanInfo{_containerVulnerabilityScanInfo}
	_expectedResultsString = "{\"ContainerVulnerabilityScanInfo\":[{\"name\":\"containerTest1\",\"image\":{\"name\":\"playground.azurecr.io\\/testrepo:1.0\",\"digest\":\"sha256:9f9ed5fe24766b31bcb64aabba73e96cc5b7c2da578f9cd2fca20846cf5d7557\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"1\",\"severity\":\"High\"}]}],\"Err\":null}"

)

type AzdSecInfoProviderTestSuite struct {
	suite.Suite
	tag2DigestResolverMock *tag2DigestResolverMocks.ITag2DigestResolver
	argDataProviderMock    *argDataProviderMocks.IARGDataProvider
	azdSecInfoProvider     *AzdSecInfoProvider
	cacheClientMock *cachemock.ICacheClient
}

// This will run before each test in the suite
func (suite *AzdSecInfoProviderTestSuite) SetupTest() {
	// Mock
	suite.tag2DigestResolverMock = &tag2DigestResolverMocks.ITag2DigestResolver{}
	suite.argDataProviderMock = &argDataProviderMocks.IARGDataProvider{}
	suite.cacheClientMock = new(cachemock.ICacheClient)
	suite.azdSecInfoProvider = NewAzdSecInfoProvider(instrumentation.NewNoOpInstrumentationProvider(), suite.argDataProviderMock, suite.tag2DigestResolverMock, &utils.TimeoutConfiguration{TimeDurationInMS: _TimeDurationGetContainersVulnerabilityScanInfo}, suite.cacheClientMock, &AzdSecInfoProviderConfiguration{CacheExpirationTimeTimeout: _cacheExpirationTimeTimeout, CacheExpirationContainerVulnerabilityScanInfo: _cacheExpirationTimeTimeout})
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_NoResultsInCache_ScannedResults() {
	suite.cacheClientMock.On("Get", mock.Anything).Return("", new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	containers := []corev1.Container{_containers[0]}
	pod := createPodForTests(containers, nil)
	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(_scanStatus, _scanFindings, nil)

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
	// Test
	suite.Equal(res, _expectedResults)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_NoResultsInCache_UnscannedResults() {
	suite.cacheClientMock.On("Get", mock.Anything).Return("", new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	containers := []corev1.Container{_containers[0]}
	pod := createPodForTests(containers, nil)
	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(contracts.Unscanned, nil, nil)

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
	// Test
	suite.Equal(res[0].ScanStatus, contracts.Unscanned)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_ResultsInCache_ScannedResults() {
	suite.cacheClientMock.On("Get", _containerVulnerabilityScanInfoForCacheKey).Return(_expectedResultsString, nil)
	suite.cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	containers := []corev1.Container{_containers[0]}
	pod := createPodForTests(containers, nil)

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
	// Test
	suite.Equal(res, _expectedResults)
	suite.AssertExpectation()
}


func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_AllContainersNil() {
	suite.cacheClientMock.On("Get", mock.Anything).Return("", new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	pod := createPodForTests(nil, nil)
	// Act
	res, err := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
	suite.NotNil(res)
	suite.Nil(err)
	suite.Len(res, 0)
}

func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_InitContainersNil() {
	suite.cacheClientMock.On("Get", mock.Anything).Return("", new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.goroutineTest(suite.getContainersVulnerabilityScanInfoTest_InitContainersNil)
}

func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_ContainersNil() {
	suite.cacheClientMock.On("Get", mock.Anything).Return("", new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.goroutineTest(suite.getContainersVulnerabilityScanInfoTest_ContainersNil)
}

func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_OneContainerOneInitContainer() {
	suite.cacheClientMock.On("Get", mock.Anything).Return("", new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.goroutineTest(suite.getContainersVulnerabilityScanInfoTest_OneContainerOneInitContainer)
}

func TestUpdateVulnSecInfoContainers(t *testing.T) {
	suite.Run(t, new(AzdSecInfoProviderTestSuite))
}

// goroutineTest running funcToRun twice.
// First - run with waiting time of _waitTwoSeconds seconds for each container.
// Second - run with waiting time of _waitTwoSeconds seconds for first container and _waitZeroSeconds for the second container.
// Afterwards, checks if the duration difference between the two runs is less than _maxAllowedDifferenceBetweenRuns which is significantly less than _waitTwoSeconds seconds.
// This way we confirm that the scan for each container is done in parallel. If the scans are not running in parallel the duration of the first run will be the sum of _waitTwoSeconds + _waitTwoSeconds seconds instead of _waitTwoSeconds.
func (suite *AzdSecInfoProviderTestSuite) goroutineTest(funcToRun func(time.Duration, time.Duration)) {
	// First run duration
	run1 := measureTime(funcToRun, _waitTwoSeconds, _waitTwoSeconds)
	// Second run duration
	run2 := measureTime(funcToRun, _waitTwoSeconds, _waitZeroSeconds)
	// Assert scans run in parallel
	suite.True(math.Abs(run1.Seconds()-run2.Seconds()) < _maxAllowedDifferenceBetweenRuns)
}

func (suite *AzdSecInfoProviderTestSuite) getContainersVulnerabilityScanInfoTest_InitContainersNil(waitFirstContainer time.Duration, waitSecondContainer time.Duration) {
	containers := []corev1.Container{_containers[0], _containers[1]}
	pod := createPodForTests(containers, nil)
	suite.getContainersVulnerabilityScanInfoTest(pod, waitFirstContainer, waitSecondContainer)
}

func (suite *AzdSecInfoProviderTestSuite) getContainersVulnerabilityScanInfoTest_ContainersNil(waitFirstContainer time.Duration, waitSecondContainer time.Duration) {
	containers := []corev1.Container{_containers[0], _containers[1]}
	pod := createPodForTests(nil, containers)
	suite.getContainersVulnerabilityScanInfoTest(pod, waitFirstContainer, waitSecondContainer)
}

func (suite *AzdSecInfoProviderTestSuite) getContainersVulnerabilityScanInfoTest_OneContainerOneInitContainer(waitFirstContainer time.Duration, waitSecondContainer time.Duration) {
	containers := []corev1.Container{_containers[0]}
	initContainers := []corev1.Container{_containers[1]}
	pod := createPodForTests(containers, initContainers)
	suite.getContainersVulnerabilityScanInfoTest(pod, waitFirstContainer, waitSecondContainer)
}

func (suite *AzdSecInfoProviderTestSuite) getContainersVulnerabilityScanInfoTest(pod *corev1.Pod, waitFirstContainer time.Duration, waitSecondContainer time.Duration) {

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once().Run(func(args mock.Arguments) {
		time.Sleep(waitFirstContainer * time.Second)
	})
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(contracts.Unscanned, nil, nil)
	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest2, _resourceCtxTest2).Return(_digestTest2, nil).Once().Run(func(args mock.Arguments) {
		time.Sleep(waitSecondContainer * time.Second)
	})
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest2.Registry(), _imageRedTest2.Repository(), _digestTest2).Once().Return(contracts.Unscanned, nil, nil)
	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(&pod.Spec, &pod.ObjectMeta, &pod.TypeMeta)
	// Test
	suite.Equal(res[0].ScanStatus, contracts.Unscanned)
	suite.Equal(res[1].ScanStatus, contracts.Unscanned)

	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) AssertExpectation() {
	suite.argDataProviderMock.AssertExpectations(suite.T())
	suite.tag2DigestResolverMock.AssertExpectations(suite.T())
}

// measureTime runs the given function f and returns its run time
func measureTime(f func(time.Duration, time.Duration), waitFirstContainer time.Duration, waitSecondContainer time.Duration) time.Duration {
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

//TODO tests for azd_sec_info_provider GetContainersVulnerabilityScanInfo results (mock on tag2digest and argdataprovide)
