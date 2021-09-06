package registry

// ImageReference represents container image processed context and original identifier
type IImageReference interface {
	// Registry image ref registry (e.g. "tomer.azurecr.io")
	Registry() string
	// Repository image ref repository (e.g. "app/redis")
	Repository() string
	// Original fully qualified reference
	Original() string
}

// imageReference an abstract struct to share reference functionality
type imageReference struct {
	original   string
	repository string
	registry   string
}

// Registry image ref registry (e.g. "tomer.azurecr.io")
func (ref *imageReference) Registry() string {
	return ref.registry
}

// Repository image ref repository (e.g. "app/redis")
func (ref *imageReference) Repository() string {
	return ref.repository
}

// Original fully qualified reference
func (ref *imageReference) Original() string {
	return ref.original
}

// newImageReference abstract Ctor for sheared functionality of references
func newImageReference(original string, registry string, repository string) *imageReference {
	return &imageReference{
		registry:   registry,
		repository: repository,
		original:   original,
	}
}

// Tag represents a tag based image reference implements IImageReference interface
type Tag struct {
	imageReference
	tag string
}

// NewDigest Tag ctor
func NewTag(original string, registry string, repository string, tag string) *Tag {
	imageReference := newImageReference(original, registry, repository)
	return &Tag{
		imageReference: *imageReference,
		tag:            tag,
	}
}

// Tag return the tag part of the reference
func (t *Tag) Tag() string {
	return t.tag
}

// Digest represents a digest based image reference implements IImageReference interface
type Digest struct {
	imageReference
	digest string
}

// Digest return the digest part of the reference
func (d *Digest) Digest() string {
	return d.digest
}

// NewDigest Digest ctor
func NewDigest(original string, registry string, repository string, digest string) *Digest {
	imageReference := newImageReference(original, registry, repository)
	return &Digest{
		imageReference: *imageReference,
		digest:         digest,
	}
}
