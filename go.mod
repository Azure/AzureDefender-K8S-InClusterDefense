module github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy

go 1.14

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/open-policy-agent/cert-controller v0.2.0
	go.uber.org/zap v1.17.0
	gomodules.xyz/jsonpatch/v2 v2.2.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog/v2 v2.8.0
	sigs.k8s.io/controller-runtime v0.9.0
)
