package runtime_tests

import (
	"github.com/stretchr/testify/suite"
	"net/http"
)

type RuntimeTestSuite struct {
	suite.Suite
}

func (suite *RuntimeTestSuite) SetupTest() {
	http.Client{}
}



