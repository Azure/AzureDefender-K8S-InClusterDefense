package arg

import (
	"context"
	"errors"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/wrappers/mocks"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"reflect"
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
	_emptyQueryResponse = resourcegraph.QueryResponse{}
	_emptyErrorString   = errors.New("")

	_firstObjectForDataArray  = int32(1)
	_secondObjectForDataArray = int32(2)
	_thirdObjectForDataArray  = int32(3)
)

// This will run before each test in the suite
func (suite *TestSuite) SetupTest() {
	suite.argBaseClientWrapperMock = &mocks.IARGBaseClientWrapper{}
}

func (suite *TestSuite) Test_QueryResources_ReturnedErrorFromArgBaseClient_ShouldReturnError() {
	// Setup
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), mock.AnythingOfType("resourcegraph.QueryRequest")).Return(_emptyQueryResponse, _emptyErrorString)
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.Nil(resources)
	suite.NotNil(err)
}

func (suite *TestSuite) Test_QueryResources_ResponseDataIsNil_ShouldReturnError() {
	// Setup
	totalRecords := int64(1)
	response := resourcegraph.QueryResponse{TotalRecords: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), mock.AnythingOfType("resourcegraph.QueryRequest")).Return(response, nil)
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.Nil(resources)
	suite.NotNil(err)
}

func (suite *TestSuite) Test_QueryResources_ResponseTotalRecordsIsNil_ShouldReturnError() {
	// Setup
	totalRecords := int64(1)
	response := resourcegraph.QueryResponse{Data: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), mock.AnythingOfType("resourcegraph.QueryRequest")).Return(response, nil)
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.Nil(resources)
	suite.NotNil(err)
}

func (suite *TestSuite) Test_QueryResources_ResponseBadFormatOfResponseData_ShouldReturnError() {
	// Setup
	tableData := resourcegraph.Table{}
	totalRecords := int64(0)
	response := resourcegraph.QueryResponse{Data: tableData, TotalRecords: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), mock.AnythingOfType("resourcegraph.QueryRequest")).Return(response, nil)
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())
	client.argQueryReqOptions.ResultFormat = resourcegraph.ResultFormatTable

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.Nil(resources)
	suite.Equal(_errArgQueryResponseIsNotAnObjectListFormat, err)
}

func (suite *TestSuite) Test_QueryResources_ResponseZeroData_ShouldReturnEmptyArray() {
	// Setup
	var arrayData []interface{}
	totalRecords := int64(0)
	response := resourcegraph.QueryResponse{Data: arrayData, TotalRecords: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), mock.AnythingOfType("resourcegraph.QueryRequest")).Return(response, nil)
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.NotNil(resources)
	suite.Nil(err)
	suite.Equal(0, len(resources))
}

func (suite *TestSuite) Test_QueryResources_Response2Items_ShouldReturnArrayWith2Items() {
	// Setup
	arrayData := make([]interface{}, 0, 2)
	arrayData = append(arrayData, _firstObjectForDataArray, _secondObjectForDataArray)
	totalRecords := int64(2)
	response := resourcegraph.QueryResponse{Data: arrayData, TotalRecords: &totalRecords, Count: &totalRecords}
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), mock.AnythingOfType("resourcegraph.QueryRequest")).Return(response, nil)
	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())

	// Act
	resources, err := client.QueryResources(_invalidQuery)

	// Test
	suite.NotNil(resources)
	suite.Nil(err)
	suite.Equal(2, len(resources))
	expected := make([]interface{}, 0, 2)
	expected = append(expected, _firstObjectForDataArray, _secondObjectForDataArray)
	suite.True(reflect.DeepEqual(expected, resources))
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
	firstResponse := resourcegraph.QueryResponse{Data: firstArrayData, TotalRecords: &totalRecords, Count: &firstCount, SkipToken: &skipToken}
	requestSkipTokenNilArgument := mock.MatchedBy(func(req resourcegraph.QueryRequest) bool { return req.Options.SkipToken == nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNilArgument).Return(firstResponse, nil)

	// Set second response
	secondArrayData := make([]interface{}, 0, 1)
	secondArrayData = append(secondArrayData, _thirdObjectForDataArray)
	secondCount := int64(1)

	secondResponse := resourcegraph.QueryResponse{Data: secondArrayData, TotalRecords: &totalRecords, Count: &secondCount}

	requestSkipTokenNotNilArgument := mock.MatchedBy(func(req resourcegraph.QueryRequest) bool { return req.Options.SkipToken != nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNotNilArgument).Return(secondResponse, nil)

	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())

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
	firstResponse := resourcegraph.QueryResponse{Data: firstArrayData, TotalRecords: &totalRecords, Count: &firstCount, SkipToken: &skipToken}
	requestSkipTokenNilArgument := mock.MatchedBy(func(req resourcegraph.QueryRequest) bool { return req.Options.SkipToken == nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNilArgument).Return(firstResponse, nil)

	// Set second response
	secondArrayData := make([]interface{}, 0, 1)
	secondArrayData = append(secondArrayData, _thirdObjectForDataArray)
	secondCount := int64(1)

	// NOTE that I deleted the total records from the initialization so there supposed to be an error.
	secondResponse := resourcegraph.QueryResponse{Data: secondArrayData, Count: &secondCount}

	requestSkipTokenNotNilArgument := mock.MatchedBy(func(req resourcegraph.QueryRequest) bool { return req.Options.SkipToken != nil })
	suite.argBaseClientWrapperMock.On("Resources", context.Background(), requestSkipTokenNotNilArgument).Return(secondResponse, nil)

	client := NewARGClient(instrumentation.NewNoOpInstrumentationProvider(), suite.argBaseClientWrapperMock, _getARGClientConfiguration())

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

func _getARGClientConfiguration() *ARGClientConfiguration {
	return &ARGClientConfiguration{Subscriptions: []string{}}

}
