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

    [Parameter(Mandatory = $True)]
    [string]$region
)

write-host "Params that were entered:`r`nresource group : $resource_group `r`ncluster name : $cluster_name,`r`nregion : $region"
#######################################################################################################################
# Function for printing new section.
$stepCount = 1
Function PrinitNewSection($stepTitle)
{
    write-host "########################################## Step: $stepCount - $stepTitle ##########################################"
    $stepCount++
}
#######################################################################################################################
# Create node resource group variable - MC_<RESOURCE_GROUP>_<CLUSTER_NAME>_<REGION>
$node_resource_group = "MC_$resource_group`_$cluster_name`_$region"

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

# Provider register: Register the Azure Policy provider
az provider register --namespace Microsoft.PolicyInsights

# Enable azure policy addon:
az aks enable-addons --addons azure-policy --name $cluster_name --resource-group $resource_group

#######################################################################################################################
PrinitNewSection("AzureDefenderInClusterDefense Dependencies")

# TODO Change the teamplate file to uri.
az deployment sub create --name "AzureDefenderInClusterDefense-dependencies" --location "$region" --template-file .\deploy\Deployment\ServiceGroupRoot\Templates\AzureDefenderInClusterDefense.Dependecies.Template.json --parameters cluster_name = $cluster_name resource_group = $node_resource_group location = $region

#######################################################################################################################
PrinitNewSection("azure-defender-k8s-security-profile Dependencies")
## Enable AKS-AzureDefender Feature flag on your subscription
$url = "https://management.azure.com/subscriptions/$subscription/providers/Microsoft.Features/providers/Microsoft.ContainerService/features/AKS-AzureDefender/register?api-version=2015-12-01";
$token = (az account get-access-token --query "accessToken" -o tsv)
$authorization_header = @{
    Authorization = "Bearer $token"
}
Invoke-WebRequest -Method POST -Uri $url -Headers $authorization_header

# Deploy arm template - containts the dependencies of azure-defender-k8s-security-profile
# TODO Change the teamplate file to uri.
az deployment sub create --name "azure-defender-k8s-security-profile" --location "$region" --template-file .\deploy\Deployment\ServiceGroupRoot\Templates\Tivan.Dependencies.Template.json --parameters subscriptionId = $subscription clusterName = $cluster_name clusterResourceGroup = $resource_group resourceLocation = $region

#######################################################################################################################
PrinitNewSection("Installing Helm Chart")

<#
Step 6: Get all helm values for installing helm chart -
    - subscription id
    - agentpool identity
    - AzureDefenderInClusterDefense client id of MI
Step 6: Install / upgrade helm chart with helm values
 #>

$kubelet_client_id = az identity show -n "$cluster_name-agentpool" -g $node_resource_group --query "clientId"
$in_cluster_defense_identity_client_id = az identity show -n "AzureDefenderInClusterDefense-$cluster_name" -g $node_resource_group --query "clientId"

write-host "kubelet_client_id: <$kubelet_client_id>"
write-host "in_cluster_defense_identity_client_id: <$in_cluster_defense_identity_client_id>"

# Switch to current context
kubectl config use-context $cluster_name

# Install helm chart
$HELM_EXPERIMENTAL_OCI = 1

# Install helm chart from mcr repo on kube-system namespace and pass subscription and client id's params.
helm install azuredefender azuredefendermcrprod.azurecr.io/public/azuredefender/stable/in-cluster-defense-helm -n kube-system --set AzDProxy.kubeletIdentity.envAzureAuthorizerConfiguration.mSIClientId = $kubelet_client_id --set AzDProxy.azdIdentity.envAzureAuthorizerConfiguration.mSIClientId = $in_cluster_defense_identity_client_id --set "AzDProxy.arg.argClientConfiguration.subscriptions={$subscription}"