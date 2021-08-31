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

// IARGClient is an interface for our arg client implemntation
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
}

// NewARGClient Constructor
func NewARGClient(instrumentationProvider instrumentation.IInstrumentationProvider, argBaseClientWrapper wrappers.IARGBaseClientWrapper) *ARGClient {
	// We need this var for unittests - in unittests we reduce it from 1000 to smaller number.
	requestQueryTop := int32(MAX_TOP_RESULTS_IN_PAGE_OF_ARG)

	return &ARGClient{
		tracerProvider:       instrumentationProvider.GetTracerProvider("ARGClient"),
		metricSubmitter:      instrumentationProvider.GetMetricSubmitter(),
		argBaseClientWrapper: argBaseClientWrapper,
		argQueryReqOptions:   &arg.QueryRequestOptions{ResultFormat: arg.ResultFormatObjectArray, Top: &requestQueryTop},
	}
}

// QueryResources gets a query and return an array object as a result
func (client *ARGClient) QueryResources(query string) ([]interface{}, error) {
	tracer := client.tracerProvider.GetTracer("QueryResources")
	// Create new totalResults array - default value is nil
	var totalResults []interface{}
	// Create request options - result format should be array. extracting values from client.argQueryReqOptions for preventing case of overriding default values (e.g. SkipToken)
	requestOptions := arg.QueryRequestOptions{
		ResultFormat: client.argQueryReqOptions.ResultFormat,
		Top:          client.argQueryReqOptions.Top,
	}

	// Create the query request
	request := arg.QueryRequest{
		Query:   &query,
		Options: &requestOptions,
		//TODO Add subscriptions?
	}

	// While loop - pagination
	for {
		// Execute query and get the response.
		response, err := client.argBaseClientWrapper.Resources(context.Background(), request)
		if err != nil {
			return nil, errors.Wrap(err, "ARGClient.QueryResources failed on baseClient.Resources")
		}

		// Check that the response is ok
		if response.TotalRecords == nil || response.Data == nil {
			err = fmt.Errorf("ARGClient.QueryResources received ARG query response with nil TotalRecords: %v or nil Data: %v", response.Count, response.Data)
			tracer.Error(err, "")
			return nil, err
		}

		// In the first time, set totalResults in the length of the totalRecords. (use this instead of just appending each time for performance)
		if totalResults == nil {
			totalResults = make([]interface{}, 0, *response.TotalRecords)
		}

		// Assert type returned is an object array correlated to options.ResultFormat(arg.ResultFormatObjectArray)
		results, ok := response.Data.([]interface{})
		if !ok {
			return nil, _errArgQueryResponseIsNotAnObjectListFormat
		}

		// Check if we got empty data.
		if len(results) == 0 {
			break
		}

		// Add results to total results
		totalResults = append(totalResults, results...)

		// pagination - if response.SkipToken is  null, we fetched all data.
		if response.SkipToken == nil {
			break
		}
		// Update requestOptions.SkipToken in order to skip to the next page.
		requestOptions.SkipToken = response.SkipToken
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
