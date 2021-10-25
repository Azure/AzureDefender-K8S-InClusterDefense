package errors

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

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

// TryParseCraneErrToRegistryKnownErr Gets an err and ref (the reference of the image that crane tried to get the digest) and try to convert it to known err
func TryParseCraneErrToRegistryKnownErr(ref string, err error) (error, bool) {
	if err == nil {
		return nil, false
	}
	// Check is the error is known error - if yes, convert it to error that is represented with our struct.
	transportError, ok := err.(*transport.Error)
	if ok {
		if isErrorInUnauthorizedErrs(transportError) {
			return errors.NewUnauthorizedErr(ref, err), true
		} else if isErrorInImageIsNotFoundErrs(transportError) {
			return errors.NewImageIsNotFoundErr(ref, err), true
		}
	}
	// Unknown error
	return err, false
}
