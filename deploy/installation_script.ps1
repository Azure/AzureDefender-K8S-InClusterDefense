<#
    Installation script for In Cluster Defense - Private Preview script
    Maintainers: @ascdetectiontomer@microsoft.com>
    opensource:  https://github.com/Azure/AzureDefender-K8S-InClusterDefense
 #>

#######################################################################################################################
# Step 1: Get arguments: resourcegroup, cluster_name
Param (
    [Parameter(Mandatory = $True)]
    [string]$resource_group,

    [Parameter(Mandatory = $True)]
    [string]$cluster_name,

    [Parameter()]
    [string]$helm_chart_version = "1.0.0",

    [Parameter()]
    [bool]$should_install_azure_addon_policy = $true,

    [Parameter()]
    [bool]$should_enable_aks_security_profile = $true
)

write-host "Params that were entered:`r`nresource group : $resource_group `r`ncluster name : $cluster_name"
#######################################################################################################################
# Function for printing new section.
$stepCount = 1
Function PrinitNewSection($stepTitle)
{
    write-host "########################################## Step: $stepCount - $stepTitle ##########################################"
    $stepCount++
}
#######################################################################################################################
#                                   Extract used variables
# Extract the region of the cluster
$region = az aks show --resource-group $resource_group --name $cluster_name --query location
# Create node resource group variable - MC_<RESOURCE_GROUP>_<CLUSTER_NAME>_<REGION>
$node_resource_group = az aks show --resource-group $resource_group --name $cluster_name --query nodeResourceGroup
# Create managed identity name
$in_cluster_defense_identity_name = "AzureDefenderInClusterDefense-$cluster_name"
# Extract kubelet identity
$kubelet_client_id = az aks show --resource-group $resource_group --name $cluster_name --query identityProfile.kubeletidentity.clientId
#######################################################################################################################

PrinitNewSection("Logging in")
# login with az login.
az login

$subscription = az account show -o tsv --query "id"
write-host "Extracted subscription <$subscription> successfully"

#######################################################################################################################
# Step 2: Install azure addon policy in the cluster if not exists
PrinitNewSection("Azure addon policy")

<#
https://docs.microsoft.com/en-us/azure/governance/policy/concepts/policy-for-kubernetes
Before installing the Azure Policy Add-on or enabling any of the service features, your subscription must enable the Microsoft.PolicyInsights resource providers.
#>
if ($should_install_azure_addon_policy)
{
    write-host "Installing Azure Addon Policy"
    # Provider register: Register the Azure Policy provider
    az provider register --namespace Microsoft.PolicyInsights

    # Enable azure policy addon:
    az aks enable-addons --addons azure-policy --name $cluster_name --resource-group $resource_group
}
else
{
    write-host "Skipping installation of Azure Addon Policy - should_install_azure_addon_policy param is false"
}


#######################################################################################################################
PrinitNewSection("AzureDefenderInClusterDefense Dependencies")

# TODO Change the teamplate file to uri.
$deployment_name = "AzureDefenderInClusterDefense-dependencies-$cluster_name-$resource_group-$region"
az deployment sub create --name  $deployment_name  --location $region `
                                                    --template-file .\deploy\azure-templates\AzureDefenderInClusterDefense.Dependecies.Template.json `
                                                    --parameters `
                                                        resource_group = $node_resource_group `
                                                        location = $region `
                                                        managedIdentityName = $in_cluster_defense_identity_name

#######################################################################################################################
PrinitNewSection("azure-defender-k8s-security-profile Dependencies")

if ($should_enable_aks_security_profile)
{
    ## Enable AKS-AzureDefender Feature flag on your subscription
    $url = "https://management.azure.com/subscriptions/$subscription/providers/Microsoft.Features/providers/Microsoft.ContainerService/features/AKS-AzureDefender/register?api-version=2015-12-01";
    $token = (az account get-access-token --query "accessToken" -o tsv)
    $authorization_header = @{
        Authorization = "Bearer $token"
    }
    Invoke-WebRequest -Method POST -Uri $url -Headers $authorization_header

    # Deploy arm template - containts the dependencies of azure-defender-k8s-security-profile
    # TODO Change the teamplate file to uri.
    $deployment_name = "azure-defender-k8s-security-profile-$cluster_name-$resource_group-$region"
    az deployment sub create --name $deployment_name    --location "$region" `
                                                        --template-file .\deploy\azure-templates\Tivan.Dependencies.Template.json `
                                                        --parameters `
                                                            subscriptionId = $subscription `
                                                            clusterName = $cluster_name `
                                                            clusterResourceGroup = $resource_group `
                                                            resourceLocation = $region

}
else
{
    write-host "Skipping on enabling of azure-defender-k8s-security-profile - should_install_enable_aks_security_profile param is false"
}

#######################################################################################################################
PrinitNewSection("Installing Helm Chart")

# Step 6: Install helm chart

$in_cluster_defense_identity_client_id = az identity show -n $in_cluster_defense_identity_name -g $node_resource_group --query clientId

# Switch to current context
kubectl config use-context $cluster_name

# Install helm chart
$HELM_EXPERIMENTAL_OCI = 1

# Install helm chart from mcr repo on kube-system namespace and pass subscription and client id's params.
helm install azuredefender azuredefendermcrprod.azurecr.io/public/azuredefender/stable/in-cluster-defense-helm:$helm_chart_version `
            -n kube-system `
                --set AzDProxy.kubeletIdentity.envAzureAuthorizerConfiguration.mSIClientId = $kubelet_client_id `
                --set AzDProxy.azdIdentity.envAzureAuthorizerConfiguration.mSIClientId = $in_cluster_defense_identity_client_id `
                --set "AzDProxy.arg.argClientConfiguration.subscriptions={$subscription}"