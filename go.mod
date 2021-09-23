module github.com/Azure/AzureDefender-K8S-InClusterDefense

go 1.16

require (
	tivan.ms/libs/instrumentation v0.0.0-20210803101155-9c6cc8e668ee

	github.com/Azure/azure-sdk-for-go v56.3.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.19
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coocood/freecache v1.1.1
	github.com/go-logr/logr v0.4.0
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-redis/redismock/v8 v8.0.6
	github.com/google/go-containerregistry v0.6.0
	github.com/google/go-containerregistry/pkg/authn/k8schain v0.0.0-20210823224117-e92a648af1b6
	github.com/modern-go/reflect2 v1.0.1
	github.com/open-policy-agent/cert-controller v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
	sigs.k8s.io/controller-runtime v0.9.0
)
