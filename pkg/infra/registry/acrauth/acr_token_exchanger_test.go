package acrauth

import (
	"bytes"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/httpclient/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

var _exchanger *ACRTokenExchanger
var _httpClientMock *mocks.IHttpClient
var _instrumentationP instrumentation.IInstrumentationProvider
var _exchanger_httpreq *http.Request
var _exchanger_httpreq_str string

const _exchanger_armTokenMock = "ARMTokenMock-Exchange"
const _exchanger_refreshTokenMock = "ACRRefreshTokenMock-Excahne"
const _excahnger_registryMock = "tomerw.azurecr.io"

type TestSuiteTokenExchanger struct {
	suite.Suite
}

func (suite *TestSuiteTokenExchanger) SetupSuite() {
	var err error
	_instrumentationP = instrumentation.NewNoOpInstrumentationProvider()
	parameters := url.Values{}
	parameters.Add("grant_type", "access_token")
	parameters.Add("service", _excahnger_registryMock)
	parameters.Add("access_token", _exchanger_armTokenMock)
	// Seems like tenantId is not required - if ever needed it should be added via:	//parameters.Add("tenant", tenantID) - maybe it is needed on cross tenant
	// Not adding it for now...

	_exchanger_httpreq, err = http.NewRequest("POST", "https://"+_excahnger_registryMock+"/oauth2/exchange", strings.NewReader(parameters.Encode()))
	suite.Nil(err)
	_exchanger_httpreq.Header.Add(_contentTypeHeaderName, _applicationUrlEncodedContentType)
	_exchanger_httpreq.Header.Add(_contentLengthHeaderName, strconv.Itoa(int(_exchanger_httpreq.ContentLength)))
	var buffer = &bytes.Buffer{}
	err = _exchanger_httpreq.Write(buffer)
	suite.Nil(err)
	_exchanger_httpreq_str = buffer.String()
}



func (suite *TestSuiteTokenExchanger) SetupTest() {
	_httpClientMock = &mocks.IHttpClient{}
	_exchanger = NewACRTokenExchanger(_instrumentationP, _httpClientMock)
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_Success() {
	_httpClientMock.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		var buffer = &bytes.Buffer{}
		err := req.Write(buffer)
		str := buffer.String()
		compare := _exchanger_httpreq_str == str
		return err == nil && compare
	})).Return(&http.Response{}, nil).Once()

	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)


	suite.Nil(err)
	suite.Equal(refresh_token, _exchanger_refreshTokenMock)

	suite.AssertExpectations()

}

func (suite *TestSuiteTokenExchanger) AssertExpectations() {
	_httpClientMock.AssertExpectations(suite.T())
}

func Test_Suite_TokenExchanger(t *testing.T) {
	suite.Run(t, new(TestSuiteTokenExchanger))
}

// access_token=ARMTokenMock-Exchange&grant_type=access_token&service=tomerw.azurecr.io
// H [0]: Content-Type -> application/x-www-form-urlencoded
// H[1] : Content-Length -> 84