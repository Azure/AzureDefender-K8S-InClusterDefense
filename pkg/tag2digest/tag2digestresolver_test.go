package tag2digest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	cachemock "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registrymocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/mocks"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestSuiteTag2DigestResolver struct {
	suite.Suite
}

var _resolverNoCacheFunctionality *Tag2DigestResolver
var _resolverWithCacheFunctionality *Tag2DigestResolver
var _registryClientMock *registrymocks.IRegistryClient
var _cacheClientMock *cachemock.ICacheClient
var _cacheClientInMemBasedMock cache.ICacheClient
var _acrImageRefTag registry.IImageReference
var _nonAcrImageRefTag registry.IImageReference
var _ctx *ResourceContext
var _ctxPullSecrets = []string{"tomer-pull-secret"}
var _expirationTime = time.Duration(1)

const _ctxNamsespace = "tomer-ns"
const _ctsServiceAccount = "tomer-sa"
const _expectedDigest = "sha256:3f85bbca16d5803f639ae7e7822c8c6686deff624de774805ab7e30d0f66e089"

func (suite *TestSuiteTag2DigestResolver) SetupTest() {
	instrumentationP := instrumentation.NewNoOpInstrumentationProvider()
	_registryClientMock = new(registrymocks.IRegistryClient)
	_cacheClientMock = new(cachemock.ICacheClient)
	_cacheClientInMemBasedMock = cachemock.NewICacheInMemBasedMock()
	_resolverNoCacheFunctionality = NewTag2DigestResolver(instrumentationP, _registryClientMock, _cacheClientMock, new(Tag2DigestResolverConfiguration))
	_resolverWithCacheFunctionality = NewTag2DigestResolver(instrumentationP, _registryClientMock, _cacheClientInMemBasedMock, &Tag2DigestResolverConfiguration{
			cacheExpirationTime: _expirationTime,
		})
	_acrImageRefTag, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_nonAcrImageRefTag, _ = registryutils.GetImageReference("tomerw.nonacr.io/redis:v0")
	_ctx = NewResourceContext(_ctxNamsespace, _ctxPullSecrets, _ctsServiceAccount)
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_DigestRefernce_ReturnDigestNoRegistryCalls() {

	imageRef, _ := registryutils.GetImageReference("tomerw.xyz.io/redis@" + _expectedDigest)
	digest, err := _resolverNoCacheFunctionality.Resolve(imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()

	digest, err := _resolverNoCacheFunctionality.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthFailK8SAuthSuccess() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolverNoCacheFunctionality.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultSuccess() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()

	digest, err := _resolverNoCacheFunctionality.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {

	expectedError := errors.New("DefaultAuthError")
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAuthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAuthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _acrImageRefTag).Return("", expectedError).Once()

	digest, err := _resolverNoCacheFunctionality.Resolve(_acrImageRefTag, _ctx)

	suite.ErrorIs(err , expectedError)
	suite.Equal("", digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_K8SAuthSuccess() {

	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolverNoCacheFunctionality.Resolve(_nonAcrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_K8SAuthFailDefaultSuccess() {

	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _nonAcrImageRefTag).Return(_expectedDigest, nil).Once()
	digest, err := _resolverNoCacheFunctionality.Resolve(_nonAcrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {

	expectedError := errors.New("DefaultAUthError")
	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _nonAcrImageRefTag).Return("", expectedError).Once()

	digest, err := _resolverNoCacheFunctionality.Resolve(_nonAcrImageRefTag, _ctx)

	suite.ErrorIs(err , expectedError)
	suite.Equal("", digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_NoKeyInCache() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()
	_, err := _cacheClientInMemBasedMock.Get(_acrImageRefTag.Original())
	suite.NotNil(err)
	digest, err := _resolverWithCacheFunctionality.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	digest, err = _cacheClientInMemBasedMock.Get(_acrImageRefTag.Original())
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_KeyInCache() {
	_ = _cacheClientInMemBasedMock.Set(_acrImageRefTag.Original(), _expectedDigest, _expirationTime)
	digest, err := _cacheClientInMemBasedMock.Get(_acrImageRefTag.Original())
	suite.Nil(err)
	digest, err = _resolverWithCacheFunctionality.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime_ScannedResults() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()
	_, err := _cacheClientInMemBasedMock.Get(_acrImageRefTag.Original())
	suite.NotNil(err)
	digest, err := _resolverWithCacheFunctionality.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	digest, err = _cacheClientInMemBasedMock.Get(_acrImageRefTag.Original())
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	digest, err = _resolverWithCacheFunctionality.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_NoKeyInCache_SetKey_GetKeySecondTryAfterExpirationTime_ScannedResults() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Twice()
	_, err := _cacheClientInMemBasedMock.Get(_acrImageRefTag.Original())
	suite.NotNil(err)
	digest, err := _resolverWithCacheFunctionality.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	digest, err = _cacheClientInMemBasedMock.Get(_acrImageRefTag.Original())
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	time.Sleep(time.Second)
	digest, err = _resolverWithCacheFunctionality.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}

func Test_Suite(t *testing.T) {
	suite.Run(t, new(TestSuiteTag2DigestResolver))
}
//TODO add tests for empty digests returned..