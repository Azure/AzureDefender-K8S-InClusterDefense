package registry

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
)

type ExtractImageRefContextUtilsTestSuite struct {
	suite.Suite
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractRegistryAndRepositoryFromImageReferencePublicImageRef() {
	ctx, err := ExtractImageRefContext("redis")
	suite.Nil(err)
	suite.Equal("index.docker.io", ctx.Registry)
	suite.Equal("library/redis", ctx.Repository)
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractRegistryAndRepositoryFromImageReferenceTag() {
	ctx, err := ExtractImageRefContext("tomer.azurecr.io/redis:v1")
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", ctx.Registry)
	suite.Equal("redis", ctx.Repository)
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_NoIdentifier_Err() {
	ctx, err := ExtractImageRefContext("tomer.azurecr.io/redis")
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", ctx.Registry)
	suite.Equal("redis", ctx.Repository)
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_Digest_Parsed() {
	imageRef := "tomer.azurecr.io/redis@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108"
	ctx, err := ExtractImageRefContext(imageRef)
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", ctx.Registry)
	suite.Equal("redis", ctx.Repository)
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_DigestBadFormat_Err() {
	imageRef := "tomer.azurecr.io/redis@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a"
	ctx, err := ExtractImageRefContext(imageRef)
	suite.Equal(reflect.TypeOf(&name.ErrBadName{}), reflect.TypeOf(err))
	suite.Nil(ctx)
}

func (suite *ExtractImageRefContextUtilsTestSuite) TestExtractImageRefContext_TagAndDigest_ParsedDigestIgnoreTag() {
	imageRef := "tomer.azurecr.io/redis:v1@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108"
	ctx, err := ExtractImageRefContext(imageRef)
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", ctx.Registry)
	suite.Equal("redis", ctx.Repository)
}

//func (suite *ExtractImageRefContextUtilsTestSuite) TestGetDigest(){
//	imageRef := "tomerwdevops.azurecr.io/alpine:v0"
//	digest, err := GetDigest(imageRef)
//	suite.Nil(err)
//	suite.Equal("sha256:d0710affa17fad5f466a70159cc458227bd25d4afb39514ef662ead3e6c99515", digest)
//}

func TestExtractImageRefContext(t *testing.T) {
	suite.Run(t, new(ExtractImageRefContextUtilsTestSuite))
}
