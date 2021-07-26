package azureauth

import (
	"context"
	"fmt"
	"testing"

	arg "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
)

func TestNewARMAuthorizer(t *testing.T) {
	factory := AzureAuthroizerFactory{
		configuration: &AzureAuthroizerConfiguration{
			isLocalDevelopmentMode: true,
			clientId:               "27f25a78-b423-4756-9bf8-151d44d46c12",
		},
	}

	authorizer, err := factory.NewARMAuthorizer()
	if err != nil {
		fmt.Printf("Got Error %q", err)
		fmt.Println()
		return
	}

	argClient := arg.New()
	argClient.Authorizer = authorizer

	RequestOptions := arg.QueryRequestOptions{
		ResultFormat: "objectArray",
	}

	query := "securityresources | take 1"
	// Create the query request
	Request := arg.QueryRequest{
		Query:   &query,
		Options: &RequestOptions,
	}

	results, err := argClient.Resources(context.Background(), Request)
	if err != nil {
		fmt.Printf("Got Error %q", err)
		fmt.Println()

	}

	fmt.Printf("Results %q", results.Data)
	fmt.Println()

}
