package errors

import (
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/pkg/errors"
)

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

// TryParseErrToUnscannedWithReason gets an error the container that the error encountered and returns the info and error according to the type of the error.
// If the error is expected error (e.g. image is not exists while trying to resolve the digest, unauthorized to arg) then
// this function create new contracts.ContainerVulnerabilityScanInfo that that status is unscanned and add in the additional metadata field
// the reason for unscanned - for example contracts.ImageDoesNotExistUnscannedReason.
// If the function doesn't recognize the error, then it returns nil, err
func TryParseErrToUnscannedWithReason(err error) (*contracts.UnscannedReason, bool) {
	// Check if the err  known error:
	// TODO - sort the errors (most frequent should be first error)
	// TODO add metrics for this method.
	// Checks if the error  image  not found error
	cause := errors.Cause(err)
	// Try to parse the cause of the error to known error -> if true, resolve to unscanned reason.
	switch cause.(type) {
	case *ImageIsNotFoundErr: // Checks if the error  Image DoesNot Exist
		unscannedReason := contracts.ImageDoesNotExistUnscannedReason
		return &unscannedReason, true
	case *UnauthorizedErr: // Checks if the error  unauthorized
		unscannedReason := contracts.RegistryUnauthorizedUnscannedReason
		return &unscannedReason, true
	case *RegistryIsNotFoundErr: // Checks if the error  NoSuchHost - it means that the registry  not found.
		unscannedReason := contracts.RegistryDoesNotExistUnscannedReason
		return &unscannedReason, true
	default: // Unexpected error
		return nil, false
	}
}

// TryParseStringToUnscannedWithReasonErr gets a string and try to convert it to a known error from contracts.
func TryParseStringToUnscannedWithReasonErr(errAsString string) (error, bool){
	switch errAsString {
	case string(contracts.ImageDoesNotExistUnscannedReason):
		return new(ImageIsNotFoundErr), true
	case string(contracts.RegistryUnauthorizedUnscannedReason):
		return new(UnauthorizedErr), true
	case string(contracts.RegistryDoesNotExistUnscannedReason):
		return new(RegistryIsNotFoundErr), true
	default: // Unknown error or not an error
		return nil, false
	}
}
