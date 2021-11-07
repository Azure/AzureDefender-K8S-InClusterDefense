package acrauth

import (
	"context"
	authmocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/acrauth/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestSuiteTokenProvider struct {
	suite.Suite
}

var _providerNoCacheFunctionality *ACRTokenProvider
var _providerWithCacheFunctionality *ACRTokenProvider
var _provider_exchangerMock *mocks.IACRTokenExchanger
var _provider_azureTokenProviderMock *authmocks.IBearerAuthorizerTokenProvider
var _provider_cacheClientMock *cachemock.ICacheClient
var _provider_cacheClientInMemBasedMock cache.ICacheClient
var _expirationTime = time.Duration(1)


const _provider_armToken = "ARMTokenMock.."
const _provider_refreshToken = "ACRRefreshTokenMock.."
const _provider_registry = "tomerw.azurecr.io"

func (suite *TestSuiteTokenProvider) SetupTest() {
	instrumentationProvider := instrumentation.NewNoOpInstrumentationProvider()
	_provider_exchangerMock = &mocks.IACRTokenExchanger{}
	_provider_azureTokenProviderMock = &authmocks.IBearerAuthorizerTokenProvider{}
	_provider_cacheClientMock = &cachemock.ICacheClient{}
	_provider_cacheClientInMemBasedMock = cachemock.NewICacheInMemBasedMock()
	_providerNoCacheFunctionality = NewACRTokenProvider(instrumentationProvider, _provider_exchangerMock, _provider_azureTokenProviderMock, _provider_cacheClientMock, new(ACRTokenProviderConfiguration))
	_providerWithCacheFunctionality = NewACRTokenProvider(instrumentationProvider, _provider_exchangerMock, _provider_azureTokenProviderMock, _provider_cacheClientInMemBasedMock, &ACRTokenProviderConfiguration{
		cacheExpirationTime: _expirationTime,
	})
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_Success() {
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return(_provider_armToken, nil).Once()
	_provider_exchangerMock.On("ExchangeACRAccessToken", _provider_registry, _provider_armToken).Return(_provider_refreshToken, nil).Once()
	_provider_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_provider_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)


	val, err := _providerNoCacheFunctionality.GetACRRefreshToken(_provider_registry)

	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_FailOnTokenGet_Error() {
	expectedError := errors.New("azureTokenProviderMockError")
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return("", expectedError).Once()
	_provider_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_provider_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)

	val, err := _providerNoCacheFunctionality.GetACRRefreshToken(_provider_registry)

	suite.Equal("", val)
	suite.ErrorIs(err, expectedError)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_FailToExchange_Error() {
	expectedError := errors.New("exchangerMockError")
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return(_provider_armToken, nil).Once()
	_provider_exchangerMock.On("ExchangeACRAccessToken", _provider_registry, _provider_armToken).Return("", expectedError).Once()
	_provider_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_provider_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)

	val, err := _providerNoCacheFunctionality.GetACRRefreshToken(_provider_registry)

	suite.Equal("", val)
	suite.ErrorIs(err, expectedError)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_Success_NoKeyInCache() {
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return(_provider_armToken, nil).Once()
	_provider_exchangerMock.On("ExchangeACRAccessToken", _provider_registry, _provider_armToken).Return(_provider_refreshToken, nil).Once()
	_, err := _provider_cacheClientInMemBasedMock.Get(_provider_registry)
	suite.NotNil(err)
	val, err := _providerWithCacheFunctionality.GetACRRefreshToken(_provider_registry)
	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	val, err = _provider_cacheClientInMemBasedMock.Get(_provider_registry)
	suite.Nil(err)
	suite.Equal(_provider_refreshToken, val)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_Success_KeyInCache() {
	_ = _provider_cacheClientInMemBasedMock.Set(_provider_registry, _provider_refreshToken, _expirationTime)
	_, err := _provider_cacheClientInMemBasedMock.Get(_provider_registry)
	suite.Nil(err)
	val, err := _providerWithCacheFunctionality.GetACRRefreshToken(_provider_registry)
	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_Success_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime_ScannedResults() {
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return(_provider_armToken, nil).Once()
	_provider_exchangerMock.On("ExchangeACRAccessToken", _provider_registry, _provider_armToken).Return(_provider_refreshToken, nil).Once()
	_, err := _provider_cacheClientInMemBasedMock.Get(_provider_registry)
	suite.NotNil(err)
	val, err := _providerWithCacheFunctionality.GetACRRefreshToken(_provider_registry)
	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	val, err = _provider_cacheClientInMemBasedMock.Get(_provider_registry)
	suite.Nil(err)
	suite.Equal(_provider_refreshToken, val)
	val, err = _providerWithCacheFunctionality.GetACRRefreshToken(_provider_registry)
	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_Success_NoKeyInCache_SetKey_GetKeySecondTryAfterExpirationTime_ScannedResults() {
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return(_provider_armToken, nil).Twice()
	_provider_exchangerMock.On("ExchangeACRAccessToken", _provider_registry, _provider_armToken).Return(_provider_refreshToken, nil).Twice()
	_, err := _provider_cacheClientInMemBasedMock.Get(_provider_registry)
	suite.NotNil(err)
	val, err := _providerWithCacheFunctionality.GetACRRefreshToken(_provider_registry)
	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	val, err = _provider_cacheClientInMemBasedMock.Get(_provider_registry)
	suite.Nil(err)
	suite.Equal(_provider_refreshToken, val)
	time.Sleep(time.Second)
	val, err = _providerWithCacheFunctionality.GetACRRefreshToken(_provider_registry)
	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	suite.AssertExpectations()
}


func (suite *TestSuiteTokenProvider) AssertExpectations(){
	_provider_exchangerMock.AssertExpectations(suite.T())
	_provider_azureTokenProviderMock.AssertExpectations(suite.T())
}

func Test_Suite_TokenProvider(t *testing.T) {
	suite.Run(t, new(TestSuiteTokenProvider))
}
