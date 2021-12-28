package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type ARGDataProviderCacheClientTestSuite struct {
	suite.Suite
	cacheMock *cachemock.ICacheClient
	argDataProviderCacheClient *ARGDataProviderCacheClient
}

func (suite *ARGDataProviderCacheClientTestSuite) SetupTest() {
	suite.cacheMock = new(cachemock.ICacheClient)
	suite.argDataProviderCacheClient = NewARGDataProviderCacheClient(instrumentation.NewNoOpInstrumentationProvider(), suite.cacheMock,
		&ARGDataProviderConfiguration{
			CacheExpirationTimeScannedResults:   _expirationTimeScanned,
			CacheExpirationTimeUnscannedResults: _expirationTimeUnscanned,
		})
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_getResultsFromCache_GetMissingKey(){
	suite.cacheMock.On("Get", _digest).Return("", new(cache.MissingKeyCacheError)).Once()
	scanStatus, scanFindings, err := suite.argDataProviderCacheClient.GetResultsFromCache(_digest)
	suite.Equal("", string(scanStatus))
	suite.Nil(scanFindings)
	suite.NotNil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_getResultsFromCache_GetError(){
	suite.cacheMock.On("Get", _digest).Return("", utils.NilArgumentError).Once()
	scanStatus, scanFindings, err := suite.argDataProviderCacheClient.GetResultsFromCache(_digest)
	suite.Equal("", string(scanStatus))
	suite.Nil(scanFindings)
	suite.NotNil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_getResultsFromCache_GetInvalidString(){
	suite.cacheMock.On("Get", _digest).Return("", nil).Once()
	scanStatus, scanFindings, err := suite.argDataProviderCacheClient.GetResultsFromCache(_digest)
	suite.Equal("", string(scanStatus))
	suite.Nil(scanFindings)
	suite.NotNil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_getResultsFromCache(){
	suite.cacheMock.On("Get", _digest).Return(_setToCacheTest1, nil).Once()
	scanStatus, scanFindings, err := suite.argDataProviderCacheClient.GetResultsFromCache(_digest)
	suite.Equal(contracts.UnhealthyScan, scanStatus)
	suite.Equal(expected_results, scanFindings)
	suite.Nil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_setScanFindingsInCache_SetError(){
	suite.cacheMock.On("Set", _digest, _setToCacheTest1, mock.Anything).Return(utils.NilArgumentError).Once()
	err := suite.argDataProviderCacheClient.SetScanFindingsInCache(expected_results, contracts.UnhealthyScan, _digest)
	suite.NotNil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_setScanFindingsInCache_SetUnscanned(){
	suite.cacheMock.On("Set", _digest, _setToCacheTest2, _expirationTimeUnscanned * time.Minute).Return(nil).Once()
	err := suite.argDataProviderCacheClient.SetScanFindingsInCache(nil, contracts.Unscanned, _digest)
	suite.Nil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_setScanFindingsInCache_SetUnscannedArrayWithNil(){
	suite.cacheMock.On("Set", _digest, _setToCacheTest2, _expirationTimeUnscanned * time.Minute).Return(nil).Once()
	err := suite.argDataProviderCacheClient.SetScanFindingsInCache([]*contracts.ScanFinding(nil), contracts.Unscanned, _digest)
	suite.Nil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}

func (suite *ARGDataProviderCacheClientTestSuite) Test_setScanFindingsInCache_SetScanned(){
	suite.cacheMock.On("Set", _digest, _setToCacheTest1, _expirationTimeScanned * time.Hour).Return(nil).Once()
	err := suite.argDataProviderCacheClient.SetScanFindingsInCache(expected_results, contracts.UnhealthyScan, _digest)
	suite.Nil(err)
	suite.cacheMock.AssertExpectations(suite.T())
}


func Test_ARGDataProviderCacheClientTestSuite(t *testing.T) {
	suite.Run(t, new(ARGDataProviderCacheClientTestSuite))
}

