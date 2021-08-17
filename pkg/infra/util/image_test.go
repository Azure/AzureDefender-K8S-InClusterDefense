package util

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
)

type ImageUtilsTestSuite struct {
	suite.Suite
}

func (suite *ImageUtilsTestSuite) TestGetRegistryAndRepositoryFromImageReferencePublicImageRef(){
	registry, repository, err := GetRegistryAndRepositoryFromImageReference("redis")
	suite.Nil(err)
	suite.Equal("index.docker.io", registry)
	suite.Equal("library/redis", repository)
}

func (suite *ImageUtilsTestSuite) TestGetRegistryAndRepositoryFromImageReferenceTag(){
	registry, repository, err := GetRegistryAndRepositoryFromImageReference("tomer.azurecr.io/redis:v1")
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", registry)
	suite.Equal("redis", repository)
}

func (suite *ImageUtilsTestSuite) TestGetRegistryAndRepositoryFromImageReferenceNoIdentifier(){
	registry, repository, err := GetRegistryAndRepositoryFromImageReference("tomer.azurecr.io/redis")
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", registry)
	suite.Equal("redis", repository)
}

func (suite *ImageUtilsTestSuite) TestGetRegistryAndRepositoryFromImageReferenceDigest(){
	imageRef := "tomer.azurecr.io/redis@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108"
	registry, repository, err := GetRegistryAndRepositoryFromImageReference(imageRef)
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", registry)
	suite.Equal("redis", repository)
}

func (suite *ImageUtilsTestSuite) TestGetRegistryAndRepositoryFromImageReferenceDigestBadFormat(){
	imageRef := "tomer.azurecr.io/redis@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a"
	registry, repository, err := GetRegistryAndRepositoryFromImageReference(imageRef)
	suite.Equal(reflect.TypeOf(&name.ErrBadName{}),reflect.TypeOf(err))
	suite.Equal("", registry)
	suite.Equal("", repository)
}

func (suite *ImageUtilsTestSuite) TestGetRegistryAndRepositoryFromImageReferenceDigestAndDigest(){
	imageRef := "tomer.azurecr.io/redis:v1@sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108"
	registry, repository, err := GetRegistryAndRepositoryFromImageReference(imageRef)
	suite.Nil(err)
	suite.Equal("tomer.azurecr.io", registry)
	suite.Equal("redis", repository)
}

func TestParseImageReference(t *testing.T) {
	suite.Run(t, new(ImageUtilsTestSuite))
}