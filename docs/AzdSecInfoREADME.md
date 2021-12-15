# AZD Sec Info Provider README

##Background

- Due to the Kubernetes API server time frame, each mutation webhook should return its response in 3 seconds (defined [here](AzureDefender-K8S-InClusterDefense/charts/azdproxy/templates/mutation-configuration.yaml), otherwise the API server will continue without the webhook response.
- As a result, providing security information for the mutation webhook should be shorter than the time limit.

## AZDSecInfoProvider

AZDSecInfoProvider is a provider of azure defender security information.

- Scenarios table:

|                                                               | Timeout NOT encountered in current run                                                                | Timeout encountered in current run                                                                                                                 |
| ------------------------------------------------------------  |-------------------------------------------------------------------------------------------------------| ---------------------------------------------------------------------------------------------------------------------------------------------------|
| **Cache is down**                                             | Block request depends on the results.                                                                 | Don't block the request.                                                                                                                           |
| **No timeout in cache**                                       | Block request depends on the results. Set scan results in cache.                                      | Block request. Set timeout status to first time encountered in cache.Continue to get scan results in parallel run and set the results in cache.    |
| **One timeout in cache**                                      | Block request depends on the results. Set scan results in cache. Reset timeout status in cache.       | Block request. Set timeout status to second time encountered in cache.Continue to get scan results in parallel run and set the results in cache.   |
| **two timeouts in cache**                                     | Block request depends on the results. Set scan results in cache. Reset timeout status in cache.       | Don't block the request. Continue to get scan results in parallel run and set the results in cache.                                                |
