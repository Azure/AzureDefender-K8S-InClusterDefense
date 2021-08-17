package util

import (
	name "github.com/google/go-containerregistry/pkg/name"
)

//GetRegistryAndRepositoryFromImageReference receives image reference string (e.g. tomer.azurecr.io/redis:v1)
// Function extract and return received ref: registry and repository (e.g registry: "tomer.azurecr,io", repository:"redis")
// If image reference is not in right format, returns error.
func GetRegistryAndRepositoryFromImageReference(imageReference string) (registry string, repository string, err error) {
	// Parse image ref
	parsedRef, err := name.ParseReference(imageReference)
	if err != nil {
		// Couldn't parse image ref
		return "", "", err
	}
	// Extract image's string representation of registry
	registry = parsedRef.Context().RegistryStr()
	// Extract image's string representation of repository
	repository = parsedRef.Context().RepositoryStr()
	return
}
