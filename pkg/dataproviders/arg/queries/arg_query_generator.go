package queries

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

const (
	// _imageScanTemplateName is constant that represent the template name that will be used when creating go template.
	_imageScanTemplateName = "ImageVulnerabilityScanQuery"
)

var (
	_nilArgumentError = errors.New("Received Null Argument")
)

// ARGQueryGenerator QueryGenerator creates ARG queries from pre-defined templates
type ARGQueryGenerator struct {
	// containerVulnerabilityScanResultsQueryTemplate  is the go template of the ARG query.
	containerVulnerabilityScanResultsQueryTemplate *template.Template
	// tracerProvider
	tracerProvider trace.ITracerProvider
	// metricSubmitter
	metricSubmitter metric.IMetricSubmitter
}

// NewArgQueryGenerator Constructor
func NewArgQueryGenerator(containerVulnerabilityScanResultsQueryTemplate *template.Template, instrumentationProvider instrumentation.IInstrumentationProvider) *ARGQueryGenerator {
	return &ARGQueryGenerator{
		containerVulnerabilityScanResultsQueryTemplate: containerVulnerabilityScanResultsQueryTemplate,
		tracerProvider:  instrumentationProvider.GetTracerProvider("ArgQueryGenerator"),
		metricSubmitter: instrumentationProvider.GetMetricSubmitter(),
	}
}

// CreateARGQueryGenerator factory to create a query generator with initialized query templates
func CreateARGQueryGenerator(instrumentationProvider instrumentation.IInstrumentationProvider) (*ARGQueryGenerator, error) {
	// Parse it on create to optimize performance
	containerVulnerabilityScanResultsQueryTemplate, err := template.New(_imageScanTemplateName).Parse(_containerVulnerabilityScanResultsQueryTemplateStr)
	if err != nil {
		return nil, err
	}
	return NewArgQueryGenerator(containerVulnerabilityScanResultsQueryTemplate, instrumentationProvider), nil
}

// GenerateImageVulnerabilityScanQuery generates a parsed container image scan results query for image using provided parameters
func (generator *ARGQueryGenerator) GenerateImageVulnerabilityScanQuery(queryParameters *ContainerVulnerabilityScanResultsQueryParameters) (string, error) {
	tracer := generator.tracerProvider.GetTracer("GenerateImageVulnerabilityScanQuery")
	if queryParameters == nil {
		tracer.Error(_nilArgumentError, "queryParameters is nil")
		return "", _nilArgumentError
	}
	tracer.Info("Generate new query", "queryParameters", queryParameters)
	// Execute template using parameters
	builder := new(strings.Builder)
	err := generator.containerVulnerabilityScanResultsQueryTemplate.Execute(builder, queryParameters)
	if err != nil {
		tracer.Error(err, "Template execution failed with parameters provided")
		return "", err
	}
	return builder.String(), nil
}
