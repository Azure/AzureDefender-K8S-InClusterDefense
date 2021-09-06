package utils

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	name "github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"strings"
)

const (
	// _azureContainerRegistrySuffix is the suffix of ACR public (todo extract per env maybe?)
	azureContainerRegistrySuffix = ".azurecr.io"
)

//GetImageReference receives image reference string (e.g. tomer.azurecr.io/redis:v1)
// Function extract and return received ref: registry and repository and identifiers like digest/tag, also dsaves the original ref str
// If image reference is not in right format or unknown, returns error.
func GetImageReference(imageRef string) (registry.IImageReference, error) {
	// Parse image ref
	parsedRef, err := name.ParseReference(imageRef)
	if err != nil {
		// Couldn't parse image ref
		return nil, errors.Wrap(err, "GetImageReference failed to parse imageRef")
	}

	// Create ref based on type
	switch refTyped := parsedRef.(type) {
	case name.Tag:
		tag := registry.NewTag(imageRef, refTyped.RegistryStr(), refTyped.RepositoryStr(), refTyped.TagStr())
		return tag, nil
	case name.Digest:
		digest := registry.NewDigest(imageRef, refTyped.RegistryStr(), refTyped.RepositoryStr(), refTyped.DigestStr())
		return digest, nil
	default:
		return nil, errors.New("GetImageReference Unknown parsed Ref type")
	}
}

// IsRegistryEndpointACR return is registryEndpoing is ACR based (ACR suffix)
func IsRegistryEndpointACR(registryEndpoint string) bool {
	return strings.HasSuffix(strings.ToLower(registryEndpoint), azureContainerRegistrySuffix)
}
