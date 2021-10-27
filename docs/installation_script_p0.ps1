<#
    Installation script for In Cluster Defense - Private Preview script
    Maintainers: @ascdetectiontomer@microsoft.com>
    opensource:  https://github.com/Azure/AzureDefender-K8S-InClusterDefense
 #>

#######################################################################################################################
# Step 1: Get arguments: resourcegroup, cluster_name
# df
Param (
    [Parameter(Mandatory = $True)]
    [string]$resource_group,

    [Parameter(Mandatory = $True)]
    [string]$cluster_name,

    [Parameter(Mandatory = $True)]
    [string]$region
)

# Create node resource group variable - MC_<RESOURCE_GROUP>_<CLUSTER_NAME>_<REGION>
$node_resource_group = "MC_$resource_group`_$cluster_name`_$region"

write-host "Params that were entered:`r`nresource group : $resource_group `r`ncluster name : $cluster_name,`r`nregion : $region"

#######################################################################################################################
# login with az login.
# az login

write-host "extracting subscription"
$subscription = "TODO"
#######################################################################################################################
# Step 2: Install azure addon policy in the cluster if not exists

<# https://docs.microsoft.com/en-us/azure/governance/policy/concepts/policy-for-kubernetes
Before installing the Azure Policy Add-on or enabling any of the service features, your subscription must enable the Microsoft.PolicyInsights resource providers.
#>
# Provider register: Register the Azure Policy provider
# az provider register --namespace Microsoft.PolicyInsights
# The AKS cluster must be version 1.14 or higher. Use the following script to validate your AKS cluster version:
# # Look for the value in kubernetesVersion
# az aks list
# az aks enable-addons --addons azure-policy --name MyAKSCluster --resource-group MyResourceGroup

#######################################################################################################################
# Step 3: Create block identity (User managed identity) if not exists.
write-host "Checking if there is already idenity in $node_resource_group resource group"
$in_cluster_defense_identity_name = "$cluster_name`-in-cluster-defense"
$data_of_id = az identity show -n $in_cluster_defense_identity_name -g $node_resource_group | ConvertFrom-Json
if ($LASTEXITCODE -eq 3)
{
    write-host "idenity is not exist... creating new MI with name $in_cluster_defense_identity_name at resource group $resource_group"
    # create new identity
    $data_of_id = az identity create --name $in_cluster_defense_identity_name -g $node_resource_group | ConvertFrom-Json
    write-host "Created new MI successfully"
}
else
{
    write-host "idenity is already exists"
}
# Extract client id into variable - it will be passed to helm.
$in_cluster_defense_identity_client_id = $data_of_id.clientId
write-host "The client id that will be passed to helm of In Cluster Defense MI is: $in_cluster_defense_identity_client_id"
#######################################################################################################################
# Step 4: Create RBAC of subscription reader (RBAC of block identity) if not exists
write-host "Checking if there is already subscription($subscription) reader RBAC for $in_cluster_defense_identity_client_id"

#######################################################################################################################
# Step 5 ?: install Tivan if not exists (for publisher ?)

#######################################################################################################################
<# Step 6: Get all helm values for installing helm chart -
    - agentpool identity
    - subscription Identity
    - block identity
 #>

#######################################################################################################################
# Step 6: Install / upgrade helm chart with helm values