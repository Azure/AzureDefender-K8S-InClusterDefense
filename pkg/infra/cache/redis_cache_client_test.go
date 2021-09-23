package cache

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/pkg/errors"
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
	_ctx = context.Background()
)

func (suite *TestSuite) Test_Get_KeyIsExist_ShouldReturnValue() {
	// Setup
	expectedValue := _value

	clientMock, mock := redismock.NewClientMock()
	mock.ExpectGet(_key).SetVal(expectedValue)
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock)

	// Act
	actual, err := client.Get(_ctx, _key)

	// Test
	suite.Nil(err)
	suite.Equal(expectedValue, actual)
}

func (suite *TestSuite) Test_Get_KeyIsNotExist_ShouldReturnErr() {
	// Setup
	clientMock, mock := redismock.NewClientMock()
	mock.ExpectGet(_key).SetErr(errors.New("key is not exist"))
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock)

	// Act
	_, err := client.Get(_ctx, _key)

	// Test
	suite.NotNil(err)
}

func (suite *TestSuite) Test_Set_NewKey_ShouldReturnNil() {
	// Setup
	duration := time.Duration(3)
	clientMock, mock := redismock.NewClientMock()
	mock.ExpectSet(_key, _value, duration).RedisNil()
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock)

	// Act
	err := client.Set(_ctx, _key, _value, duration)
	suite.IsType(redis.Nil, err)
}

func (suite *TestSuite) Test_Set_NegativeExpiration_ShouldReturnErr() {
	// Setup
	duration := time.Duration(-3)
	clientMock, mock := redismock.NewClientMock()
	mock.ExpectSet(_key, _value, duration).SetVal(_value)
	client := NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), clientMock)

	// Act
	err := client.Set(_ctx, _key, _value, duration)
	suite.IsType(&NegativeExpirationCacheError{}, err)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestRedisCacheClient(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
