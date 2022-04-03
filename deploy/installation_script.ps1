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
    [string]$helm_chart_version = "0.1.0",

    [Parameter()]
    [bool]$should_install_azure_addon_policy = $true,

    [Parameter()]
    [bool]$should_enable_aks_security_profile = $true,
	
	[Parameter()]
    [bool]$should_install_inclusterdefense_dependencies = $true,
	
	[Parameter()]
    [bool]$should_install_inclusterdefense_vmss_assign_identities = $true
)

write-host "Params that were entered:`r`nsubscription: $subscription `r`nresource group : $resource_group `r`ncluster name : $cluster_name"
#######################################################################################################################
# Function for printing new section.

Function PrintNewSection($stepTitle)
{
    write-host "########################################## Step: $stepTitle ##########################################"
}

# Function for checking if an extension exists.
function Check-Command($extensionName)
{
    return [bool](Get-Command -Name $extensionName)
}
#######################################################################################################################
PrintNewSection("Checking Prerequisite")

write-host "Checking if azure-cli is installed"
if (Check-Command -extensionName 'az'){
    Write-Host "azure-cli is installed"
}else{
    Write-error "Did not find azure-cli installed, please make sure you install it and add it to the PATH variables. For more information https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
}

write-host "Checking if helm is installed"
if (Check-Command -extensionName 'helm'){
    Write-Host "Helm is installed"
}else{
    Write-error "Did not find helm installed, please make sure you install it and add it to the PATH variables (recommended version: helm v3.6.3). For more information https://helm.sh/docs/intro/install/"
    exit 1
}

write-host "Checking if kubectl is installed"
if (Check-Command -extensionName 'kubectl'){
    Write-Host "kubectl is installed"
}else{
    Write-error "Did not find kubectl installed, please make sure you install it and add it to the PATH variables. For more information https://kubernetes.io/docs/tasks/tools/#kubectl"
    exit 1
}

#######################################################################################################################

PrintNewSection("Setting account to subscription")
# login with az login.

az account set -s $subscription
if ($LASTEXITCODE -gt 0)
{
    write-error "Subscription doesn't exit or wrong permissions. Try 'az login --use-device-code' and rerun the script"
    exit 1
}

# Extract used variables
# Extract the region of the cluster
$region = az aks show --resource-group $resource_group --name $cluster_name --query location -o tsv
$azureResourceID = az aks show --resource-group $resource_group --name $cluster_name --query id -o tsv

if ($LASTEXITCODE -ge 1)
{
    write-error "Failed to get the region of the cluster"
    exit $LASTEXITCODE
}
# Create node resource group variable - MC_<RESOURCE_GROUP>_<CLUSTER_NAME>_<REGION>
$node_resource_group = az aks show --resource-group $resource_group --name $cluster_name --query nodeResourceGroup -o tsv
if ($LASTEXITCODE -ge 1)
{
    write-error "Failed to get the node resource group of the cluster"
    exit $LASTEXITCODE
}
# Create managed identity name
$in_cluster_defense_identity_name = "AzureDefenderInClusterDefense-$cluster_name"
# Extract kubelet identity
$kubelet_client_id = az aks show --resource-group $resource_group --name $cluster_name --query identityProfile.kubeletidentity.clientId
if ($LASTEXITCODE -ge 1 -or $kubelet_client_id -eq $null)
{
    write-error "Failed to get the kubelet client id of the cluster, currently only MSI cluster identity is supported and not SPN"
    exit $LASTEXITCODE
}
# VMSS names list - we should explicity convert to array beacause if one vmss is returened, then it is returned as string.
$vmss_list = [array](az vmss list --resource-group $node_resource_group --query '[].name' -o tsv)
if ($LASTEXITCODE -ge 1 -or $vmss_list.Length -eq 0)
{
    write-error "Failed to get the list of VMSS in the node resource group"
    exit $LASTEXITCODE
}

write-host "Set account to subscription successfully"

#######################################################################################################################
# Step 2: Install azure addon policy in the cluster if not exists
PrintNewSection("Azure addon policy")

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
    $_ = az aks enable-addons --addons azure-policy --name $cluster_name --resource-group $resource_group

    if ($LASTEXITCODE -ge 1)
    {
        write-error "Failed to enable azure policy addon on cluster"
        exit $LASTEXITCODE
    }
}

#######################################################################################################################
PrintNewSection("AzureDefenderInClusterDefense Dependencies")
# deployment_name must be shorter than 62 letters
$deployment_name = "mdfc-incluster-$cluster_name-$region"

$isDepExists = $false
if ($should_install_inclusterdefense_dependencies -ne $false)
{
    # will get the same deployment name but will have different node resource group name).
    $is_in_cluster_defense_identity_exists = $in_cluster_defense_identity_name -in $(az identity list -g $node_resource_group --query [*].name -o json | ConvertFrom-Json)
    if (($LASTEXITCODE -eq 0) -and ($is_in_cluster_defense_identity_exists))
    {
        $objectId = az identity show -g $node_resource_group -n $in_cluster_defense_identity_name --query principalId -o tsv
        $assignmentExist = $((az role assignment list --assignee $objectId --query "[?roleDefinitionName=='Reader' && scope=='/subscriptions/$subscription']" -o json | ConvertFrom-Json).Length -ge 1)
        if ($assignmentExist)
        {
            write-host "Skipping installation of InClusterDefense Dependencies - already exist, setting 'isDepExists' to false"
            $isDepExists = $true
        }
    }
}

if (($should_install_inclusterdefense_dependencies -eq $false) -or ($isDepExists -eq $true))
{
   write-host "Skipping installation of InClusterDefense Dependencies - should_install_inclusterdefense_dependencies param is false or role assignment exist"
}
else
{
	# TODO Change the teamplate file to uri.
	$_ = az deployment sub create --name  $deployment_name  --location $region `
														--template-file .\azure-templates\AzureDefenderInClusterDefense.Dependecies.Template.json `
														--parameters `
															resource_group=$node_resource_group `
															location=$region `
                                                            managedIdentityName=$in_cluster_defense_identity_name
    if ($LASTEXITCODE -ge 1)
    {
        write-error "Failed to create AzureDefenderInClusterDefense dependencies"
        exit $LASTEXITCODE
    }
    write-host "Created AzureDefenderInClusterDefense dependencies successfully"
}

# #####################################################################################################################
PrintNewSection("azure-defender-k8s-security-profile Dependencies")

if ($should_enable_aks_security_profile -eq $false)
{
    write-host "Skipping on enabling azure-defender-k8s-security-profile - should_enable_aks_security_profile param is false"
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

    if ($LASTEXITCODE -ge 1 -or $token -eq "")
    {
        write-error "Failed to get access token"
        exit $LASTEXITCODE
    }

    $authorization_header = @{
        Authorization = "Bearer $token"
    }
    $response = Invoke-WebRequest -Method POST -Uri $url -Headers $authorization_header -UseBasicParsing

    if ($LASTEXITCODE -ge 1 -or $response.StatusCode -ne 200)
    {
        write-error "Failed to enable AKS-AzureDefender feature flag"
        exit $LASTEXITCODE
    }

    # Deploy arm template - containts the dependencies of azure-defender-k8s-security-profile
    # TODO Change the teamplate file to uri.
    $deployment_name = "mdfc-profile-$cluster_name-$region"
    $_ = az deployment sub create --name $deployment_name    --location "$region" `
                                                        --template-file .\azure-templates\Tivan.Dependencies.Template.json `
                                                        --parameters `
                                                            subscriptionId=$subscription `
                                                            clusterName=$cluster_name `
                                                            clusterResourceGroup=$resource_group `
                                                            resourceLocation=$region

    if ($LASTEXITCODE -ge 1)
    {
        write-error "Failed to create azure-defender-k8s-security-profile dependencies"
        exit $LASTEXITCODE
    }
    write-host "created azure-defender-k8s-security-profile dependencies successfully"
}
#######################################################################################################################
PrintNewSection("Attach identity to VMSS on node resource group")
if ($should_install_inclusterdefense_vmss_assign_identities -eq $false)
{
    write-host "Skipping on enabling assign vmss identities - should_install_inclusterdefense_vmss_assign_identities param is false"
}
else
{
	For($i = 0; $i -lt $vmss_list.Length; $i++){
		write-host "Assigning identity to vmss <$vmss_list[$i]>"
		$_ = az vmss identity assign --resource-group $node_resource_group --name $vmss_list[$i] --identities $in_cluster_defense_identity_name
	
		if ($LASTEXITCODE -ge 1)
		{
			write-error "Failed to attach identity to vmss <$vmss_list[$i]>"
			exit $LASTEXITCODE
		}
	}
    write-host "Attached identity to VMSS on node resource group successfully"
}

#######################################################################################################################
PrintNewSection("Installing Helm Chart")

# Step 6: Install helm chart

$in_cluster_defense_identity_client_id = az identity show -n $in_cluster_defense_identity_name -g $node_resource_group --query clientId

if ($LASTEXITCODE -ge 1 -or $in_cluster_defense_identity_client_id -eq "")
{
    write-error "Failed to get client id of in-cluster-defense identity"
    exit $LASTEXITCODE
}

# Get Cluster's creds and switch to current context
az aks get-credentials --resource-group $resource_group --name $cluster_name
kubectl config use-context $cluster_name

# Install helm chart
$env:HELM_EXPERIMENTAL_OCI=1
# Install helm chart from mcr repo on kube-system namespace and pass subscription and client id's params.
helm upgrade in-cluster-defense ./charts/azdproxy --install --wait `
            -n kube-system `
                --set AzDProxy.kubeletIdentity.envAzureAuthorizerConfiguration.mSIClientId=$kubelet_client_id `
                --set AzDProxy.azdIdentity.envAzureAuthorizerConfiguration.mSIClientId=$in_cluster_defense_identity_client_id `
                --set "AzDProxy.arg.argClientConfiguration.subscriptions={$subscription}" `
                --set AzDProxy.instrumentation.tivan.tivanInstrumentationConfiguration.region=$region `
                --set AzDProxy.instrumentation.tivan.tivanInstrumentationConfiguration.azureResourceID=$azureResourceID `
                --set AzDProxy.instrumentation.tivan.tivanInstrumentationConfiguration.componentName="InClusterDefense"