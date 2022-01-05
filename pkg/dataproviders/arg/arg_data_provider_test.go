package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/mocks"
	queriesmock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

const (
	_registryMock   = "tomerw.azurecr.io"
	_repositoryMock = "sqlo"
	_digestMock     = "xckjhcdjdjhdh"

	_expirationTimeScanned   = 1 // for scanned results - 1 minutes
	_expirationTimeUnscanned = 2 // for unscanned results - 1 seconds
	_registry                = "imagescane2eacrdev.azurecr.io"
	_repository              = "pushunhealthyimage/vulnerables/cve-2014-6271"
	_digest                  = "sha256:bdac8529e22931c1d99bf4907e12df3c2df0214070635a0b076fb11e66409883"
	_setToCacheTest1         = "{\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"1\",\"severity\":\"High\"}]}"
	_setToCacheTest2         = "{\"scanStatus\":\"unscanned\",\"scanFindings\":null}"
)

var (
	_results = []interface{}{
		map[string]string{
			"id":                  "123456",
			"registry":            _registry,
			"repository":          _repository,
			"digest":              _digest,
			"scanStatus":          "Unhealthy",
			"scanFindingSeverity": "High",
			"findingsIds":         "1",
			"patchable":           "true",
		},
	}

	_resultsTest2 = []interface{}{}

	expected_results = []*contracts.ScanFinding{
		{
			Patchable: true,
			Id:        "1",
			Severity:  "High",
		},
	}

	configuration = &ARGDataProviderConfiguration{
		CacheExpirationTimeScannedResults:   _expirationTimeScanned,
		CacheExpirationTimeUnscannedResults: _expirationTimeUnscanned,
	}
)

type ARGDataProviderTestSuite struct {
	suite.Suite
	provider           *ARGDataProvider
	argClientMock      *mocks.IARGClient
	queryGeneratorMock *queriesmock.IARGQueryGenerator
	cacheMock          *mocks.IARGDataProviderCacheClient
}

func (suite *ARGDataProviderTestSuite) SetupTest() {
	suite.argClientMock = new(mocks.IARGClient)
	suite.queryGeneratorMock = new(queriesmock.IARGQueryGenerator)
	suite.cacheMock = new(mocks.IARGDataProviderCacheClient)
	suite.provider = NewARGDataProvider(instrumentation.NewNoOpInstrumentationProvider(), suite.argClientMock, suite.queryGeneratorMock, suite.cacheMock,
		&ARGDataProviderConfiguration{
			CacheExpirationTimeScannedResults:   _expirationTimeScanned,
			CacheExpirationTimeUnscannedResults: _expirationTimeUnscanned,
		})
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache() {
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.ScanStatus(""), nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheMock.On("SetScanFindingsInCache", expected_results, contracts.UnhealthyScan, _digest).Return(nil).Maybe()
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_results, nil)

	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings, expected_results)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_KeyInCache() {
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.UnhealthyScan, expected_results, nil).Once()

	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings, expected_results)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime_ScannedResults() {
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.ScanStatus(""), nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.UnhealthyScan, expected_results, nil).Once()
	suite.cacheMock.On("SetScanFindingsInCache", expected_results, contracts.UnhealthyScan, _digest).Return(nil).Maybe()
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_results, nil)

	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings, expected_results)
	scanStatus, scanFindings, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings, expected_results)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime_UncannedResults() {
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.ScanStatus(""), nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.Unscanned, nil, nil).Once()
	suite.cacheMock.On("SetScanFindingsInCache", []*contracts.ScanFinding(nil), contracts.Unscanned, _digest).Return(nil).Maybe()
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_resultsTest2, nil)

	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	suite.Nil(scanFindings)
	scanStatus, scanFindings, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	suite.Nil(scanFindings)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_NoKeyInCache_SetKey_GetKeySecondTryAfterExpirationTime_UncannedResults() {
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.ScanStatus(""), nil, new(cache.MissingKeyCacheError)).Twice()
	suite.cacheMock.On("SetScanFindingsInCache", []*contracts.ScanFinding(nil), contracts.Unscanned, _digest).Return(nil).Maybe()
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Twice().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Twice().Return(_resultsTest2, nil)

	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	suite.Nil(scanFindings)
	scanStatus, scanFindings, err = suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.Unscanned)
	suite.Nil(scanFindings)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_ErrGetFromCache() {
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.ScanStatus(""), nil, utils.NilArgumentError).Once()
	suite.cacheMock.On("SetScanFindingsInCache", expected_results, contracts.UnhealthyScan, _digest).Return(nil).Maybe()
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_results, nil)

	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings, expected_results)
	suite.AssertExpectation()
}

func (suite *ARGDataProviderTestSuite) Test_GetImageVulnerabilityScanResults_ErrSetToCache() {
	suite.cacheMock.On("GetResultsFromCache", _digest).Return(contracts.ScanStatus(""), nil, new(cache.MissingKeyCacheError)).Once()
	suite.cacheMock.On("SetScanFindingsInCache", expected_results, contracts.UnhealthyScan, _digest).Once().Return(utils.NilArgumentError)
	suite.queryGeneratorMock.On("GenerateImageVulnerabilityScanQuery", mock.Anything).Once().Return("Test1", nil)
	suite.argClientMock.On("QueryResources", "Test1").Once().Return(_results, nil)

	scanStatus, scanFindings, err := suite.provider.GetImageVulnerabilityScanResults(_registry, _repository, _digest)
	time.Sleep(time.Second)
	suite.Nil(err)
	suite.Equal(scanStatus, contracts.UnhealthyScan)
	suite.Equal(scanFindings, expected_results)
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
	suite.cacheMock.AssertExpectations(suite.T())
}
func Test_ARGDataProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ARGDataProviderTestSuite))
}
