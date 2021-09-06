package registry

// ImageReference represents container image - context and identifiers
type IImageReference interface {
	// Registry image ref registry (e.g. "tomer.azurecr.io")
	Registry() string
	// Repository image ref repository (e.g. "app/redis")
	Repository() string
	//Original fully qualified reference
	Original() string
}

type imageReference struct {
	original   string
	repository string
	registry   string
}

func (ref *imageReference) Registry() string {
	return ref.registry
}
func (ref *imageReference) Repository() string {
	return ref.repository
}
func (ref *imageReference) Original() string {
	return ref.original
}

func newImageReference(original string, registry string, repository string) *imageReference {
	return &imageReference{
		registry:   registry,
		repository: repository,
		original:   original,
	}
}

type Tag struct {
	imageReference
	tag string
}

func NewTag(original string, registry string, repository string, tag string) *Tag {
	imageReference := newImageReference(original, registry, repository)
	return &Tag{
		imageReference: *imageReference,
		tag:            tag,
	}
}

func (t *Tag) Tag() string {
	return t.tag
}

type Digest struct {
	imageReference
	digest string
}

func (d *Digest) Digest() string {
	return d.digest
}

func NewDigest(original string, registry string, repository string, digest string) *Digest {
	imageReference := newImageReference(original, registry, repository)
	return &Digest{
		imageReference: *imageReference,
		digest:         digest,
	}
}
