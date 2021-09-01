package azure

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/azureauth/wrappers"
	"testing"
)

func TestACRTokenProvider_GetACRTokenFromARMToken(t *testing.T) {
	authorizerFactory := azureauth.NewEnvAzureAuthorizerFactory(&azureauth.EnvAzureAuthorizerConfiguration{
		IsLocalDevelopmentMode: true,
	}, new(wrappers.AzureAuthWrapper))

	authorizer, err := authorizerFactory.CreateARMAuthorizer()
	if err != nil{
		t.Error(err)
	}
	bearer,ok := authorizer.(azureauth.IBearerAuthorizer)
	if !ok{
		t.Error()
	}
	a := NewACRTokenProvider(nil, bearer)
	_,_ = a.GetACRTokenFromARMToken("tomerwdevopsstage.azurecr.io")
}
