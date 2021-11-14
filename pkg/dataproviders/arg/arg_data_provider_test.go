package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/mocks"
	queriesmock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	registryerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)
const (
	_registryMock = "tomerw.azurecr.io"
	_repositoryMock = "sqlo"
	_digestMock = "xckjhcdjdjhdh"

	_expirationTimeScanned = "1m" // for scanned results - 1 minutes
	_expirationTimeUnscanned = "1s" // for unscanned results - 1 seconds
	_expirationTimeErrors = "1s" // for errors - 1 seconds
	_registry = "imagescane2eacrdev.azurecr.io"
	_repository = "pushunhealthyimage/vulnerables/cve-2014-6271"
	_digest = "sha256:bdac8529e22931c1d99bf4907e12df3c2df0214070635a0b076fb11e66409883"
	_setToCacheTest1 = "{\"ScanStatus\":\"unhealthyScan\",\"ScanFindings\":[{\"patchable\":true,\"id\":\"1\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"1\",\"severity\":\"High\"},{\"patchable\":false,\"id\":\"1\",\"severity\":\"High\"}]}"

)

var (
	_expirationTimeScannedParsed, _ = time.ParseDuration(_expirationTimeScanned)

	_results = []interface{}{
		map[string]string{
			"id": "123456",
			"registry": _registry,
			"repository": _repository,
			"digest": _digest,
			"scanStatus": "Unhealthy",
			"scanFindingSeverity": "High",
			"findingsIds": "1",
			"patchable": "true",
		},
	}

	_resultsTest2 = []interface{}{}
)

type ARGDataProviderTestSuite struct {
	suite.Suite
	provider       *ARGDataProvider
	argClientMock *mocks.IARGClient
	queryGeneratorMock *queriesmock.IARGQueryGenerator
	cacheMock cache.ICacheClient
}

func (suite *ARGDataProviderTestSuite) SetupTest() {
	suite.argClientMock = new(mocks.IARGClient)
	suite.queryGeneratorMock = new(queriesmock.IARGQueryGenerator)
	suite.cacheMock = cachemock.NewICacheInMemBasedMock()
	argDataProviderFactory := NewARGDataProviderFactory()
	suite.provider, _ = argDataProviderFactory.CreateARGDataProvider(instrumentation.NewNoOpInstrumentationProvider(), suite.argClientMock, suite.queryGeneratorMock, suite.cacheMock,
		&ARGDataProviderConfiguration{
			CacheExpirationTimeScannedResults:   _expirationTimeScanned,
			CacheExpirationTimeUnscannedResults: _expirationTimeUnscanned,
			CacheExpirationTimeForErrors: _expirationTimeErrors,
		})
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache(){
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery",mock.Anything).Once().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_results, nil)
	_, err := suite.cacheMock.Get(_digest)
	suite.NotNil(err)
	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings[0].Id, "1")
	_, err = suite.cacheMock.Get(_digest)
	suite.Nil(err)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_KeyInCache(){
	_ = suite.cacheMock.Set(_digest, _setToCacheTest1, _expirationTimeScannedParsed)
	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings[0].Id, "1")
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime_ScannedResults() {

	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_results, nil)
	_, err := suite.cacheMock.Get(_digest)
	suite.NotNil(err)
	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings[0].Id, "1")
	_, err = suite.cacheMock.Get(_digest)
	suite.Nil(err)
	scanStatus, scanFindings, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings[0].Id, "1")
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime_UncannedResults() {
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("Test1", nil) // Note we expect it to be called only once
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_resultsTest2, nil) // Note we expect it to be called only once
	_, err := suite.cacheMock.Get(_digest)
	suite.NotNil(err)
	scanStatus, _, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	_, err = suite.cacheMock.Get(_digest)
	suite.Nil(err)
	scanStatus, _, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryAfterExpirationTime_UncannedResults(){
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Twice().Return("Test1", nil) // Note we expect it to be called twice
	suite.argClientMock.On("QueryResources", "Test1").Twice().Return(_resultsTest2, nil) // Note we expect it to be called twice
	_, err := suite.cacheMock.Get(_digest)
	suite.NotNil(err)
	scanStatus, _, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	_, err = suite.cacheMock.Get(_digest)
	suite.Nil(err)
	time.Sleep(time.Second)
	scanStatus, _, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime_ErrorAsValue() {
	_expectedError := new(registryerrors.ImageIsNotFoundErr)
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("", _expectedError) // Note we expect it to be called only once
	_, err := suite.cacheMock.Get(_digest)
	suite.NotNil(err)
	_, _, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.NotNil(err)
	suite.IsType(_expectedError, errors.Cause(err))
	_, err = suite.cacheMock.Get(_digest)
	suite.Nil(err)
	_, _, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.NotNil(err)
	suite.IsType(_expectedError, errors.Cause(err))
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryAfterExpirationTime_ErrorAsValue() {
	_expectedError := new(registryerrors.ImageIsNotFoundErr)
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Twice().Return("", _expectedError) // Note we expect it to be called only once
	_, err := suite.cacheMock.Get(_digest)
	suite.NotNil(err)
	_, _, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.NotNil(err)
	suite.IsType(_expectedError, errors.Cause(err))
	_, err = suite.cacheMock.Get(_digest)
	suite.Nil(err)
	time.Sleep(time.Second)
	_, _, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.NotNil(err)
	suite.IsType(_expectedError, errors.Cause(err))
	suite.AssertExpectation()
}



func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults() {

	//	 TODO
	//status, findings , err := suite.provider.GetImageVulnerabilityScanResults(_registryMock, _repositoryMock, _digestMock)

}

func (suite *ARGDataProviderTestSuite) AssertExpectation() {
	suite.argClientMock.AssertExpectations(suite.T())
	suite.queryGeneratorMock.AssertExpectations(suite.T())
	suite.argClientMock.AssertExpectations(suite.T())
}
func Test_ARGDataProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ARGDataProviderTestSuite))
}
