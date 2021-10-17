package azureauth

import (
	"context"
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/pkg/errors"
	"net/http"
)

// IBearerAuthorizer is a bearer token based authorizer
type IBearerAuthorizer interface {
	// TokenProvider is the provider for getting a token
	TokenProvider() adal.OAuthTokenProvider
}

var _ IBearerAuthorizer = &autorest.BearerAuthorizer{}

type IBearerAuthorizerTokenProvider interface{

	GetOAuthToken(ctx context.Context) (string, error)
}

var _ IBearerAuthorizerTokenProvider = &BearerAuthorizerTokenProvider{}

// TODO add testing for this struct..(How?)
type BearerAuthorizerTokenProvider struct {
	bearerAuthorizer IBearerAuthorizer
}


func NewBearerAuthorizerTokenProvider(bearerAuthorizer IBearerAuthorizer) *BearerAuthorizerTokenProvider {
	return &BearerAuthorizerTokenProvider{bearerAuthorizer: bearerAuthorizer}
}

func (provider *BearerAuthorizerTokenProvider) GetOAuthToken(ctx context.Context) (string, error) {
	err := provider.refreshBearerAuthorizer(ctx)
	if err != nil{
		return "", errors.Wrap(err, "BearerAuthorizerTokenProvider.GetOAuthToken")
	}

	token :=  provider.bearerAuthorizer.TokenProvider().OAuthToken()
	return token, nil
}

// RefreshBearerAuthorizer receives a bearer authorizer and check try to refresh it with context provided
// Taken from azure auth bearer token refresh logic
// First try to refresh with ctx otherwise fallback to non ctx refresh
func (provider *BearerAuthorizerTokenProvider) refreshBearerAuthorizer(ctx context.Context) error {
	var err error  = nil

	if provider.bearerAuthorizer == nil {
		return errors.Wrap(utils.NilArgumentError, "BearerAuthorizerTokenProvider.RefreshBearerAuthorizer")
	}

	// TODO add retry policy?


	// the ordering is important here, prefer RefresherWithContext if available
	if refresher, ok := provider.bearerAuthorizer.TokenProvider().(adal.RefresherWithContext); ok {
		err = refresher.EnsureFreshWithContext(ctx)
	} else if refresher, ok := provider.bearerAuthorizer.TokenProvider().(adal.Refresher); ok {
		err = refresher.EnsureFresh()
	}
	if err != nil {
		var resp *http.Response
		if tokError, ok := err.(adal.TokenRefreshError); ok {
			resp = tokError.Response()
		}
		return errors.Wrap(err, fmt.Sprint("azure.BearerAuthorizer", "RefreshBearerAuthorizer", resp, "Failed to refresh the Token"))
	}

	return nil
}
