package registry

// IRegistryClient container registry based client
type IRegistryClient interface {
	// GetDigest receives image reference string and resolve it's digest.
	GetDigestUsingDefaultAuth(imageReference IImageReference) (string, error)

	GetDigestUsingACRAttachAuth(imageReference IImageReference) (string, error)

	GetDigestUsingK8SAuth(imageReference IImageReference, namespace string , imagePullSecrets []string, serviceAccountName string) (string, error)
}


