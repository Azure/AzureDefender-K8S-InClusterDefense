package wrappers

import (
	"context"
	arg "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
)

type IARGBaseClientWrapper interface{
	Resources(ctx context.Context, query arg.QueryRequest) (result arg.QueryResponse, err error)
}