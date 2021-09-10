package wrappers

import (
	"context"
	arg "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
)

// IARGBaseClientWrapper is a wrapper interface for base client of arg
type IARGBaseClientWrapper interface {
	// Resources is wrapping arg base client resources
	Resources(ctx context.Context, query arg.QueryRequest) (result arg.QueryResponse, err error)
}
