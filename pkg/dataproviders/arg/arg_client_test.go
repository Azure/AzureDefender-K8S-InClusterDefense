package arg

import (
	"context"
	"errors"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/wrappers/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	argsdk "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"reflect"
	"strings"
	"testing"
)

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct
type TestSuite struct {
	suite.Suite
	argBaseClientWrapperMock *mocks.IARGBaseClientWrapper
}

const (
	_invalidQuery = "invalid query"
)

var (
	_emptyQueryResponse = argsdk.QueryResponse{}
	_emptyErrorString   = errors.New("")
	_retryPolicy *retrypolicy.RetryPolicy
	_request argsdk.QueryRequest

	_firstObjectForDataArray  = int32(1)
	_secondObjectForDataArray = int32(2)
	_thirdObjectForDataArray  = int32(3)
)

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	suite.argBaseClientWrapperMock = &mocks.IARGBaseClientWrapper{}
	retryPolicyConfiguration := &retrypolicy.RetryPolicyConfiguration{RetryAttempts: 2, RetryDurationInMS: 10}
	_retryPolicy = retrypolicy.NewRetryPolicy(instrumentation.NewNoOpInstrumentationProvider(), retryPolicyConfiguration)

	var top int32 = 1000
	// Create the query request
	_request = argsdk.QueryRequest{
		Options:       &argsdk.QueryRequestOptions{
			ResultFormat: argsdk.ResultFormatObjectArray,
			Top:          &top,
		},
		Subscriptions: &_getARGClientConfiguration().Subscriptions,
	}
}

func (suite *TestSuite) Test_QueryResources_ReturnedErrorFromArgBaseClient_ShouldReturnError() {
	// Setup
	query := _invalidQuery
	_request.Query = &query
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), _request).Return(_emptyQueryResponse, _emptyErrorString).Once()
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)

	// Act
	resources, err := client.QueryResources(query)

	// Test
	suite.Nil(resources)
	suite.NotNil(err)
	suite.argBaseClientWrapperMock.AssertExpectations(suite.T())
}

func (suite *TestSuite) Test_QueryResources_ResponseDataIsNil_ShouldReturnError() {
	// Setup
	query := _invalidQuery
	_request.Query = &query
	totalRecords := int64(1)
	response := argsdk.QueryResponse{TotalRecords: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), _request).Return(response, nil).Once()
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)

	// Act
	resources, err := client.QueryResources(query)

	// Test
	suite.Nil(resources)
	suite.True(err != nil  && strings.Contains(err.Error(), "ARG query response with nil TotalRecords"))
	suite.argBaseClientWrapperMock.AssertExpectations(suite.T())
}

func (suite *TestSuite) Test_QueryResources_ResponseTotalRecordsIsNil_ShouldReturnError() {
	// Setup
	query := _invalidQuery
	_request.Query = &query
	totalRecords := int64(1)
	response := argsdk.QueryResponse{Data: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), _request).Return(response, nil).Once()
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)

	// Act
	resources, err := client.QueryResources(query)

	// Test
	suite.Nil(resources)
	suite.NotNil(err)
	suite.True(err != nil  && strings.Contains(err.Error(), "ARG query response with nil TotalRecords"))
	suite.argBaseClientWrapperMock.AssertExpectations(suite.T())
}

func (suite *TestSuite) Test_QueryResources_ResponseBadFormatOfResponseData_ShouldReturnError() {
	// Setup
	query := _invalidQuery
	_request.Query = &query
	_request.Options.ResultFormat = argsdk.ResultFormatTable
	tableData := argsdk.Table{}
	totalRecords := int64(0)
	response := argsdk.QueryResponse{Data: tableData, TotalRecords: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), _request).Return(response, nil).Once()
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)
	client.argQueryReqOptions.ResultFormat = argsdk.ResultFormatTable
	// Act
	resources, err := client.QueryResources(query)

	// Test
	suite.Nil(resources)
	suite.Equal(_errArgQueryResponseIsNotAnObjectListFormat, err)
	suite.argBaseClientWrapperMock.AssertExpectations(suite.T())
}

func (suite *TestSuite) Test_QueryResources_ResponseZeroData_ShouldReturnEmptyArray_RetryTwice() {
	// Setup
	var arrayData []interface{}
	query := _invalidQuery
	_request.Query = &query
	totalRecords := int64(0)
	response := argsdk.QueryResponse{Data: arrayData, TotalRecords: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), _request).Return(response, nil).Times(2)
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)

	// Act
	resources, err := client.QueryResources(query)

	// Test
	suite.NotNil(resources)
	suite.Nil(err)
	suite.Equal(0, len(resources))
	suite.Equal(make([]interface{}, 0, *response.TotalRecords), resources)
	suite.argBaseClientWrapperMock.AssertExpectations(suite.T())
}

func (suite *TestSuite) Test_QueryResources_Response2Items_ShouldReturnArrayWith2Items() {
	// Setup
	query := _invalidQuery
	_request.Query = &query
	arrayData := make([]interface{}, 0, 2)
	arrayData = append(arrayData, _firstObjectForDataArray, _secondObjectForDataArray)
	totalRecords := int64(2)
	response := argsdk.QueryResponse{Data: arrayData, TotalRecords: &totalRecords, Count: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), _request).Return(response, nil).Once()
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)

	// Act
	resources, err := client.QueryResources(query)

	// Test
	suite.NotNil(resources)
	suite.Nil(err)
	suite.Equal(2, len(resources))
	expected := make([]interface{}, 0, 2)
	expected = append(expected, _firstObjectForDataArray, _secondObjectForDataArray)
	suite.True(reflect.DeepEqual(expected, resources))
	suite.argBaseClientWrapperMock.AssertExpectations(suite.T())
}

func (suite *TestSuite) Test_QueryResources_Response3ItemsTop2_ShouldReturnArrayWith3ItemsAfterPagination() {
	// Setup
	totalRecords := int64(3)

	// Set first response
	firstCount := int64(2)
	firstArrayData := make([]interface{}, 0, 2)
	firstArrayData = append(firstArrayData, _firstObjectForDataArray, _secondObjectForDataArray)
	// Set skiptoken for the first response
	skipToken := "skiptoken"
	firstResponse := argsdk.QueryResponse{Data: firstArrayData, TotalRecords: &totalRecords, Count: &firstCount, SkipToken: &skipToken}
	requestSkipTokenNilArgument := mock.MatchedBy(func(req argsdk.QueryRequest) bool { return req.Options.SkipToken == nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNilArgument).Return(firstResponse, nil)

	// Set second response
	secondArrayData := make([]interface{}, 0, 1)
	secondArrayData = append(secondArrayData, _thirdObjectForDataArray)
	secondCount := int64(1)

	secondResponse := argsdk.QueryResponse{Data: secondArrayData, TotalRecords: &totalRecords, Count: &secondCount}

	requestSkipTokenNotNilArgument := mock.MatchedBy(func(req argsdk.QueryRequest) bool { return req.Options.SkipToken != nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNotNilArgument).Return(secondResponse, nil)

	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.NotNil(resources)
	suite.Nil(err)
	suite.Equal(3, len(resources))
	expected := make([]interface{}, 0, 3)
	expected = append(expected, _firstObjectForDataArray, _secondObjectForDataArray, _thirdObjectForDataArray)
	suite.True(reflect.DeepEqual(expected, resources))
}

func (suite *TestSuite) Test_QueryResources_ResponseErrorInTheSecondPage_ShouldReturnError() {
	// Setup
	totalRecords := int64(3)

	// Set first response
	firstCount := int64(2)
	firstArrayData := make([]interface{}, 0, 2)
	firstArrayData = append(firstArrayData, _firstObjectForDataArray, _secondObjectForDataArray)
	// Set skiptoken for the first response
	skipToken := "skiptoken"
	firstResponse := argsdk.QueryResponse{Data: firstArrayData, TotalRecords: &totalRecords, Count: &firstCount, SkipToken: &skipToken}
	requestSkipTokenNilArgument := mock.MatchedBy(func(req argsdk.QueryRequest) bool { return req.Options.SkipToken == nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNilArgument).Return(firstResponse, nil)

	// Set second response
	secondArrayData := make([]interface{}, 0, 1)
	secondArrayData = append(secondArrayData, _thirdObjectForDataArray)
	secondCount := int64(1)

	// NOTE that I deleted the total records from the initialization so there supposed to be an error.
	secondResponse := argsdk.QueryResponse{Data: secondArrayData, Count: &secondCount}

	requestSkipTokenNotNilArgument := mock.MatchedBy(func(req argsdk.QueryRequest) bool { return req.Options.SkipToken != nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNotNilArgument).Return(secondResponse, nil)

	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration(), _retryPolicy)

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.Nil(resources)
	suite.NotNil(err)
}

// We need this function to kick off the test suite, otherwise
// "go test" won't know about our tests
func TestArgClientTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// _getARGClientConfiguration returns default ARGClientConfiguration. needed for tests
func _getARGClientConfiguration() *ARGClientConfiguration {
	return &ARGClientConfiguration{Subscriptions: []string{"dddd"}}

}
