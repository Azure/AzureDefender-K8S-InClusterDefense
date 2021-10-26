<#
    Installation script for In Cluster Defense - Private Preview script
    Maintainers: @ascdetectiontomer@microsoft.com>
    opensource:  https://github.com/Azure/AzureDefender-K8S-InClusterDefense
 # >

# Step 1: Get arguments: resourcegroup, cluster_name

# Step 2: Install azure addon policy in the cluster if not exists

#  Step 3: Create block identity (User managed identity) if not exists.

# Step 4: Create RBAC of subscription reader (RBAC of block identity) if not exists

# Step 5 ?: install Tivan if not exists (for publisher ?)

<# Step 6: Get all helm values for installing helm chart -
    - agentpool identity
    - subscription Identity
    - block identity
 #>

# Step 6: Install / upgrade helm chart with helm values