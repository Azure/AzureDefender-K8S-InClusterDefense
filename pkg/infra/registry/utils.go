package registry

import (
	name "github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"strings"
)

const (
	digestDelim = "@"
)

// ImageRefContext represents image ref context - registry and repository
type ImageRefContext struct {
	// Registry image ref registry (e.g. "tomer.azurecr.io")
	Registry string
	// Repository image ref repository (e.g. "app/redis")
	Repository string
}

//ExtractImageRefContext receives image reference string (e.g. tomer.azurecr.io/redis:v1)
// Function extract and return received ref: registry and repository (e.g registry: "tomer.azurecr,io", repository:"redis")
// If image reference is not in right format, returns error.
func ExtractImageRefContext(imageRef string) (*ImageRefContext, error) {
	// Parse image ref
	parsedRef, err := name.ParseReference(imageRef)
	if err != nil {
		// Couldn't parse image ref
		return nil, errors.Wrap(err, "ExtractImageRefContext failed to parse imageRef")
	}

	ctx := &ImageRefContext{
		// Extract image's string representation of registry
		Registry: parsedRef.Context().RegistryStr(),
		// Extract image's string representation of repository
		Repository: parsedRef.Context().RepositoryStr(),
	}
	return ctx, nil
}

func TryExtractDigestFromImageRef(imageRef string) (isDigestBasedImageRef bool, digest string, err error) {
	if strings.Contains(imageRef, digestDelim) {
		ref, err := name.ParseReference(imageRef)
		if err != nil {
			// Report error - should be a valid reference
			err = errors.Wrap(err, "utils.TryExtractDigestFromImageRef:")
			return false, "", err
		}

		digestImageRef, ok := ref.(name.Digest)
		if ok {
			// Digest based image ref - extract it from
			return true, digestImageRef.DigestStr(), nil
		}
	}
	return false, "", nil
}
