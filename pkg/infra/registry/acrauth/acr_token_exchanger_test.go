package acrauth

import (
	"bytes"
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/httpclient/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
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
const _exchanger_refreshTokenMock = "ACRRefreshTokenMock-Exchange"
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

	// TODO Add tests that use retrypolicy mock!
	retryPolicy, err := retrypolicy.NewRetryPolicy(_instrumentationP, &retrypolicy.RetryPolicyConfiguration{RetryAttempts: 1, RetryDuration: "3ms"})
	suite.Nil(err)
	_exchanger = NewACRTokenExchanger(_instrumentationP, _httpClientMock, retryPolicy)
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_Success() {
	expectedResponse := suite.generateTokenResponse(http.StatusOK, suite.generateTokenResponseBody(_exchanger_refreshTokenMock))
	_httpClientMock.On("Do", mock.MatchedBy(suite.isRequestExpected)).Return(expectedResponse, nil).Once()
	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)

	suite.Nil(err)
	suite.Equal(_exchanger_refreshTokenMock, refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_ErrorInHttp_ErrorPropagated() {
	expectedErr := errors.New("HttpErrorMock")
	_httpClientMock.On("Do", mock.MatchedBy(suite.isRequestExpected)).Return(nil, expectedErr).Once()
	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)

	suite.ErrorIs(err, expectedErr)
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_ErrorCodeInHttpWithBody_ErrorPropagated() {
	expectedResponse := suite.generateTokenResponse(http.StatusUnauthorized, "MockError")
	_httpClientMock.On("Do", mock.MatchedBy(suite.isRequestExpected)).Return(expectedResponse, nil).Once()
	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)

	suite.Error(err)
	suite.True(strings.Contains(err.Error(), "401") && strings.Contains(err.Error(), "MockError"))
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_ErrorCodeInHttpWithNilBody_ErrorPropagated() {
	expectedResponse := &http.Response{StatusCode: http.StatusUnauthorized}
	_httpClientMock.On("Do", mock.MatchedBy(suite.isRequestExpected)).Return(expectedResponse, nil).Once()
	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)

	suite.Error(err)
	suite.True(strings.Contains(err.Error(), "401"))
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_OkCodeNilBody_ErrorPropagated() {
	expectedResponse := &http.Response{StatusCode: http.StatusOK}
	_httpClientMock.On("Do", mock.MatchedBy(suite.isRequestExpected)).Return(expectedResponse, nil).Once()
	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)

	suite.Error(err)
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_OkCodeBadFormatBody_ErrorPropagated() {
	expectedResponse := suite.generateTokenResponse(http.StatusOK, "MockBadBody!")
	_httpClientMock.On("Do", mock.MatchedBy(suite.isRequestExpected)).Return(expectedResponse, nil).Once()
	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)

	suite.Error(err)
	cause := errors.Cause(err)
	a := &json.SyntaxError{}
	suite.ErrorAs(cause, &a)
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_OkCodeBadEmptyRefreshCode_ErrorPropagated() {
	expectedResponse := suite.generateTokenResponse(http.StatusOK, "{}")
	_httpClientMock.On("Do", mock.MatchedBy(suite.isRequestExpected)).Return(expectedResponse, nil).Once()
	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, _exchanger_armTokenMock)

	suite.Error(err)
	suite.ErrorIs(err, _refreshTokenEmptyError)
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_EmptyRegistry_Error() {

	refresh_token, err := _exchanger.ExchangeACRAccessToken("", _exchanger_armTokenMock)

	suite.Error(err)
	suite.ErrorIs(err, utils.NilArgumentError)
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) Test_ExchangeACRAccessToken_EmptyARMToken_Error() {

	refresh_token, err := _exchanger.ExchangeACRAccessToken(_excahnger_registryMock, "")

	suite.Error(err)
	suite.ErrorIs(err, utils.NilArgumentError)
	suite.Equal("", refresh_token)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenExchanger) AssertExpectations() {
	_httpClientMock.AssertExpectations(suite.T())
}

func (*TestSuiteTokenExchanger) generateTokenResponse(httpStatus int, body string) *http.Response {
	return &http.Response{
		StatusCode: httpStatus,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func (*TestSuiteTokenExchanger) generateTokenResponseBody(refreshToken string) string {
	return `{"refresh_token":"` + refreshToken + `"}`
}

func (*TestSuiteTokenExchanger) isRequestExpected(req *http.Request) bool {
	//TODO this method is not working well... we should change it once we will add tests for the retry policy.
	// 	http.Request has URL field that it's pointer so this method checks if the ref is eqaul and not the value!
	// 	probably should use somehow recursive function that checks all the inner fields that are pointer.
	var buffer = &bytes.Buffer{}
	err := req.Write(buffer)
	str := buffer.String()
	return err == nil && _exchanger_httpreq_str == str
}

func Test_Suite_TokenExchanger(t *testing.T) {
	suite.Run(t, new(TestSuiteTokenExchanger))
}
