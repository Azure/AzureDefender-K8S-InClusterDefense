# Azure Defender K8S In Cluster Defense

![Alt text](.attachments/under-construction.png?raw=true "Under Construction")

## Installation

TODO Describe more on the installation process. \
\
Use installation_script.ps1 to install Azure Defender in your cluster.\
The params that you should pass are: \
- resource_group - Mandatory - the resource group of the cluster that you want to install Azure Defender on.\
- cluster_name - Mandatory - the name of the cluster that you want to install Azure Defender on.\
- helm_chart_version - the version of the helm chart.\
- should_install_azure_addon_policy - flag that indicates if you want to install Azure Addon Policy. if not, you should
install Gatekeeper manually. \
- should_enable_aks_security_profile - flag that indicates if you want to enable Azure Security Profile.\

```powershell
# Install Azure Defender in your cluster
.\installation_script.ps1 -subscription <subscription_id> -resource_group <resource_group> -cluster_name <cluster_name>
#>
```
