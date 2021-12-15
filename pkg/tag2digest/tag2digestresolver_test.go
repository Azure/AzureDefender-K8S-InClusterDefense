package tag2digest

import (
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
)

type TestSuiteTag2DigestResolver struct {
	suite.Suite
}

var _resolver *Tag2DigestResolver
var _registryClientMock *registrymocks.IRegistryClient
var _cacheClientMock *cachemock.ICacheClient
var _acrImageRefTag registry.IImageReference
var _nonAcrImageRefTag registry.IImageReference
var _ctx *ResourceContext
var _ctxPullSecrets = []string{"tomer-pull-secret"}
var _expirationTime= 1

const _ctxNamsespace = "tomer-ns"
const _ctsServiceAccount = "tomer-sa"
const _expectedDigest = "sha256:3f85bbca16d5803f639ae7e7822c8c6686deff624de774805ab7e30d0f66e089"

func (suite *TestSuiteTag2DigestResolver) SetupTest() {
	instrumentationP := instrumentation.NewNoOpInstrumentationProvider()
	_registryClientMock = new(registrymocks.IRegistryClient)
	_cacheClientMock = new(cachemock.ICacheClient)
	_tag2DigestResolverConfiguration := &Tag2DigestResolverConfiguration{ CacheExpirationTimeForResults: _expirationTime}
	_resolver = NewTag2DigestResolver(instrumentationP, _registryClientMock, _cacheClientMock, _tag2DigestResolverConfiguration)
	_acrImageRefTag, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_nonAcrImageRefTag, _ = registryutils.GetImageReference("tomerw.nonacr.io/redis:v0")
	_ctx = NewResourceContext(_ctxNamsespace, _ctxPullSecrets, _ctsServiceAccount)
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_DigestRefernce_ReturnDigestNoRegistryCalls() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	imageRef, _ := registryutils.GetImageReference("tomerw.xyz.io/redis@" + _expectedDigest)
	digest, err := _resolver.Resolve(imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthFailK8SAuthSuccess() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultSuccess() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	expectedError := errors.New("DefaultAuthError")
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAuthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAuthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _acrImageRefTag).Return("", expectedError).Once()

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)

	suite.ErrorIs(err , expectedError)
	suite.Equal("", digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_K8SAuthSuccess() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_nonAcrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_K8SAuthFailDefaultSuccess() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _nonAcrImageRefTag).Return(_expectedDigest, nil).Once()
	digest, err := _resolver.Resolve(_nonAcrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {
	_cacheClientMock.On("Get", mock.Anything).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(utils.NilArgumentError)
	expectedError := errors.New("DefaultAUthError")
	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _nonAcrImageRefTag).Return("", expectedError).Once()

	digest, err := _resolver.Resolve(_nonAcrImageRefTag, _ctx)

	suite.ErrorIs(err , expectedError)
	suite.Equal("", digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_NoKeyInCache() {
	_cacheClientMock.On("Get", _acrImageRefTag.Original()).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", _acrImageRefTag.Original(), mock.Anything, mock.Anything).Return(nil)

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()
	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_KeyInCache() {
	_cacheClientMock.On("Get", _acrImageRefTag.Original()).Return(_expectedDigest, nil)
	_cacheClientMock.On("Set", _acrImageRefTag.Original(), mock.Anything, mock.Anything).Return(nil)

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_NoKeyInCache_SetKey_GetKeySecondTryBeforeExpirationTime() {
	_cacheClientMock.On("Get", _acrImageRefTag.Original()).Return("", utils.NilArgumentError).Once()
	_cacheClientMock.On("Get", _acrImageRefTag.Original()).Return(_expectedDigest, nil).Once()
	_cacheClientMock.On("Set", _acrImageRefTag.Original(), mock.Anything, mock.Anything).Return(nil)

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()
	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	digest, err = _resolver.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess_NoKeyInCache_SetKey_GetKeySecondTryAfterExpirationTime() {
	_cacheClientMock.On("Get", _acrImageRefTag.Original()).Return("", utils.NilArgumentError)
	_cacheClientMock.On("Set", _acrImageRefTag.Original(), mock.Anything, mock.Anything).Return(nil)

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Twice()
	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	digest, err = _resolver.Resolve(_acrImageRefTag, _ctx)
	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)
	_registryClientMock.AssertExpectations(suite.T())
}


func Test_Suite(t *testing.T) {
	suite.Run(t, new(TestSuiteTag2DigestResolver))
}
//TODO add tests for empty digests returned..