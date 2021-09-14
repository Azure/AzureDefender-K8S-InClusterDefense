package registry

// IRegistryClient container registry based client
type IRegistryClient interface {
	// GetDigestUsingDefaultAuth receives image reference and get it's digest using the default docker config auth
	GetDigestUsingDefaultAuth(imageReference IImageReference) (string, error)

	// GetDigestUsingACRAttachAuth receives image reference and get it's digest using ACR attach authntication
	// ACR attach auth is based MSI token used to access the registry
	GetDigestUsingACRAttachAuth(imageReference IImageReference) (string, error)

	// GetDigestUsingK8SAuth receives image reference and get it's digest using K8S secerts and auth
	// K8S auth is based image pull secrets used in deployment or attached to service account to pull the image
	GetDigestUsingK8SAuth(imageReference IImageReference, namespace string , imagePullSecrets []string, serviceAccountName string) (string, error)
}