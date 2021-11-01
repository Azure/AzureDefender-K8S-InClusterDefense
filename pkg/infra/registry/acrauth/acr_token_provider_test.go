package acrauth

import (
	"context"
	authmocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/mocks"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/acrauth/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuiteTokenProvider struct {
	suite.Suite
}

var _provider *ACRTokenProvider
var _provider_exchangerMock *mocks.IACRTokenExchanger
var _provider_azureTokenProviderMock *authmocks.IBearerAuthorizerTokenProvider
var _provider_cacheClientMock *cachemock.ICacheClient

const _provider_armToken = "ARMTokenMock.."
const _provider_refreshToken = "ACRRefreshTokenMock.."
const _provider_registry = "tomerw.azurecr.io"

func (suite *TestSuiteTokenProvider) SetupTest() {
	instrumentationProvider := instrumentation.NewNoOpInstrumentationProvider()
	_provider_exchangerMock = &mocks.IACRTokenExchanger{}
	_provider_azureTokenProviderMock = &authmocks.IBearerAuthorizerTokenProvider{}
	_provider_cacheClientMock = &cachemock.ICacheClient{}
	_provider = NewACRTokenProvider(instrumentationProvider, _provider_exchangerMock, _provider_azureTokenProviderMock, _provider_cacheClientMock)
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_Success() {
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return(_provider_armToken, nil).Once()
	_provider_exchangerMock.On("ExchangeACRAccessToken", _provider_registry, _provider_armToken).Return(_provider_refreshToken, nil).Once()
	_provider_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_provider_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)


	val, err := _provider.GetACRRefreshToken(_provider_registry)

	suite.Equal(_provider_refreshToken, val)
	suite.Nil(err)
	suite.AssertExpectations()
}

func (suite *TestSuiteTokenProvider) Test_GetACRRefreshToken_FailOnTokenGet_Error() {
	expectedError := errors.New("azureTokenProviderMockError")
	_provider_azureTokenProviderMock.On("GetOAuthToken", context.Background()).Return("", expectedError).Once()
	_provider_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_provider_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)

	val, err := _provider.GetACRRefreshToken(_provider_registry)

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

	val, err := _provider.GetACRRefreshToken(_provider_registry)

	suite.Equal("", val)
	suite.ErrorIs(err, expectedError)
	suite.AssertExpectations()
}


func (suite *TestSuiteTokenProvider) AssertExpectations(){
	_provider_exchangerMock.AssertExpectations(suite.T())
	_provider_azureTokenProviderMock.AssertExpectations(suite.T())
}

func Test_Suite_TokenProvider(t *testing.T) {
	suite.Run(t, new(TestSuiteTokenProvider))
}
