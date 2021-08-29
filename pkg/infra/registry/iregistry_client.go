package registry

// IRegistryClient container registry based client
type IRegistryClient interface {
	// GetDigest receives image reference string and resolve it's digest.
	GetDigest(imageRef string) (string, error)
}
