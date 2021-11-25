package cache

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct
type TestSuiteSafeCache struct {
	suite.Suite
}

const (
	key   = "hello"
	value = "world"
)

var (
	cacheContext = context.Background()
	retryPolicy     retrypolicy.IRetryPolicy
	redisClient          *RedisCacheClient
	redisClientMock *redis.Client
	redisMock       redismock.ClientMock
	client *SafeCacheClient
)

func (suite *TestSuiteSafeCache) SetupTest() {
	// TODO retry mocking in all places
	redisClientMock, redisMock = redismock.NewClientMock()
	retryPolicyConfiguration := &retrypolicy.RetryPolicyConfiguration{RetryAttempts: 1, RetryDurationInMS: 10}
	retryPolicy = retrypolicy.NewRetryPolicy(instrumentation.NewNoOpInstrumentationProvider(), retryPolicyConfiguration)
	redisClient = NewRedisCacheClient(instrumentation.NewNoOpInstrumentationProvider(), redisClientMock, retryPolicy, cacheContext)
	client = NewSafeCacheClient(redisClient)
}

func (suite *TestSuiteSafeCache) Test_Set_NewKey_ShouldReturnNil() {
	// Setup
	duration := time.Duration(3)
	redisMock.ExpectSet(_key, _value, duration).RedisNil()

	// Act
	err := client.Set(_key, _value, duration)
	suite.Nil(err)
}

func (suite *TestSuiteSafeCache) Test_Set_ZeroExpiration_ShouldReturnErr() {
	// Setup
	duration := time.Duration(0)
	redisMock.ExpectSet(key, value, duration).SetVal(value)

	// Act
	err := client.Set(key, value, duration)
	suite.IsType(utils.SetKeyInCacheWithNoExpiration, err)
}


// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestSafeCacheClient(t *testing.T) {
	suite.Run(t, new(TestSuiteSafeCache))
}
