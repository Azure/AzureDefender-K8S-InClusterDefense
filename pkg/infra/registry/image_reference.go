package registry

// IImageReference represents container image processed context and original identifier
type IImageReference interface {
	// Registry image ref registry (e.g. "tomer.azurecr.io")
	Registry() string
	// Repository image ref repository (e.g. "app/redis")
	Repository() string
	// Original fully qualified reference
	Original() string
}

// Digest implements IImageReference interface
var _ IImageReference = (*Digest)(nil)

// Digest represents a digest based image reference implements IImageReference interface
type Digest struct {
	imageReference
	digest string
}

// NewDigest Digest ctor
func NewDigest(original string, registry string, repository string, digest string) *Digest {
	imageReference := newImageReference(original, registry, repository)
	return &Digest{
		imageReference: *imageReference,
		digest:         digest,
	}
}

// Digest return the digest part of the reference
func (d *Digest) Digest() string {
	return d.digest
}

// Tag implements IImageReference interface
var _ IImageReference = (*Tag)(nil)

// Tag represents a tag based image reference implements IImageReference interface
type Tag struct {
	imageReference
	tag string
}

// NewTag Tag ctor
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

// imageReference implements IImageReference interface
var _ IImageReference = (*imageReference)(nil)

// imageReference an abstract struct to share reference functionality
type imageReference struct {
	original   string
	repository string
	registry   string
}

// newImageReference abstract Ctor for sheared functionality of references
func newImageReference(original string, registry string, repository string) *imageReference {
	return &imageReference{
		registry:   registry,
		repository: repository,
		original:   original,
	}
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
