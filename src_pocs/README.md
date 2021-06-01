# Demo

# Architecture
![Alt text](architecture.png?raw=true "Demo Architecture")
# Requirements
1. Create Local secret in src_pocs/proxy-arg/deployment/azcred.yaml:
    apiVersion: v1
    kind: Secret
    metadata:
    name: azcred
    namespace: proxy-arg-ns
    stringData:
    clientid: <ClientID>
    clientsecret: <ClientSecret>

    you can get the client ID ans client secret using the following queries:
     1. az identity show --resource-group mc_block-img-playgroung_blockimagesplayground_eastus2 --name block-pod-id --query clientId
     2. az identity show --resource-group mc_block-img-playgroung_blockimagesplayground_eastus2 --name block-pod-id --query clientSecretUrl
2. Deploy the webhook mutation server.
3. Verify that you have tag_2_digest.sh in the container machine (using exec command)
4. Verify that the policies are deployed (template and assignment) (gatekeeper/constraints/image_scan_constraint.yaml)
# Usage
1. You can modify the constraint to block instead of audit (dryrun) by deleting "  enforcementAction: dryrun" in the constraint file (gatekeeper/constraints/image_scan_constraint.yaml)
# References
## Mutation webhook server
1. Cert Controller - useful for creating certifacts for https server - https://github.com/open-policy-agent/cert-controller
2. How to create mutation webhook- https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74
## Tag2Digest:
- Digester - Google open source - https://github.com/google/k8s-digester
## ARG Proxy
### Authentication
- Pod Identity 
    1. Repo - https://github.com/Azure/aad-pod-identity
    2. How to use pod identity - https://blog.baeke.info/2019/02/02/aks-managed-pod-identity-and-access-to-azure-storage/
- ARG Client repo: https://github.com/Azure/azure-sdk-for-go/tree/master/services/resourcegraph/mgmt/2021-03-01/resourcegraph
## Policies
GateKeeper
1. open source of GK - https://github.com/open-policy-agent/gatekeeper
2. Webinar : how to use gk - https://www.youtube.com/watch?v=v4wJE3I8BYM
