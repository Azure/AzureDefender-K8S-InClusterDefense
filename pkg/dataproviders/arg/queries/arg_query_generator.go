package queries

import (
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

const (
	_imageScanTemplateName = "ImageVulnerabilityScanQuery"
)

var(
	_nilArgumentError = errors.New("Received Null Argument")
)

// QueryGenerator creates ARG queries from pre-defined templates
type ARGQueryGenerator struct {
	containerVulnerabilityScanResultsQueryTemplate *template.Template
}

// Constructor
func newQueryGenerator(containerVulnerabilityScanResultsQueryTemplate *template.Template) *ARGQueryGenerator {
	return &ARGQueryGenerator{
		containerVulnerabilityScanResultsQueryTemplate: containerVulnerabilityScanResultsQueryTemplate,
	}
}

// CreateARGQueryGenerator factory to create a query generator with initalized query templates
func CreateARGQueryGenerator() (*ARGQueryGenerator, error) {
	// Parse it on create to optimize performance
	containerVulnerabilityScanResultsQueryTemplate, err := template.New(_imageScanTemplateName).Parse(_containerVulnerabilityScanResultsQueryTemplateStr)
	if err != nil {
		return nil, err
	}
	return newQueryGenerator(containerVulnerabilityScanResultsQueryTemplate), nil
}

// GenerateImageVulnerabilityScanQuery generates a parsed container image scan results query for image using provided parameters
func (generator *ARGQueryGenerator) GenerateImageVulnerabilityScanQuery(queryParameters *ContainerVulnerabilityScanResultsQueryParameters) (string, error) {
	if queryParameters == nil {
		return "", _nilArgumentError
	}

	// Execute template using paramters
	builder := new(strings.Builder)
	err := generator.containerVulnerabilityScanResultsQueryTemplate.Execute(builder, queryParameters)
	if err != nil {
		// Template execuition failed with paramaters provided
		return "", err
	}
	return builder.String(), nil
}
