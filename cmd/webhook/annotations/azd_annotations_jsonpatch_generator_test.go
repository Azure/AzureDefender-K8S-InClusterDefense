package annotations

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
}

func (suite *TestSuite) SetupSuite() {

}

func (suite *TestSuite) Test_AssertInfoMarshalString() {

}

func (suite *TestSuite) Test_ScanStausEnumsValues() {

}

func Test_ContainerScanVulnerabilities(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
