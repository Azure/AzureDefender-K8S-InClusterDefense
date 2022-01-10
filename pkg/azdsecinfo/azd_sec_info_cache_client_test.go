package azdsecinfo

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"testing"
	"time"
)

const (
	_podSpecCacheKeyTest1                   = "Test1"
	_timeoutKeyTest1                        = _timeoutPrefixForCacheKey + _podSpecCacheKeyTest1
	_containerVulnerabilityScanInfoKeyTest1 = _containerVulnerabilityScanInfoPrefixForCacheKey + _podSpecCacheKeyTest1
	_resultsStringTestWithError             = "{\"containerVulnerabilityScanInfo\":null,\"err\":\"NilArgumentError\"}"
	_expectedErrorString                    = "NilArgumentError"
)

var (
	_podSpecCacheKeyTest = "containerTest1:" + _imageOriginalTest1 + "," + "containerTest2:" + _imageOriginalTest2
	_expectedResultsWrapperTest1 = &containerVulnerabilityCacheResultsWrapper{ContainerVulnerabilityScanInfo: _expectedResultsTest1, ErrString: ""}
)

type AzdSecInfoProviderCacheClientTestSuite struct {
	suite.Suite
	cacheClientMock               *cachemock.ICacheClient
	azdSecInfoProviderCacheClient *AzdSecInfoProviderCacheClient
}

// This will run before each test in the suite
func (suite *AzdSecInfoProviderCacheClientTestSuite) SetupTest() {
	// Mock
	suite.cacheClientMock = new(cachemock.ICacheClient)
	suite.azdSecInfoProviderCacheClient = NewAzdSecInfoProviderCacheClient(instrumentation.NewNoOpInstrumentationProvider(), suite.cacheClientMock, &AzdSecInfoProviderConfiguration{CacheExpirationTimeTimeout: _cacheExpirationTimeTimeout, CacheExpirationContainerVulnerabilityScanInfo: _cacheExpirationTimeTimeout})

}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getContainerVulnerabilityScanInfofromCache_KeyNotFound() {
	suite.cacheClientMock.On("Get", _containerVulnerabilityScanInfoKeyTest1).Return("", new(cache.MissingKeyCacheError))
	containersVulnerabilityScanInfo, errorStoredInCache, err := suite.azdSecInfoProviderCacheClient.GetContainerVulnerabilityScanInfofromCache(_podSpecCacheKeyTest1)
	suite.Nil(containersVulnerabilityScanInfo)
	suite.Nil(errorStoredInCache)
	suite.IsTypef(cache.NewMissingKeyCacheError(""), err, "")
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getContainerVulnerabilityScanInfofromCache_UnmarshalError() {
	suite.cacheClientMock.On("Get", _containerVulnerabilityScanInfoKeyTest1).Return("", nil)
	containersVulnerabilityScanInfo, errorStoredInCache, err := suite.azdSecInfoProviderCacheClient.GetContainerVulnerabilityScanInfofromCache(_podSpecCacheKeyTest1)
	suite.Nil(containersVulnerabilityScanInfo)
	suite.Nil(errorStoredInCache)
	suite.NotNil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getContainerVulnerabilityScanInfofromCache_ErrorStoredInCache() {
	suite.cacheClientMock.On("Get", _containerVulnerabilityScanInfoKeyTest1).Return(_resultsStringTestWithError, nil)
	result, errorStoredInCache, err := suite.azdSecInfoProviderCacheClient.GetContainerVulnerabilityScanInfofromCache(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.Equal(errorStoredInCache.Error(), _expectedErrorString)
	suite.Nil(result)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getContainerVulnerabilityScanInfofromCache() {
	suite.cacheClientMock.On("Get", _containerVulnerabilityScanInfoKeyTest1).Return(_expectedResultsStringTestScanned, nil)
	result, errorStoredInCache, err := suite.azdSecInfoProviderCacheClient.GetContainerVulnerabilityScanInfofromCache(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.Nil(errorStoredInCache)
	suite.Equal(result, _expectedResultsTest1)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_setContainerVulnerabilityScanInfoInCache_SetGotError() {
	suite.cacheClientMock.On("Set", _containerVulnerabilityScanInfoKeyTest1, _expectedResultsStringTestScanned, mock.Anything).Return(utils.NilArgumentError)
	err := suite.azdSecInfoProviderCacheClient.SetContainerVulnerabilityScanInfoInCache(_podSpecCacheKeyTest1, _expectedResultsWrapperTest1.ContainerVulnerabilityScanInfo, nil)
	suite.NotNil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_setContainerVulnerabilityScanInfoInCache_SetErrorInCache() {
	suite.cacheClientMock.On("Set", _containerVulnerabilityScanInfoKeyTest1, _resultsStringTestWithError, mock.Anything).Return(nil)
	err := suite.azdSecInfoProviderCacheClient.SetContainerVulnerabilityScanInfoInCache(_podSpecCacheKeyTest1, nil, utils.NilArgumentError)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_setContainerVulnerabilityScanInfoInCache() {
	suite.cacheClientMock.On("Set", _containerVulnerabilityScanInfoKeyTest1, _expectedResultsStringTestScanned, mock.Anything).Return(nil)
	err := suite.azdSecInfoProviderCacheClient.SetContainerVulnerabilityScanInfoInCache(_podSpecCacheKeyTest1, _expectedResultsWrapperTest1.ContainerVulnerabilityScanInfo, nil)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getTimeOutStatus_KeyNotFound() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("", new(cache.MissingKeyCacheError))
	timeoutStatus, err := suite.azdSecInfoProviderCacheClient.GetTimeOutStatus(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.Equal(timeoutStatus, _noTimeOutEncountered)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getTimeOutStatus_GotError() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("", utils.NilArgumentError)
	timeoutStatus, err := suite.azdSecInfoProviderCacheClient.GetTimeOutStatus(_podSpecCacheKeyTest1)
	suite.NotNil(err)
	suite.Equal(timeoutStatus, _unknownTimeOutStatus)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getTimeOutStatus_ValueNotValidInt() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("invalid int", nil)
	timeoutStatus, err := suite.azdSecInfoProviderCacheClient.GetTimeOutStatus(_podSpecCacheKeyTest1)
	suite.NotNil(err)
	suite.Equal(timeoutStatus, _unknownTimeOutStatus)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_getTimeOutStatus() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("1", nil)
	timeoutStatus, err := suite.azdSecInfoProviderCacheClient.GetTimeOutStatus(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.Equal(timeoutStatus, 1)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_setTimeOutStatusAfterEncounteredTimeout_TwoEncountered() {
	suite.cacheClientMock.On("Set", _timeoutKeyTest1, "2", mock.Anything).Return(nil)
	err := suite.azdSecInfoProviderCacheClient.SetTimeOutStatusAfterEncounteredTimeout(_podSpecCacheKeyTest1, 2)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_setTimeOutStatusAfterEncounteredTimeout_OneEncountered() {
	suite.cacheClientMock.On("Set", _timeoutKeyTest1, "1", mock.Anything).Return(nil)
	err := suite.azdSecInfoProviderCacheClient.SetTimeOutStatusAfterEncounteredTimeout(_podSpecCacheKeyTest1, 1)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_setTimeOutStatusAfterEncounteredTimeout_Error() {
	suite.cacheClientMock.On("Set", _timeoutKeyTest1, "1", mock.Anything).Return(utils.NilArgumentError)
	err := suite.azdSecInfoProviderCacheClient.SetTimeOutStatusAfterEncounteredTimeout(_podSpecCacheKeyTest1, 1)
	suite.NotNil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_resetTimeOutInCacheAfterGettingScanResults_KeyNotFound() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("", new(cache.MissingKeyCacheError))
	err := suite.azdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_resetTimeOutInCacheAfterGettingScanResults_GetError() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("", utils.NilArgumentError)
	err := suite.azdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults(_podSpecCacheKeyTest1)
	suite.NotNil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_resetTimeOutInCacheAfterGettingScanResults_GetInvalidValue() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("invalid value", nil)
	err := suite.azdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults(_podSpecCacheKeyTest1)
	suite.NotNil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_resetTimeOutInCacheAfterGettingScanResults_GotZero() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("0", nil)
	err := suite.azdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_resetTimeOutInCacheAfterGettingScanResults_GotOne() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("1", nil)
	suite.cacheClientMock.On("Set", _timeoutKeyTest1, "0", time.Duration(1)).Return(nil)
	err := suite.azdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_resetTimeOutInCacheAfterGettingScanResults_GotTwo() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("2", nil)
	suite.cacheClientMock.On("Set", _timeoutKeyTest1, "0", time.Duration(1)).Return(nil)
	err := suite.azdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults(_podSpecCacheKeyTest1)
	suite.Nil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_resetTimeOutInCacheAfterGettingScanResults_SetError() {
	suite.cacheClientMock.On("Get", _timeoutKeyTest1).Return("2", nil)
	suite.cacheClientMock.On("Set", _timeoutKeyTest1, "0", time.Duration(1)).Return(utils.NilArgumentError)
	err := suite.azdSecInfoProviderCacheClient.ResetTimeOutInCacheAfterGettingScanResults(_podSpecCacheKeyTest1)
	suite.NotNil(err)
	suite.cacheClientMock.AssertExpectations(suite.T())
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_GetPodSpecCacheKey_Containers() {
	containers := []corev1.Container{_containers[0], _containers[1]}
	pod := createPodForTests(containers, nil)
	result := suite.azdSecInfoProviderCacheClient.GetPodSpecCacheKey(&pod.Spec)
	suite.Equal(_podSpecCacheKeyTest, result)
}

func (suite *AzdSecInfoProviderCacheClientTestSuite) Test_GetPodSpecCacheKey_InitContainers() {
	containers := []corev1.Container{_containers[0], _containers[1]}
	pod := createPodForTests(nil, containers)
	result := suite.azdSecInfoProviderCacheClient.GetPodSpecCacheKey(&pod.Spec)
	suite.Equal(_podSpecCacheKeyTest, result)
}

func TestAzdSecInfoProviderCacheClient(t *testing.T) {
	suite.Run(t, new(AzdSecInfoProviderCacheClientTestSuite))
}
