package wrappers

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	argbase "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"github.com/Azure/go-autorest/autorest"
)

// arg.BaseClient implements IARGBaseClientWrapper interface
var _ IARGBaseClientWrapper = (*argbase.BaseClient)(nil)

// IARGBaseClientWrapper is a wrapper interface for base client of arg
type IARGBaseClientWrapper interface {
	// Resources is wrapping arg base client resources
	Resources(ctx context.Context, query argbase.QueryRequest) (result argbase.QueryResponse, err error)
}

// NewArgBaseClientWrapper get authorizer from auth.NewAuthorizerFromCLIWithResource
func NewArgBaseClientWrapper(retryPolicyConfig *retrypolicy.RetryPolicyConfiguration, authorizer autorest.Authorizer) (*argbase.BaseClient, error) {
	// Create new client
	argBaseClient := argbase.New()
	// Assign the retry policy configuration to the client.
	argBaseClient.RetryAttempts = retryPolicyConfig.RetryAttempts
	retryDuration := retryPolicyConfig.GetBackOffDuration()
	argBaseClient.RetryDuration = retryDuration

	// Assign the authorizer to the client.
	argBaseClient.Authorizer = authorizer

	return &argBaseClient, nil
}
