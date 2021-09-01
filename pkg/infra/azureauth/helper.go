package azureauth

import (
	"context"
	"fmt"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/pkg/errors"
	"net/http"
)

type IBearerAuthorizer interface {
	TokenProvider() adal.OAuthTokenProvider
}

func RefreshBearerAuthorizer(bearerAuthorizer IBearerAuthorizer, ctx context.Context) error {
	var err error  = nil

	// the ordering is important here, prefer RefresherWithContext if available
	if refresher, ok := bearerAuthorizer.TokenProvider().(adal.RefresherWithContext); ok {
		err = refresher.EnsureFreshWithContext(ctx)
	} else if refresher, ok := bearerAuthorizer.TokenProvider().(adal.Refresher); ok {
		err = refresher.EnsureFresh()
	}
	if err != nil {
		var resp *http.Response
		if tokError, ok := err.(adal.TokenRefreshError); ok {
			resp = tokError.Response()
		}
		return errors.Wrap(err, fmt.Sprint("azure.BearerAuthorizer", "WithAuthorization", resp, "Failed to refresh the Token"))
	}

	return nil
}
