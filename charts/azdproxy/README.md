# AzDProxy Helm Chart

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.0](https://img.shields.io/badge/AppVersion-1.0.0-informational?style=flat-square) 

A Helm chart for AzDProxy

**Homepage:** <https://github.com/Azure/AzureDefender-K8S-InClusterDefense>

## Source Code

* <https://github.com/Azure/AzureDefender-K8S-InClusterDefense.git>

## Install Chart

```console
helm install azdproxy -n kube-system
```

The command deploys `azdproxy` on the Kubernetes cluster with the default configuration in the kube-system namespace.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] [CHART] --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Azure Security Center Detection Tomer's Team | ascdetectiontomer@microsoft.com |  |

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values azdproxy
```

The following table lists the configurable parameters of the azdproxy chart and their default values.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| AppConfig.instrumentation.LoggerCondifuration.logLevel | string | `"INFO"` |  |
| AppConfig.instrumentation.LoggerCondifuration.logLevelEncoder | string | `"lower"` |  |
| AppConfig.instrumentation.LoggerCondifuration.logLevelKey | string | `"level"` |  |
| AppConfig.webhook.CertRotatorConfiguration.CaName | string | `"azure-defender-proxy-ca"` |  |
| AppConfig.webhook.CertRotatorConfiguration.CaOrganization | string | `"azure-defender-proxy"` |  |
| AppConfig.webhook.CertRotatorConfiguration.CertDir | string | `"/certs"` |  |
| AppConfig.webhook.CertRotatorConfiguration.SecretName | string | `"azure-defender-proxy-cert"` |  |
| AppConfig.webhook.CertRotatorConfiguration.ServiceName | string | `"azure-defender-proxy-service"` |  |
| AppConfig.webhook.CertRotatorConfiguration.WebhookName | string | `"azure-defender-proxy-mutating-webhook-configuration"` |  |
| AppConfig.webhook.ManagerConfiguration.CertDir | string | `"/certs"` |  |
| AppConfig.webhook.ManagerConfiguration.Port | int | `8000` |  |
| AppConfig.webhook.ServerConfiguration.EnableCertRotation | bool | `true` |  |
| AppConfig.webhook.ServerConfiguration.Path | string | `"/mutate"` |  |
| AppConfig.webhook.ServerConfiguration.RunOnDryRunMode | bool | `false` |  |
| AzDProxy.configuration.volume.mountPath | string | `"/configs"` | The mount path of the volume. |
| AzDProxy.configuration.volume.name | string | `"config"` | The name of the volume. |
| AzDProxy.prefixResourceDeployment | string | `"azure-defender-proxy"` | common prefix name for all resources. |
| AzDProxy.service.targetPort | int | `8000` | The port on which the service will send requests to, so the webhook be listening on. |
| AzDProxy.webhook.image.name | string | `"azdproxy-image"` | Official image. |
| AzDProxy.webhook.image.pullPolicy | string | `"Always"` | Default for always. in case that you want to use local registry, change to 'Never'. |
| AzDProxy.webhook.mutationPath | string | `"/mutate"` | The path that the webhook handler will be listening on. |
| AzDProxy.webhook.replicas | int | `3` | Amount of replicas of azdproxy. |
| AzDProxy.webhook.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"}}` | The resources of the webhook. |
| AzDProxy.webhook.volume.mountPath | string | `"/certs"` | The mount path of the volume. |
| AzDProxy.webhook.volume.name | string | `"cert"` | The name of the volume. |
| AzDProxy.webhook_configuration.timeoutSeconds | int | `3` | Webhook timeout in seconds |

<!-- markdownlint-enable MD013 MD034 -->
<!-- markdownlint-restore -->