# Set account to dev
az account set --subscription 4009f3ee-43c4-4f19-97e4-32b6f2285a68;
# Login to ACR
az acr login --name orplay;
# Build in case that we use instrumentation of AzureDevOps
docker build -t orplay.azurecr.io/azdproxy-image -f build/Dockerfile . --build-arg DEVOPS_TOKEN=inydlknpe32ajxqmzpxmp3vpct5reabgghmotj5npgjlrz6hj4aq;
# Push to ACR
docker push orplay.azurecr.io/azdproxy-image;
# Uninstall last helm update:
helm uninstall azure-defender -n kube-system
# Install new helm with dev values.
helm install --namespace kube-system azure-defender charts/azdproxy --set AzDProxy.webhook.replicas=1 --set AzDProxy.webhook.image.name=orplay.azurecr.io/azdproxy-image --set AzDProxy.webhook_configuration.timeoutSeconds=10 --set AzDProxy.azdIdentity.envAzureAuthorizerConfiguration.mSIClientId="3fa51155-37d1-4987-84f4-9f42f4b03c4d" --set AzDProxy.kubeletIdentity.envAzureAuthorizerConfiguration.mSIClientId="45af1e95-1288-4579-b8a7-cddd6cb06094" --set "AzDProxy.arg.argClientConfiguration.subscriptions={409111bf-3097-421c-ad68-a44e716edf58}" 