package arg

import (
	"context"
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/wrappers"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	arg "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"github.com/pkg/errors"
)

// MAX_TOP_RESULTS_IN_PAGE_OF_ARG is the maximum. please see more information in https://docs.microsoft.com/en-us/azure/governance/resource-graph/concepts/work-with-data#paging-results
const MAX_TOP_RESULTS_IN_PAGE_OF_ARG = 1000

var (
	_errArgQueryResponseIsNotAnObjectListFormat = fmt.Errorf("ARGClient.QueryResources ARG query response data is not an object list")
)

// IARGClient is an interface for our arg client implementation
type IARGClient interface {
	// QueryResources gets a query and return an array object as a result
	QueryResources(query string) ([]interface{}, error)
}

// ARGClient is our implementation for ARG client
type ARGClient struct {
	tracerProvider  trace.ITracerProvider
	metricSubmitter metric.IMetricSubmitter
	// argBaseClientWrapper is the wrapper for ARG base client for the Resources function.
	argBaseClientWrapper wrappers.IARGBaseClientWrapper
	//argQueryReqOptions is the options for query evaluation of the ARGClient
	argQueryReqOptions *arg.QueryRequestOptions
	subscriptions      *[]string
}

type ARGClientConfiguration struct {
	// Subscriptions is array of subscriptions that will be the scope of the query to ARG.
	Subscriptions []string
}

// NewARGClient Constructor
func NewARGClient(instrumentationProvider instrumentation.IInstrumentationProvider, argBaseClientWrapper wrappers.IARGBaseClientWrapper, configuration *ARGClientConfiguration) *ARGClient {
	// We need this var for unittests - in unittests we reduce it from 1000 to smaller number.
	requestQueryTop := int32(MAX_TOP_RESULTS_IN_PAGE_OF_ARG)
	subscriptions := &configuration.Subscriptions
	// If the subscriptions is empty then work on tenat scope.
	// TODO Is it the behavior that we want? bad performance.
	if len(*subscriptions) == 0 {
		subscriptions = nil
	}

	return &ARGClient{
		tracerProvider:       instrumentationProvider.GetTracerProvider("ARGClient"),
		metricSubmitter:      instrumentationProvider.GetMetricSubmitter(),
		argBaseClientWrapper: argBaseClientWrapper,
		argQueryReqOptions:   &arg.QueryRequestOptions{ResultFormat: arg.ResultFormatObjectArray, Top: &requestQueryTop},
		subscriptions:        subscriptions,
	}
}

// QueryResources gets a query and return an array object as a result
func (client *ARGClient) QueryResources(query string) ([]interface{}, error) {
	tracer := client.tracerProvider.GetTracer("QueryResources")
	// Creates new request
	request := client.initDefaultQueryRequest(query)

	totalResults, err := client.fetchAllResults(&request)
	if err != nil {
		return nil, err
	}

	// In case that totalResults is still nil - shouldn't happen
	if totalResults == nil {
		nilError := errors.New("nil error")
		tracer.Error(nilError, "totalResults is nil - unknown behavior")
		return nil, nilError
	}

	tracer.Info("ARG query", "totalResults", len(totalResults))
	return totalResults, nil
}

// fetchAllResults from ARG using pagination. the pagination based on the skiptoken that is returned in the
// response of ARG.
func (client *ARGClient) fetchAllResults(request *arg.QueryRequest) ([]interface{}, error) {
	tracer := client.tracerProvider.GetTracer("fetchAllResults")
	// Create new totalResults array - default value is nil
	var totalResults []interface{}

	// While loop - pagination
	for totalResults == nil || request.Options.SkipToken != nil {
		// Execute query and get the response.
		response, err := client.argBaseClientWrapper.Resources(context.Background(), *request)
		if err != nil {
			return nil, errors.Wrap(err, "ARGClient.QueryResources failed on baseClient.Resources")
		}

		// Check that the response is ok
		if response.TotalRecords == nil || response.Data == nil {
			err = fmt.Errorf("ARGClient.QueryResources received ARG query response with nil TotalRecords: %v or nil Data: %v", response.Count, response.Data)
			tracer.Error(err, "")
			return nil, err
		}

		// Assert type returned is an object array correlated to options.ResultFormat(arg.ResultFormatObjectArray)
		results, ok := response.Data.([]interface{})
		if !ok {
			return nil, _errArgQueryResponseIsNotAnObjectListFormat
		}

		// In the first time, set totalResults in the length of the totalRecords. (use this instead of just appending each time for performance)
		if totalResults == nil {
			totalResults = make([]interface{}, 0, *response.TotalRecords)
		}
		// Add results to total results
		totalResults = append(totalResults, results...)

		// Update requestOptions.SkipToken in order to skip to the next page, if it's nil, it won't enter to another iteration.
		request.Options.SkipToken = response.SkipToken
	}
	return totalResults, nil
}

// initDefaultQueryRequest initialize default arg.QueryRequest.
func (client *ARGClient) initDefaultQueryRequest(query string) arg.QueryRequest {
	// Create request options - result format should be array. extracting values from client.argQueryReqOptions for preventing case of overriding default values (e.g. SkipToken)
	requestOptions := arg.QueryRequestOptions{
		ResultFormat: client.argQueryReqOptions.ResultFormat,
		Top:          client.argQueryReqOptions.Top,
	}

	// Create the query request
	request := arg.QueryRequest{
		Query:         &query,
		Options:       &requestOptions,
		Subscriptions: client.subscriptions,
	}
	return request
}
