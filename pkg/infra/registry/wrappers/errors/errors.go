package errors

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// ImageIsNotFoundErr  implements errors.error interface
var _ error = (*ImageIsNotFoundErr)(nil)

// ImageIsNotFoundErr  implements errors.error interface
var _ error = (*UnauthorizedErr)(nil)

// ImageIsNotFoundErr is error that returns when crane is not found an image.
type ImageIsNotFoundErr struct {
	imageRef string
	err      error
}

// GetImageIsNotFoundErrType returns double ptr to ImageIsNotFoundErr
func GetImageIsNotFoundErrType() **ImageIsNotFoundErr {
	err := ImageIsNotFoundErr{}
	ptrToErr := &err
	twoPtrToErr := &ptrToErr
	return twoPtrToErr
}

func newImageIsNotFoundErr(imageRef string, err error) *ImageIsNotFoundErr {
	return &ImageIsNotFoundErr{imageRef: imageRef, err: err}
}

func (err ImageIsNotFoundErr) Error() string {
	msg := fmt.Sprintf("Image is not found when trying to resolve image <%s>.\n error: <%s>", err.imageRef, err.err)
	return msg
}

// UnauthorizedErr error that returns due to unauthorized from crane.
type UnauthorizedErr struct {
	imageRef string
	err      error
}

// GetUnauthorizedErrType returns double ptr to UnauthorizedErr
func GetUnauthorizedErrType() **UnauthorizedErr {
	err := UnauthorizedErr{}
	ptrToErr := &err
	twoPtrToErr := &ptrToErr
	return twoPtrToErr
}

func newUnauthorizedErr(imageRef string, err error) *UnauthorizedErr {
	return &UnauthorizedErr{imageRef: imageRef, err: err}
}

func (err UnauthorizedErr) Error() string {
	msg := fmt.Sprintf("Unauthorized when trying to resolve image <%s>.\n error: <%s>", err.imageRef, err.err)
	return msg
}

var (
	//ImageIsNotFoundErrs is array that contains all error codes of image not found.
	// errors description: https://github.com/distribution/distribution/blob/main/docs/spec/api.md#errors-2
	ImageIsNotFoundErrs = []transport.ErrorCode{
		transport.ManifestUnknownErrorCode, //This error is returned when the manifest, identified by name and tag is unknown to the repository.
		transport.NameUnknownErrorCode,     // This is returned if the name used during an operation is unknown to the registry.
		transport.NameInvalidErrorCode,     // Invalid repository name encountered either during manifest validation or any API operation.
	}
)

func isErrorInImageIsNotFoundErrs(err *transport.Error) bool {
	if err == nil {
		return false
	}
	for _, errElem := range err.Errors {
		for _, imageIsNotFoundErr := range ImageIsNotFoundErrs {
			if errElem.Code == imageIsNotFoundErr {
				return true
			}
		}
	}
	return false
}

var (
	//UnauthorizedErrs is array that contains all error codes of unauthorized.
	// errors description: https://github.com/distribution/distribution/blob/main/docs/spec/api.md#errors-2
	UnauthorizedErrs = []transport.ErrorCode{
		transport.UnauthorizedErrorCode, //The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate.
	}
)

func isErrorInUnauthorizedErrs(err *transport.Error) bool {
	if err == nil {
		return false
	}
	for _, errElem := range err.Errors {
		for _, unauthorizedErr := range UnauthorizedErrs {
			if errElem.Code == unauthorizedErr {
				return true
			}
		}
	}
	return false
}

func ConvertErrToKnownErr(ref string, err error) error {
	if err == nil {
		return nil
	}
	// Check is the error is known error - if yes, convert it to error that is represented with our struct.
	transportError, ok := err.(*transport.Error)
	if ok {
		if isErrorInUnauthorizedErrs(transportError) {
			return newUnauthorizedErr(ref, err)
		} else if isErrorInImageIsNotFoundErrs(transportError) {
			return newImageIsNotFoundErr(ref, err)
		}
	}
	// Unknown error
	return err
}
