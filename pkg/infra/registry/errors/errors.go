package errors

import "fmt"

// ImageIsNotFoundErr  implements errors.error interface
var _ error = (*ImageIsNotFoundErr)(nil)

// ImageIsNotFoundErr is error that returns when crane is not found an image.
type ImageIsNotFoundErr struct {
	imageRef string
	err      error
}

func NewImageIsNotFoundErr(imageRef string, err error) *ImageIsNotFoundErr {
	return &ImageIsNotFoundErr{imageRef: imageRef, err: err}
}

func (err *ImageIsNotFoundErr) Error() string {
	msg := fmt.Sprintf("Image is not found when trying to resolve image <%s>.\n error: <%s>", err.imageRef, err.err)
	return msg
}

// RegistryIsNotFoundErr  implements errors.error interface
var _ error = (*RegistryIsNotFoundErr)(nil)

// RegistryIsNotFoundErr error that returns due to registry is not found.
type RegistryIsNotFoundErr struct {
	registry string
	err      error
}

// NewRegistryIsNotFoundErr  Constructor for RegistryIsNotFoundErr
func NewRegistryIsNotFoundErr(registry string, err error) *RegistryIsNotFoundErr {
	return &RegistryIsNotFoundErr{registry: registry, err: err}
}

func (err *RegistryIsNotFoundErr) Error() string {
	msg := fmt.Sprintf("Registry is not found when trying to resolve image <%s>.\n error: <%s>", err.registry, err.err)
	return msg
}

// ImageIsNotFoundErr  implements errors.error interface
var _ error = (*UnauthorizedErr)(nil)

// UnauthorizedErr error that returns due to unauthorized from crane.
type UnauthorizedErr struct {
	imageRef string
	err      error
}

// NewUnauthorizedErr  Constructor for UnauthorizedErr
func NewUnauthorizedErr(imageRef string, err error) *UnauthorizedErr {
	return &UnauthorizedErr{imageRef: imageRef, err: err}
}

func (err *UnauthorizedErr) Error() string {
	msg := fmt.Sprintf("Unauthorized when trying to resolve image <%s>.\n error: <%s>", err.imageRef, err.err)
	return msg
}
