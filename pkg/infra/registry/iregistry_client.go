package registry

import registryauth "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/auth"

// IRegistryClient container registry based client
type IRegistryClient interface {
	// GetDigest receives image reference string and resolve it's digest.
	GetDigest(imageRef string, registryAuthCtx *registryauth.AuthContext) (string, error)
}
