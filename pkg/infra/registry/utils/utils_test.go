package utils

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ExtractImageRefContextUtilsTestSuite struct {
	suite.Suite
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractRegistryAndRepositoryFromImageReferencePublicImageRef() {
	ref, err := GetImageReference("redis")
	suite.Nil(err)
	suite.Equal("index.docker.io", ref.Registry())
	suite.Equal("library/redis", ref.Repository())
	suite.Equal("redis", ref.Original())
	tag, ok := ref.(*registry.Tag)
	suite.True(ok)
	suite.NotNil( tag)
	suite.Equal("latest", tag.Tag())
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractRegistryAndRepositoryFromImageReferenceTag() {
	ref, err := GetImageReference("tomer.azurecr.io/redis:v1")
	suite.Nil(err)
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", ref.Registry())
	suite.Equal("redis", ref.Repository())
	suite.Equal("tomer.azurecr.io/redis:v1", ref.Original())
	tag, ok := ref.(*registry.Tag)
	suite.True(ok)
	suite.NotNil( tag)
	suite.Equal("v1", tag.Tag())
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_NoIdentifier() {
	ref, err := GetImageReference("tomer.azurecr.io/redis")
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", ref.Registry())
	suite.Equal("redis", ref.Repository())
	suite.Equal("tomer.azurecr.io/redis", ref.Original())
	tag, ok := ref.(*registry.Tag)
	suite.True(ok)
	suite.NotNil( tag)
	suite.Equal("latest", tag.Tag())
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_Digest_Parsed() {
	imageRef := "tomer.azurecr.io/redis@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108"
	ref, err := GetImageReference(imageRef)
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", ref.Registry())
	suite.Equal("redis", ref.Repository())
	suite.Equal("tomer.azurecr.io/redis@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108", ref.Original())
	digest, ok := ref.(*registry.Digest)
	suite.True(ok)
	suite.NotNil( digest)
	suite.Equal("sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108", digest.Digest())
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_DigestBadFormat_Err() {
	// The last 4 chars of the digest are deleted:
	imageRef := "tomer.azurecr.io/redis@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a"
	ref, err := GetImageReference(imageRef)
	err = errors.Cause(err)
	_, ok := err.(*name.ErrBadName)
	suite.True(ok)
	suite.Nil(ref)
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_TagAndDigest_ParsedDigestIgnoreTag() {
	imageRef := "tomer.azurecr.io/redis:v1@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108"
	ref, err := GetImageReference(imageRef)
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io/redis:v1@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108", ref.Original())
	digest, ok := ref.(*registry.Digest)
	suite.True(ok)
	suite.NotNil(digest)
	suite.Equal("sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108", digest.Digest())
}

func TestExtractImageRefContext(t *testing.T) {
	suite.Run(t, new(ExtractImageRefContextUtilsTestSuite))
}
