package registry

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

const _imageref_registry = "tomerw.devops.io"
const _imageref_repository = "redis"
const _imageref_tag = "v3"
const _imageref_digest = "xxxxxxxyyyyyxxxx"

type TestSuiteImageRef struct {
	suite.Suite
}

func (suite *TestSuiteImageRef) Test_TagRef() {
	original := _imageref_registry + "/" + _imageref_repository + ":" + _imageref_tag
	tagRef := NewTag(original, _imageref_registry, _imageref_repository, _imageref_tag)

	tagRefExpected := &Tag{imageReference{original:  original, repository: _imageref_repository, registry: _imageref_registry}, _imageref_tag}
	suite.Exactly(tagRefExpected, tagRef)
	suite.Equal(_imageref_tag, tagRef.Tag())
	suite.Equal(_imageref_repository, tagRef.Repository())
	suite.Equal(_imageref_registry, tagRef.Registry())
	suite.Equal(original, tagRef.Original())
}

func (suite *TestSuiteImageRef) Test_DigestRef() {
	original := _imageref_registry + "/" + _imageref_repository + "@" + _imageref_digest
	digestRef := NewDigest(original, _imageref_registry, _imageref_repository, _imageref_digest)

	digestRefExpected := &Digest{imageReference{original:  original, repository: _imageref_repository, registry: _imageref_registry}, _imageref_digest}
	suite.Exactly(digestRefExpected, digestRef)
	suite.Equal(_imageref_digest, digestRef.Digest())
	suite.Equal(_imageref_repository, digestRef.Repository())
	suite.Equal(_imageref_registry, digestRef.Registry())
	suite.Equal(original, digestRef.Original())
}

func Test_Suite(t *testing.T) {
	suite.Run(t, new(TestSuiteImageRef))
}
