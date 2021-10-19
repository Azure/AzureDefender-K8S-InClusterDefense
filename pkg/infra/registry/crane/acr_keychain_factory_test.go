package crane

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/acrauth/mocks"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"testing"
)

const _acrKcRegistryMock = "tomerw.devops.io"
const _acrKcRefreshTokenMock = "refreshTokenMock!"

type TestSuiteACRKCFactorySuite struct {
	suite.Suite
	factory      *ACRKeychainFactory
	acrTokenMock *mocks.IACRTokenProvider
}

func (suite *TestSuiteACRKCFactorySuite) SetupTest() {
	noopsIntrumentation := instrumentation.NewNoOpInstrumentationProvider()
	suite.acrTokenMock = new(mocks.IACRTokenProvider)
	suite.factory = NewACRKeychainFactory(noopsIntrumentation, suite.acrTokenMock)
}

func (suite *TestSuiteACRKCFactorySuite) Test_Create_Success() {

	expectedKC := &ACRKeyChain{Token: _acrKcRefreshTokenMock}
	suite.acrTokenMock.On("GetACRRefreshToken", _acrKcRegistryMock).Return(_acrKcRefreshTokenMock, nil).Once()
	kc, err := suite.factory.Create(_acrKcRegistryMock)

	suite.Nil(err)
	suite.Exactly(expectedKC, kc)
	suite.AssertExpectation()
}

func (suite *TestSuiteACRKCFactorySuite) Test_Create_TokenError() {

	expectedError := errors.New("TokenErrorMock!")
	suite.acrTokenMock.On("GetACRRefreshToken", _acrKcRegistryMock).Return("", expectedError).Once()
	kc, err := suite.factory.Create(_acrKcRegistryMock)

	suite.Error(err)
	suite.ErrorIs(err, expectedError)
	suite.Equal(nil, kc)
	suite.AssertExpectation()
}

func (suite *TestSuiteACRKCFactorySuite) Test_ACRKC_ResolveACR() {

	expectedKC := authn.FromConfig(authn.AuthConfig{IdentityToken: _acrKcRefreshTokenMock})
	acrKC := &ACRKeyChain{Token: _acrKcRefreshTokenMock}
	resource, err := name.NewRegistry("tomer.azurecr.io")
	suite.Nil(err)

	kc, err := acrKC.Resolve(resource)

	suite.Nil(err)
	suite.Exactly(expectedKC, kc)
	suite.AssertExpectation()
}

func (suite *TestSuiteACRKCFactorySuite) Test_NonACRKC_ResolveAnon() {

	expectedKC := authn.Anonymous
	acrKC := &ACRKeyChain{Token: _acrKcRefreshTokenMock}
	resource, err := name.NewRegistry("tomer.azu.io")
	suite.Nil(err)

	kc, err := acrKC.Resolve(resource)

	suite.Nil(err)
	suite.Exactly(expectedKC, kc)
	suite.AssertExpectation()
}

func (suite *TestSuiteACRKCFactorySuite) AssertExpectation() {
	suite.acrTokenMock.AssertExpectations(suite.T())
}

func Test_TestSuiteACRKCFactorySuite(t *testing.T) {
	suite.Run(t, new(TestSuiteACRKCFactorySuite))
}
