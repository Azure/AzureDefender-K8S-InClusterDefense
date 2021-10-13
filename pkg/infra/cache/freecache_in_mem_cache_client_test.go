package cache

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	_configuration = &wrappers.FreeCacheInMemWrapperCacheConfiguration{CacheSize: 10000000}
)

func (suite *TestSuite) TestFreeCacheInMemCacheClient_Get_KeyIsExist_ShouldReturnValue() {
	// Setup
	expectedValue := _value
	wrapper := wrappers.NewFreeCacheInMem(_configuration)
	wrapper.Set([]byte(_key), []byte(_value), 100)
	client := NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), wrapper)

	// Act
	actual, err := client.Get(nil, _key)

	// Test
	suite.Nil(err)
	suite.Equal(expectedValue, actual)
}

func (suite *TestSuite) TestFreeCacheInMemCacheClient_Get_KeyIsNotExist_ShouldReturnErr() {
	// Setup
	wrapper := wrappers.NewFreeCacheInMem(_configuration)
	client := NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), wrapper)

	// Act
	_, err := client.Get(nil, _key)

	// Test
	suite.NotNil(err)
}

func (suite *TestSuite) TestFreeCacheInMemCacheClient_Set_NewKey_ShouldReturnNil() {
	// Setup
	duration := time.Duration(3)
	expectedValue := _value
	wrapper := wrappers.NewFreeCacheInMem(_configuration)
	client := NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), wrapper)

	// Act
	err := client.Set(nil, _key, _value, duration)

	// Test
	suite.Nil(err)
	extractedValue, err := wrapper.Get([]byte(_key))
	suite.Nil(err)
	suite.Equal(expectedValue, string(extractedValue))
}

func (suite *TestSuite) TestFreeCacheInMemCacheClient_Set_KeyAlreadyExist_ShouldReturnNil() {
	// Setup
	duration := time.Duration(3)
	expectedValue := _value
	wrapper := wrappers.NewFreeCacheInMem(_configuration)
	wrapper.Set([]byte(_key), []byte(_value), 100)
	client := NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), wrapper)

	// Act
	err := client.Set(nil, _key, _value, duration)

	// Test
	suite.Nil(err)
	extractedValue, err := wrapper.Get([]byte(_key))
	suite.Nil(err)
	suite.Equal(expectedValue, string(extractedValue))
}

func (suite *TestSuite) TestFreeCacheInMemCacheClient_Set_NegativeExpiration_ShouldReturnErr() {
	// Setup
	duration := time.Duration(-3)
	wrapper := wrappers.NewFreeCacheInMem(_configuration)
	client := NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), wrapper)

	// Act
	err := client.Set(nil, _key, _value, duration)

	// Test
	suite.NotNil(err)
	suite.IsType(&NegativeExpirationCacheError{}, err)
}

func (suite *TestSuite) TestFreeCacheInMemCacheClient_Get_MissingKey_ShouldReturnErr() {
	// Setup
	wrapper := wrappers.NewFreeCacheInMem(_configuration)
	client := NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), wrapper)

	// Act
	val, err := client.Get(nil, _key)

	// Test
	suite.NotNil(err)
	suite.IsType(&MissingKeyCacheError{}, err)
	suite.Equal("", val)
}

func (suite *TestSuite) TestFreeCacheInMemCacheClient_Get_ExpiredKey_ShouldReturnErr() {
	// Setup
	wrapper := wrappers.NewFreeCacheInMem(_configuration)
	client := NewFreeCacheInMemCacheClient(instrumentation.NewNoOpInstrumentationProvider(), wrapper)
	duration := 1 * time.Second
	durationToSleep := 3 * time.Second

	client.Set(nil, _key, _value, duration)
	time.Sleep(durationToSleep)
	// Act
	val, err := client.Get(nil, _key)

	// Test
	suite.NotNil(err)
	suite.Equal("", val)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestFreeCacheInMemCacheClient(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
