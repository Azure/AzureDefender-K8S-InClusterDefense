# Project

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