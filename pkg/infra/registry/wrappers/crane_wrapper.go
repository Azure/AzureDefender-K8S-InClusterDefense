package wrappers

import "github.com/google/go-containerregistry/pkg/crane"

// ICraneWrapper wraps crane operations
type ICraneWrapper interface {
	// Digest get image digest using image ref using crane Digest call
	Digest(ref string) (string, error)
}

// CraneWrapperwraps crante operations
type CraneWrapper struct{}

// Todo add auth options to pull secrets and ACR MSI based - currently only supports docker config auth
// K8s chain pull secrets ref: https://github.com/google/go-containerregistry/blob/main/pkg/authn/k8schain/k8schain.go
// ACR ref: // https://github.com/Azure/acr-docker-credential-helper/blob/master/src/docker-credential-acr/acr_login.go
func (*CraneWrapper) Digest(ref string) (string, error) {
	return crane.Digest(ref)
}
