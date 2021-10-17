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
var _imageRef registry.IImageReference
var _ctx *ResourceContext
var _ctxPullSecrets = []string{"tomer-pull-secret"}

const _ctxNamsespace = "tomer-ns"
const _ctsServiceAccount = "tomer-sa"
const _expectedDigest = "sha256:3f85bbca16d5803f639ae7e7822c8c6686deff624de774805ab7e30d0f66e089"

func (suite *TestSuiteTag2DigestResolver) SetupTest() {
	instrumentationP := instrumentation.NewNoOpInstrumentationProvider()
	_registryClientMock = new(registrymocks.IRegistryClient)
	_resolver = NewTag2DigestResolver(instrumentationP, _registryClientMock)
	_imageRef, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_ctx = NewResourceContext(_ctxNamsespace, _ctxPullSecrets, _ctsServiceAccount)
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_DigestRefernce_ReturnDigestNoRegistryCalls() {

	//_registryClientMock.On("")
	_imageRef, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis@" + _expectedDigest)
	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthSuccess() {

	_imageRef, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _imageRef).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthFailK8SAuthSuccess() {

	_imageRef, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _imageRef).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _imageRef, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultSuccess() {

	_imageRef, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _imageRef).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _imageRef, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _imageRef).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_ACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {

	expectedError := errors.New("DefaultAUthError")
	_imageRef, _ = registryutils.GetImageReference("tomerw.azurecr.io/redis:v0")
	_registryClientMock.On("GetDigestUsingACRAttachAuth", _imageRef).Return("", errors.New("ACRAUthError")).Once()
	_registryClientMock.On("GetDigestUsingK8SAuth", _imageRef, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _imageRef).Return("", expectedError).Once()

	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.ErrorIs(err , expectedError)
	suite.Equal("", digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_K8SAuthSuccess() {

	_imageRef, _ = registryutils.GetImageReference("tomerw.nonacr.io/redis:v0")
	_registryClientMock.On("GetDigestUsingK8SAuth", _imageRef, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return(_expectedDigest, nil).Once()

	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_K8SAuthFailDefaultSuccess() {

	_imageRef, _ = registryutils.GetImageReference("tomerw.nonacr.io/redis:v0")
	_registryClientMock.On("GetDigestUsingK8SAuth", _imageRef, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _imageRef).Return(_expectedDigest, nil).Once()
	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.Nil(err)
	suite.Equal(_expectedDigest, digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func (suite *TestSuiteTag2DigestResolver) Test_Resolve_NonACRReference_ACRAuthAndK8SAuthFailDefaultFail_ReflectError() {

	expectedError := errors.New("DefaultAUthError")
	_imageRef, _ = registryutils.GetImageReference("tomerw.nonacr.io/redis:v0")
	_registryClientMock.On("GetDigestUsingK8SAuth", _imageRef, _ctxNamsespace, _ctx.imagePullSecrets, _ctsServiceAccount).Return("", errors.New("K8SAUthError")).Once()
	_registryClientMock.On("GetDigestUsingDefaultAuth", _imageRef).Return("", expectedError).Once()

	digest, err := _resolver.Resolve(_imageRef, _ctx)

	suite.ErrorIs(err , expectedError)
	suite.Equal("", digest)

	_registryClientMock.AssertExpectations(suite.T())
}

func Test_Suite(t *testing.T) {
	suite.Run(t, new(TestSuiteTag2DigestResolver))
}
