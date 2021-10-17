package acrauth

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/httpclient/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/stretchr/testify/suite"
	"testing"
)

var _exchanger *ACRTokenExchanger
var _httpClientMock *mocks.IHttpClient

const _exchanger_armTokenMock = "ARMTokenMock-Exchange"
const _exchanger_refreshTokenMock = "ACRRefreshTokenMock-Excahne"
const _excahnger_registryMock = "tomerw.azurecr.io"


type TestSuiteTokenExchanger struct {
	suite.Suite
}


func (suite *TestSuiteTokenExchanger) SetupTest() {
	instrumentationP := instrumentation.NewNoOpInstrumentationProvider()
	_httpClientMock = &mocks.IHttpClient{}
	_exchanger = NewACRTokenExchanger(instrumentationP, _httpClientMock)
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_Success(){
	_httpClientMock.On("Do").Return()
}


func (suite *TestSuiteTokenExchanger) AssertExpectations(){
	_httpClientMock.AssertExpectations(suite.T())
}

func Test_Suite_TokenExchanger(t *testing.T) {
	suite.Run(t, new(TestSuiteTokenExchanger))
}