package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/mocks"
	queriesmock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries/mocks"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/stretchr/testify/suite"
	"testing"
)

const _registryMock = "tomerw.azurecr.io"
const _repositoryMock = "sqlo"
const _digestMock = "xckjhcdjdjhdh"

type ARGDataProviderTestSuite struct {
	suite.Suite
	provider       *ARGDataProvider
	argClientMock *mocks.IARGClient
	queryGeneratorMock *queriesmock.IARGQueryGenerator
}

func (suite *ARGDataProviderTestSuite) SetupTest() {
	instrumentationP := instrumentation.NewNoOpInstrumentationProvider()
	suite.argClientMock = new(mocks.IARGClient)
	suite.queryGeneratorMock = new(queriesmock.IARGQueryGenerator)
	suite.provider = NewARGDataProvider(instrumentationP, suite.argClientMock, suite.queryGeneratorMock, new(cachemock.ICacheClient))
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
