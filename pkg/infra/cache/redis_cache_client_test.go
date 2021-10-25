package cache

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct
type TestSuite struct {
	suite.Suite
}

const (
	_key   = "hello"
	_value = "world"
)

var (
	_retryPolicyConfiguration = &retrypolicy.RetryPolicyConfiguration{RetryAttempts: 1, RetryDuration: 10, TimeUnit: "ms"}
	_retryPolicy, _           = retrypolicy.NewRetryPolicy(instrumentation.NewNoOpInstrumentationProvider(), _retryPolicyConfiguration)
)

func (suite *TestSuite) Test_Get_KeyIsExist_ShouldReturnValue() {
	// Setup
	expectedValue := _value

	clientMock, mock := redismock.NewClientMock()
	mock.ExpectGet(_key).SetVal(expectedValue)
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock, _retryPolicy)

	// Act
	actual, err := client.Get(_key)

	// Test
	suite.Nil(err)
	suite.Equal(expectedValue, actual)
}

func (suite *TestSuite) Test_Get_KeyIsNotExist_ShouldReturnErr() {
	// Setup
	clientMock, mock := redismock.NewClientMock()
	mock.ExpectGet(_key).SetErr(redis.Nil)
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock, _retryPolicy)

	// Act
	_, err := client.Get(_key)

	// Test
	suite.NotNil(err)
}

func (suite *TestSuite) Test_Set_NewKey_ShouldReturnNil() {
	// Setup
	duration := time.Duration(3)
	clientMock, mock := redismock.NewClientMock()
	mock.ExpectSet(_key, _value, duration).RedisNil()
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock, _retryPolicy)

	// Act
	err := client.Set(_key, _value, duration)
	suite.Nil(err)
}

func (suite *TestSuite) Test_Set_NegativeExpiration_ShouldReturnErr() {
	// Setup
	duration := time.Duration(-3)
	clientMock, mock := redismock.NewClientMock()
	mock.ExpectSet(_key, _value, duration).SetVal(_value)
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock, _retryPolicy)

	// Act
	err := client.Set(_key, _value, duration)
	suite.IsType(&NegativeExpirationCacheError{}, err)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestRedisCacheClient(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
