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

type IARGClient interface {
	QueryResources(query string) ([]interface{}, error)
}

type ARGClient struct {
	tracerProvider       trace.ITracerProvider
	metricSubmitter      metric.IMetricSubmitter
	argBaseClientWrapper wrappers.IARGBaseClientWrapper
	argReqOptions        *arg.QueryRequestOptions
}

func NewARGClient(instrumentationProvider instrumentation.IInstrumentationProvider, argBaseClientWrapper wrappers.IARGBaseClientWrapper) *ARGClient {
	return &ARGClient{
		tracerProvider:       instrumentationProvider.GetTracerProvider("ARGClient"),
		metricSubmitter:      instrumentationProvider.GetMetricSubmitter(),
		argBaseClientWrapper: argBaseClientWrapper,
		argReqOptions: &arg.QueryRequestOptions{
			ResultFormat: arg.ResultFormatObjectArray,
		},
	}
}

func (client *ARGClient) QueryResources(query string) ([]interface{}, error) {
	tracer := client.tracerProvider.GetTracer("QueryResources")

	// Create the query request
	Request := &arg.QueryRequest{
		Query:   &query,
		Options: client.argReqOptions,
	}

	tracer.Info("ARG query", "Request", Request)

	response, err := client.argBaseClientWrapper.Resources(context.Background(), *Request)
	if err != nil {
		return nil, errors.Wrap(err, "ARGClient.QueryResources failed on baseClient.Resources")
	}

	// TODO support paging with SkipToken

	tracer.Info("ARG query", "Response", response)

	// Check if response cound and data aren't null
	if response.Count == nil || response.Data == nil {
		err = fmt.Errorf("ARGClient.QueryResources received ARG query response with nil count: %v or nil data: %v", response.Count, response.Data)
		tracer.Error(err, "")
		return nil, err
	}
	// Assert type returned is an object array correlated to options.ResultFormat(arg.ResultFormatObjectArray)
	results, ok := response.Data.([]interface{})
	if ok == false {
		err = fmt.Errorf("ARGClient.QueryResources received ARG query response date is not an object list")
		tracer.Error(err, "")
		return nil, err
	}

	return results, nil
}
