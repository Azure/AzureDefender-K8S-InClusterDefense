package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/mocks"
	argDataProviderMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registryErrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest"
	tag2DigestResolverMocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/tag2digest/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

	_noTimeOutEncounteredTestString   = "0"
	_oneTimeOutEncounteredTestString  = "1"
	_twoTimesOutEncounteredTestString = "2"
	_unknownTimeOutStatusTestString   = "-1"
)

var (
	_imageRegistry                             = "playground.azurecr.io"
	_imageRepo                                 = "testrepo"
	_imageTagTest1                             = "1.0"
	_imageTagTest2                             = "2.0"
	_imageOriginalTest1                        = _imageRegistry + "/" + _imageRepo + ":" + _imageTagTest1
	_imageOriginalTest2                        = _imageRegistry + "/" + _imageRepo + ":" + _imageTagTest2
	_containerVulnerabilityScanInfoForCacheKey = _containerVulnerabilityScanInfoPrefixForCacheKey + _imageOriginalTest1
	_timeoutForCacheKey                        = _timeoutPrefixForCacheKey + _imageOriginalTest1

	// Test1
	_containers = []admisionrequest.Container{
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

	_scanStatus                     = contracts.UnhealthyScan
	_scanFindings                   = []*contracts.ScanFinding{{Patchable: true, Id: "1", Severity: "High"}}
	_containerVulnerabilityScanInfo = &contracts.ContainerVulnerabilityScanInfo{
		Name:         _containers[0].Name,
		Image:        &contracts.Image{Name: _imageOriginalTest1, Digest: _digestTest1},
		ScanStatus:   _scanStatus,
		ScanFindings: _scanFindings,
	}
	_expectedResultsTest1               = []*contracts.ContainerVulnerabilityScanInfo{_containerVulnerabilityScanInfo}
	_expectedResultsStringTestScanned   = "{\"containerVulnerabilityScanInfo\":[{\"name\":\"containerTest1\",\"image\":{\"name\":\"playground.azurecr.io/testrepo:1.0\",\"digest\":\"sha256:9f9ed5fe24766b31bcb64aabba73e96cc5b7c2da578f9cd2fca20846cf5d7557\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"1\",\"severity\":\"High\"}]}],\"err\":\"\"}"
	_expectedResultsStringTestUnscanned = "{\"containerVulnerabilityScanInfo\":[{\"name\":\"containerTest1\",\"image\":{\"name\":\"playground.azurecr.io/testrepo:1.0\",\"digest\":\"sha256:9f9ed5fe24766b31bcb64aabba73e96cc5b7c2da578f9cd2fca20846cf5d7557\"},\"scanStatus\":\"unscanned\",\"scanFindings\":null}],\"err\":\"\"}"

	// Test2
	info = &contracts.ContainerVulnerabilityScanInfo{
		Name: "containerTest1",
		Image: &contracts.Image{
			Name:   "playground.azurecr.io/testrepo:1.0",
			Digest: "",
		},
		ScanStatus:   contracts.Unscanned,
		ScanFindings: nil,
		AdditionalData: map[string]string{
			contracts.UnscannedReasonAnnotationKey: string(contracts.ImageDoesNotExistUnscannedReason),
		},
	}

	_expectedResultsTest2 = []*contracts.ContainerVulnerabilityScanInfo{info}

	// Test3
	info3 = &contracts.ContainerVulnerabilityScanInfo{
		Name: _containers[0].Name,
		Image: &contracts.Image{
			Name:   _containers[0].Image,
			Digest: "",
		},
		ScanStatus:   contracts.Unscanned,
		ScanFindings: nil,
		AdditionalData: map[string]string{
			contracts.UnscannedReasonAnnotationKey: string(contracts.GetContainersVulnerabilityScanInfoTimeoutUnscannedReason),
		},
	}
	_expectedResultsTest3 = []*contracts.ContainerVulnerabilityScanInfo{info3}
)

type AzdSecInfoProviderTestSuite struct {
	suite.Suite
	tag2DigestResolverMock *tag2DigestResolverMocks.ITag2DigestResolver
	argDataProviderMock    *argDataProviderMocks.IARGDataProvider
	azdSecInfoProvider     *AzdSecInfoProvider
	cacheClientMock        *mocks.IAzdSecInfoProviderCacheClient
}

// This will run before each test in the suite
func (suite *AzdSecInfoProviderTestSuite) SetupTest() {
	// Mock
	suite.tag2DigestResolverMock = &tag2DigestResolverMocks.ITag2DigestResolver{}
	suite.argDataProviderMock = &argDataProviderMocks.IARGDataProvider{}
	suite.cacheClientMock = new(mocks.IAzdSecInfoProviderCacheClient)
	suite.azdSecInfoProvider = NewAzdSecInfoProvider(instrumentation.NewNoOpInstrumentationProvider(), suite.argDataProviderMock, suite.tag2DigestResolverMock, &utils.TimeoutConfiguration{TimeDurationInMS: _TimeDurationGetContainersVulnerabilityScanInfo}, suite.cacheClientMock)
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_NoResultsInCache_ScannedResults() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheClientMock.On("ResetTimeOutInCacheAfterGettingScanResults", _imageOriginalTest1).Return(nil).Maybe()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", _imageOriginalTest1, _expectedResultsTest1, nil).Return(nil).Maybe()

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(_scanStatus, _scanFindings, nil)

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Equal(res, _expectedResultsTest1)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_NoResultsInCache_UnscannedResults() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheClientMock.On("ResetTimeOutInCacheAfterGettingScanResults", _imageOriginalTest1).Return(nil).Maybe()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", _imageOriginalTest1, _expectedResultsTest2, nil).Return(nil).Maybe()

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(contracts.Unscanned, nil, registryErrors.NewImageIsNotFoundErr("", errors.New("")))

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Equal(res[0].ScanStatus, contracts.Unscanned)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_ResultsInCache_ScannedResults() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(_expectedResultsTest1, nil, nil).Once()

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Equal(res, _expectedResultsTest1)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_ResultsInCache_ErrorAsResult() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, errors.New(""), nil).Once()

	// Act
	res, err := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.NotNil(err)
	suite.Nil(res)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_NoResultsInCache_ScannedResults_NoTimeOut_ErrorFromCache() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, nil, errors.New("")).Once()
	suite.cacheClientMock.On("ResetTimeOutInCacheAfterGettingScanResults", _imageOriginalTest1).Return(nil).Maybe()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", _imageOriginalTest1, _expectedResultsTest1, nil).Return(nil).Maybe()

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(_scanStatus, _scanFindings, nil)

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Equal(res, _expectedResultsTest1)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_EncounteredTimeout_FirstTimeout() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheClientMock.On("GetTimeOutStatus", _imageOriginalTest1).Return(0, nil).Once()
	suite.cacheClientMock.On("SetTimeOutStatusAfterEncounteredTimeout", _imageOriginalTest1, 1).Return(nil).Once()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", _imageOriginalTest1, _expectedResultsTest1, nil).Return(nil).Maybe()

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(_scanStatus, _scanFindings, nil).Run(func(args mock.Arguments) {
		time.Sleep(_defaultTimeDurationGetContainersVulnerabilityScanInfo)
	})

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Equal(res, _expectedResultsTest3)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_EncounteredTimeout_SecondTimeout() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheClientMock.On("GetTimeOutStatus", _imageOriginalTest1).Return(1, nil).Once()
	suite.cacheClientMock.On("SetTimeOutStatusAfterEncounteredTimeout", _imageOriginalTest1, 2).Return(nil).Once()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", _imageOriginalTest1, _expectedResultsTest1, nil).Return(nil).Maybe()

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(_scanStatus, _scanFindings, nil).Run(func(args mock.Arguments) {
		time.Sleep(_defaultTimeDurationGetContainersVulnerabilityScanInfo)
	})

	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Equal(res, _expectedResultsTest3)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_EncounteredTimeout_ThirdTimeout() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheClientMock.On("GetTimeOutStatus", _imageOriginalTest1).Return(2, nil).Once()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", _imageOriginalTest1, _expectedResultsTest1, nil).Return(nil).Maybe()

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(_scanStatus, _scanFindings, nil).Run(func(args mock.Arguments) {
		time.Sleep(_defaultTimeDurationGetContainersVulnerabilityScanInfo * time.Second)
	})

	// Act
	res, err := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Nil(res)
	suite.NotNil(err)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_getContainersVulnerabilityScanInfo_EncounteredTimeout_ErrorFromCache() {
	containers := []*admisionrequest.Container{&_containers[0]}
	workloadResource := createWorkloadResourceForTests(containers, nil)

	suite.cacheClientMock.On("GetPodSpecCacheKey", workloadResource.Spec).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", _imageOriginalTest1).Return(nil, nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheClientMock.On("GetTimeOutStatus", _imageOriginalTest1).Return(-1, errors.New("")).Once()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", _imageOriginalTest1, _expectedResultsTest1, nil).Return(nil).Maybe()

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once()
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(_scanStatus, _scanFindings, nil).Run(func(args mock.Arguments) {
		time.Sleep(_defaultTimeDurationGetContainersVulnerabilityScanInfo * time.Second)
	})

	// Act
	res, err := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Nil(res)
	suite.NotNil(err)
	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_AllContainersNil() {

	suite.cacheClientMock.On("GetPodSpecCacheKey", mock.Anything).Return(_imageOriginalTest1).Once()
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", mock.Anything).Return(nil, nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheClientMock.On("ResetTimeOutInCacheAfterGettingScanResults", mock.Anything).Return(nil).Maybe()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	workloadResource := createWorkloadResourceForTests(nil, nil)
	// Act
	res, err := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	suite.NotNil(res)
	suite.Nil(err)
	suite.Len(res, 0)
}

func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_InitContainersNil() {
	suite.cacheClientMock.On("GetPodSpecCacheKey", mock.Anything).Return(_imageOriginalTest1)
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", mock.Anything).Return(nil, nil, new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("ResetTimeOutInCacheAfterGettingScanResults", mock.Anything).Return(nil).Maybe()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.goroutineTest(suite.getContainersVulnerabilityScanInfoTest_InitContainersNil)
}

func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_ContainersNil() {
	suite.cacheClientMock.On("GetPodSpecCacheKey", mock.Anything).Return(_imageOriginalTest1)
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", mock.Anything).Return(nil, nil, new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("ResetTimeOutInCacheAfterGettingScanResults", mock.Anything).Return(nil).Maybe()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.goroutineTest(suite.getContainersVulnerabilityScanInfoTest_ContainersNil)
}

func (suite *AzdSecInfoProviderTestSuite) Test_GetContainersVulnerabilityScanInfo_Run_In_Parallel_OneContainerOneInitContainer() {
	suite.cacheClientMock.On("GetPodSpecCacheKey", mock.Anything).Return(_imageOriginalTest1)
	suite.cacheClientMock.On("GetContainerVulnerabilityScanInfofromCache", mock.Anything).Return(nil, nil, new(cache.MissingKeyCacheError))
	suite.cacheClientMock.On("ResetTimeOutInCacheAfterGettingScanResults", mock.Anything).Return(nil).Maybe()
	suite.cacheClientMock.On("SetContainerVulnerabilityScanInfoInCache", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
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
	containers := []*admisionrequest.Container{&_containers[0], &_containers[1]}
	workloadResource := createWorkloadResourceForTests(containers, nil)
	suite.getContainersVulnerabilityScanInfoTest(workloadResource, waitFirstContainer, waitSecondContainer)
}

func (suite *AzdSecInfoProviderTestSuite) getContainersVulnerabilityScanInfoTest_ContainersNil(waitFirstContainer time.Duration, waitSecondContainer time.Duration) {
	containers := []*admisionrequest.Container{&_containers[0], &_containers[1]}
	workloadResource := createWorkloadResourceForTests(nil, containers)
	suite.getContainersVulnerabilityScanInfoTest(workloadResource, waitFirstContainer, waitSecondContainer)
}

func (suite *AzdSecInfoProviderTestSuite) getContainersVulnerabilityScanInfoTest_OneContainerOneInitContainer(waitFirstContainer time.Duration, waitSecondContainer time.Duration) {
	containers := []*admisionrequest.Container{&_containers[0]}
	initContainers := []*admisionrequest.Container{&_containers[1]}
	pod := createWorkloadResourceForTests(containers, initContainers)
	suite.getContainersVulnerabilityScanInfoTest(pod, waitFirstContainer, waitSecondContainer)
}

func (suite *AzdSecInfoProviderTestSuite) getContainersVulnerabilityScanInfoTest(workloadResource *admisionrequest.WorkloadResource, waitFirstContainer time.Duration, waitSecondContainer time.Duration) {

	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest1, _resourceCtxTest1).Return(_digestTest1, nil).Once().Run(func(args mock.Arguments) {
		time.Sleep(waitFirstContainer * time.Second)
	})
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest1.Registry(), _imageRedTest1.Repository(), _digestTest1).Once().Return(contracts.Unscanned, nil, nil)
	suite.tag2DigestResolverMock.On("Resolve", _imageRedTest2, _resourceCtxTest2).Return(_digestTest2, nil).Once().Run(func(args mock.Arguments) {
		time.Sleep(waitSecondContainer * time.Second)
	})
	suite.argDataProviderMock.On("GetImageVulnerabilityScanResults", _imageRedTest2.Registry(), _imageRedTest2.Repository(), _digestTest2).Once().Return(contracts.Unscanned, nil, nil)
	// Act
	res, _ := suite.azdSecInfoProvider.GetContainersVulnerabilityScanInfo(workloadResource)
	// Test
	suite.Equal(res[0].ScanStatus, contracts.Unscanned)
	suite.Equal(res[1].ScanStatus, contracts.Unscanned)

	suite.AssertExpectation()
}

func (suite *AzdSecInfoProviderTestSuite) AssertExpectation() {
	suite.argDataProviderMock.AssertExpectations(suite.T())
	suite.tag2DigestResolverMock.AssertExpectations(suite.T())
	suite.cacheClientMock.AssertExpectations(suite.T())
}

// measureTime runs the given function f and returns its run time
func measureTime(f func(time.Duration, time.Duration), waitFirstContainer time.Duration, waitSecondContainer time.Duration) time.Duration {
	start := time.Now()
	f(waitFirstContainer, waitSecondContainer)
	return time.Since(start)
}

func createWorkloadResourceForTests(containers []*admisionrequest.Container, initContainers []*admisionrequest.Container) *admisionrequest.WorkloadResource {
	return &admisionrequest.WorkloadResource{
		Metadata: &admisionrequest.ObjectMetadata{
			Name:      "podTest",
			Namespace: "default",
		},
		Spec: &admisionrequest.PodSpec{
			Containers:     containers,
			InitContainers: initContainers,
		},
	}
}

//TODO tests for azd_sec_info_provider GetContainersVulnerabilityScanInfo results (mock on tag2digest and argdataprovide)
