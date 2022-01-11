package queries

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/metric/util"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"strings"
	"text/template"
)

const (
	// _imageScanTemplateName is constant that represent the template name that will be used when creating go template.
	_imageScanTemplateName = "ImageVulnerabilityScanQuery"
)

type IARGQueryGenerator interface {
	// GenerateImageVulnerabilityScanQuery generates a parsed container image scan results query for image using provided parameters
	GenerateImageVulnerabilityScanQuery(queryParameters *ContainerVulnerabilityScanResultsQueryParameters) (string, error)
}

var _ IARGQueryGenerator = &ARGQueryGenerator{}

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
		tracer.Error(utils.NilArgumentError, "queryParameters is nil")
		generator.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(utils.NilArgumentError, "ARGQueryGenerator.GenerateImageVulnerabilityScanQuery"))
		return "", utils.NilArgumentError
	}
	tracer.Info("Generate new query", "queryParameters", *queryParameters)
	// Execute template using parameters
	builder := new(strings.Builder)
	err := generator.containerVulnerabilityScanResultsQueryTemplate.Execute(builder, queryParameters)
	if err != nil {
		generator.metricSubmitter.SendMetric(1, util.NewErrorEncounteredMetric(err, "ARGQueryGenerator.GenerateImageVulnerabilityScanQuery"))
		tracer.Error(err, "Template execution failed with parameters provided")
		return "", err
	}
	return builder.String(), nil
}
