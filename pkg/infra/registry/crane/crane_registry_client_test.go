package crane

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/crane/mocks"
	wrappersmocks "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/wrappers/mocks"
	"github.com/stretchr/testify/suite"
	"testing"
)

const _expectedDigestMock = "xxyxyyxxyxsss"
var _mockACRKC_RegistryClient = &ACRKeyChain{Token: "kakaksjdjkd"}

type CraneRegistryTestSuite struct {
	suite.Suite
	client           *CraneRegistryClient
	craneWrapperMock *wrappersmocks.ICraneWrapper
	acrKCFactoryMock *mocks.IACRKeychainFactory
	k8sKCFactoryMock *mocks.IK8SKeychainFactory
}

func (suite *CraneRegistryTestSuite) SetupTest() {
	instrumentationP := instrumentation.NewNoOpInstrumentationProvider()
	suite.craneWrapperMock = new(wrappersmocks.ICraneWrapper)
	suite.acrKCFactoryMock = new(mocks.IACRKeychainFactory)
	suite.k8sKCFactoryMock = new(mocks.IK8SKeychainFactory)
	suite.client = NewCraneRegistryClient(instrumentationP, suite.craneWrapperMock, suite.acrKCFactoryMock, suite.k8sKCFactoryMock)
}

func (suite *CraneRegistryTestSuite) Test_GetDigest() {
//TODO
}

func (suite *CraneRegistryTestSuite) AssertExpectation() {
	suite.craneWrapperMock.AssertExpectations(suite.T())
	suite.acrKCFactoryMock.AssertExpectations(suite.T())
	suite.k8sKCFactoryMock.AssertExpectations(suite.T())
}

func Test_CraneRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(CraneRegistryTestSuite))
}
