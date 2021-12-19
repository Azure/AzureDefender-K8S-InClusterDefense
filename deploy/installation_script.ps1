<#
    Installation script for In Cluster Defense - Private Preview script
    Maintainers: @ascdetectiontomer@microsoft.com>
    opensource:  https://github.com/Azure/AzureDefender-K8S-InClusterDefense
 #>

#######################################################################################################################
# Step 1: Get arguments: resourcegroup, cluster_name
Param (
    [Parameter(Mandatory = $True)]
    [string]$subscription,

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

write-host "Params that were entered:`r`nresource group : $resource_group `r`ncluster name : $cluster_name `r`n subscription: $subscription"
#######################################################################################################################
# Function for printing new section.

Function PrinitNewSection($stepTitle)
{
    write-host "########################################## Step: $stepTitle ##########################################"
}
#######################################################################################################################
#                                   Extract used variables
# Extract the region of the cluster
$region = az aks show --resource-group $resource_group --name $cluster_name --query location
if ($LASTEXITCODE -eq 3)
{
    write-error "Failed to get the region of the cluster"
    exit $LASTEXITCODE
}
# Create node resource group variable - MC_<RESOURCE_GROUP>_<CLUSTER_NAME>_<REGION>
$node_resource_group = az aks show --resource-group $resource_group --name $cluster_name --query nodeResourceGroup -o tsv
if ($LASTEXITCODE -eq 3)
{
    write-error "Failed to get the node resource group of the cluster"
    exit $LASTEXITCODE
}
# Create managed identity name
$in_cluster_defense_identity_name = "AzureDefenderInClusterDefense-$cluster_name"
# Extract kubelet identity
$kubelet_client_id = az aks show --resource-group $resource_group --name $cluster_name --query identityProfile.kubeletidentity.clientId
if ($LASTEXITCODE -eq 3)
{
    write-error "Failed to get the kubelet client id of the cluster"
    exit $LASTEXITCODE
}
# VMSS names list - we should explicity convert to array beacause if one vmss is returened, then it is returned as string.
$vmss_list = [array](az vmss list --resource-group $node_resource_group --query '[].name' -o tsv)
if ($LASTEXITCODE -eq 3 -or $vmss_list.Length -eq 0)
{
    write-error "Failed to get the list of VMSS in the node resource group"
    exit $LASTEXITCODE
}
#######################################################################################################################

PrinitNewSection("Setting account to subscription")
# login with az login.
az account set -s $subscription


#######################################################################################################################
# Step 2: Install azure addon policy in the cluster if not exists
PrinitNewSection("Azure addon policy")

<#
https://docs.microsoft.com/en-us/azure/governance/policy/concepts/policy-for-kubernetes
Before installing the Azure Policy Add-on or enabling any of the service features, your subscription must enable the Microsoft.PolicyInsights resource providers.
#>
# If flag of installation azure addon policy is false , skip this step.
if ($should_install_azure_addon_policy -eq $false)
{
    write-host "Skipping installation of Azure Addon Policy - should_install_azure_addon_policy param is false"
}
# If azure addon policy is already installed on cluster, skip this step.
elseif ((az aks show --resource-group $resource_group --name $cluster_name --query addonProfiles.azurepolicy.enabled) -eq $true)
{
    write-host "Skipping installation of Azure Addon Policy - already installed on cluster"
}
# Install Azure Addon Policy on cluster.
else
{
    write-host "Installing Azure Addon Policy"
    # Provider register: Register the Azure Policy provider
    az provider register --namespace Microsoft.PolicyInsights

    # Enable azure policy addon:
    az aks enable-addons --addons azure-policy --name $cluster_name --resource-group $resource_group

    if ($LAStEXITCODE -eq 3)
    {
        write-error "Failed to enable azure policy addon on cluster"
        exit $LASTEXITCODE
    }
}

#######################################################################################################################
PrinitNewSection("AzureDefenderInClusterDefense Dependencies")

# TODO Change the teamplate file to uri.
$deployment_name = "mdfc-incluster-$cluster_name-$region"

az deployment sub create --name  $deployment_name  --location $region `
                                                    --template-file .\deploy\azure-templates\AzureDefenderInClusterDefense.Dependecies.Template.json `
                                                    --parameters `
                                                        resource_group=$node_resource_group `
                                                        location=$region `
                                                        managedIdentityName=$in_cluster_defense_identity_name

if ($LASTEXITCODE -eq 3)
{
    write-error "Failed to create AzureDefenderInClusterDefense dependencies"
    exit $LASTEXITCODE
}
# #####################################################################################################################
PrinitNewSection("azure-defender-k8s-security-profile Dependencies")

if ($should_enable_aks_security_profile -eq $false)
{
    write-host "Skipping on enabling azure-defender-k8s-security-profile - should_install_enable_aks_security_profile param is false"
}
elseif ((az aks show --resource-group $resource_group --name $cluster_name --query securityProfile.azureDefender.enabled) -eq $true)
{
    write-host "Skipping on enabling of azure-defender-k8s-security-profile - already enabled on cluster"
}
else
{
    ## Enable AKS-AzureDefender Feature flag on your subscription
    write-host "Enabling of azure-defender-k8s-security-profile"

    $url = "https://management.azure.com/subscriptions/$subscription/providers/Microsoft.Features/providers/Microsoft.ContainerService/features/AKS-AzureDefender/register?api-version=2015-12-01";
    $token = (az account get-access-token --query "accessToken" -o tsv)

    if ($LASTEXITCODE -eq 3 -or $token -eq "")
    {
        write-error "Failed to get access token"
        exit $LASTEXITCODE
    }

    $authorization_header = @{
        Authorization = "Bearer $token"
    }
    $response = Invoke-WebRequest -Method POST -Uri $url -Headers $authorization_header -UseBasicParsing

    if ($LASTEXITCODE -eq 3 -or $response.StatusCode -ne 200)
    {
        write-error "Failed to enable AKS-AzureDefender feature flag"
        exit $LASTEXITCODE
    }

    # Deploy arm template - containts the dependencies of azure-defender-k8s-security-profile
    # TODO Change the teamplate file to uri.
    $deployment_name = "mdfc-profile-$cluster_name-$region"
    az deployment sub create --name $deployment_name    --location "$region" `
                                                        --template-file .\deploy\azure-templates\Tivan.Dependencies.Template.json `
                                                        --parameters `
                                                            subscriptionId=$subscription `
                                                            clusterName=$cluster_name `
                                                            clusterResourceGroup=$resource_group `
                                                            resourceLocation=$region

    if ($LASTEXITCODE -eq 3)
    {
        write-error "Failed to create azure-defender-k8s-security-profile dependencies"
        exit $LASTEXITCODE
    }
}
#######################################################################################################################
PrinitNewSection("Attach identity to VMSS on node resource group")

For($i = 0; $i -lt $vmss_list.Length; $i++){
    write-host "Assigning identity to vmss <$vmss_list[$i]>"
    az vmss identity assign --resource-group $node_resource_group --name $vmss_list[$i] --identities $in_cluster_defense_identity_name

    if ($LASTEXITCODE -eq 3)
    {
        write-error "Failed to attach identity to vmss <$vmss_list[$i]>"
        exit $LASTEXITCODE
    }
}
#######################################################################################################################
PrinitNewSection("Installing Helm Chart")

# Step 6: Install helm chart

$in_cluster_defense_identity_client_id = az identity show -n $in_cluster_defense_identity_name -g $node_resource_group --query clientId

if ($LASTEXITCODE -eq 3 -or $in_cluster_defense_identity_client_id -eq "")
{
    write-error "Failed to get client id of in-cluster-defense identity"
    exit $LASTEXITCODE
}

# Switch to current context
kubectl config use-context $cluster_name

# Install helm chart
$HELM_EXPERIMENTAL_OCI = 1

# Install helm chart from mcr repo on kube-system namespace and pass subscription and client id's params.

# TODO Change to remote repo once helm chart is published to public repo.
#helm upgrade --install microsoft-in-cluster-defense azuredefendermcrprod.azurecr.io/public/azuredefender/stable/in-cluster-defense-helm:$helm_chart_version `
helm upgrade in-cluster-defense charts/azdproxy --install --wait `
            -n kube-system `
                --set AzDProxy.kubeletIdentity.envAzureAuthorizerConfiguration.mSIClientId=$kubelet_client_id `
                --set AzDProxy.azdIdentity.envAzureAuthorizerConfiguration.mSIClientId=$in_cluster_defense_identity_client_id `
                --set "AzDProxy.arg.argClientConfiguration.subscriptions={$subscription}" `
                --set AzDProxy.webhook.image.name=blockregistrydev.azurecr.io/azdproxy-image                 # TODO Delete above line once helm chart is published to public repo.
