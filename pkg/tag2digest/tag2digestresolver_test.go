package tag2digest

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	registrymocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/mocks"
	registryutils "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuiteTag2DigestResolver struct {
	suite.Suite
}

var _resolver *Tag2DigestResolver
var _registryClientMock *registrymocks.IRegistryClient
var _acrImageRefTag registry.IImageReference
var _nonAcrImageRefTag registry.IImageReference
var _ctx *ResourceContext
var _ctxPullSecrets = []string{"tomer-pull-secret"}

const _ctxNamsespace = "tomer-ns"
const _ctsServiceAccount = "tomer-sa"
const _expectedDigest = "sha256:3f85bbca16d5803f639ae7e7822c8c6686deff624de774805ab7e30d0f66e089"

func (suite *TestSuiteTag2DigestResolver) SetupTest() {
	instrumentationP := instrumentation.NewNoOpInstrumentationProvider()
	_registryClientMock = new(registrymocks.IRegistryClient)
	_resolver = NewTag2DigestResolver(instrumentationP, _registryClientMock)
	_acrImageRefTag, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_nonAcrImageRefTag, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_ctx = NewResourceContext(_ctxNamsespace, _ctxPullSecrets, _ctsServiceAccount)
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_DigestRefernce_ReturnDigestNoRegistryCalls() {

	imageRef, _ := registryutils.GetImageReference("tomerw.xyz.io/redis@" + _expectedDigest)
	digest, err := _resolver.Resolve(imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthFailK8SAuthSuccess() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultSuccess() {

	_registryClientMock.On("GetDigestUsingACRAttachAuth", _acrImageRefTag).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _acrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _acrImageRefTag).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_acrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {

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

	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_nonAcrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_K8SAuthFailDefaultSuccess() {

	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _nonAcrImageRefTag).Return(_expectedDigest, nil).Once()
	digest, err := _resolver.Resolve(_nonAcrImageRefTag, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {

	expectedError := errors.New("DefaultAUthError")
	_registryClientMock.On("GetDigestUsingK8SAuth", _nonAcrImageRefTag, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _nonAcrImageRefTag).Return("", expectedError).Once()

	digest, err := _resolver.Resolve(_nonAcrImageRefTag, _ctx)

	suite.ErrorIs(err , expectedError)
	suite.Equal("", digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func Test_Suite(t *testing.T) {
	suite.Run(t, new(TestSuiteTag2DigestResolver))
}
//TODO add tests for empty digests returned..