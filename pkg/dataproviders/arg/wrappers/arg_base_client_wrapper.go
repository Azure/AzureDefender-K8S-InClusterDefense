package wrappers

import (
	"context"
	argbase "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"github.com/pkg/errors"
	"time"
)

// IARGBaseClientWrapper is a wrapper interface for base client of arg
type IARGBaseClientWrapper interface {
	// Resources is wrapping arg base client resources
	Resources(ctx context.Context, query argbase.QueryRequest) (result argbase.QueryResponse, err error)

}

type ARGBaseClientConfiguration struct {
	// Number of attempts to retrieve digest from arg
	RetryAttempts int
	// time duration between each retry
	RetryDuration time.Duration
	// time-units for the backoff duration
	TimeUnit string
}

// NewArgBaseClientWrapper get authorizer from auth.NewAuthorizerFromCLIWithResource
func (wrapper *ARGBaseClientConfiguration) NewArgBaseClientWrapper() (argbase.BaseClient, error) {
	argBaseClient := argbase.New()
	argBaseClient.RetryAttempts = wrapper.RetryAttempts
	parsedTimeUnit, err := time.ParseDuration(wrapper.TimeUnit)
	if err != nil{
		return argBaseClient, errors.Wrapf(err, "cannot parse given time unit", wrapper.TimeUnit)
	}
	argBaseClient.RetryDuration = wrapper.RetryDuration * parsedTimeUnit
	return argBaseClient, nil
}

