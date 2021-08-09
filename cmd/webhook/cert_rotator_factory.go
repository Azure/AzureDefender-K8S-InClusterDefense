package webhook

import (
	"fmt"
	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	"k8s.io/apimachinery/pkg/types"
)

// ICertRotatorFactory  is factory of rotator.CertRotator
type ICertRotatorFactory interface {
	// CreateCertRotator Creates new cert rotator
	CreateCertRotator() (certRotator *rotator.CertRotator)
}

// CertRotatorFactory implements ICertRotatorFactory interface.
// It iss factory that creates rotator.CertRotator.
type CertRotatorFactory struct {
	// configuration is the configuration of the rotator.CertRotator
	configuration *CertRotatorConfiguration
}

// CertRotatorConfiguration is the certRotator configuration.
type CertRotatorConfiguration struct {
	// Namespace is the namespace that the pod is running.
	Namespace string
	// SecretName matches the Secret name.
	SecretName string
	// SecretName matches the Service name.
	ServiceName string
	// WebhookName matches the MutatingWebhookConfiguration name.
	WebhookName string
	// CaName is the Ca name.
	CaName string
	// CaOrganization
	CaOrganization string
	// CertDir is the directory that the certificates are saved.
	CertDir string
}

// NewCertRotatorFactory Creates new cert rotator factory
func NewCertRotatorFactory(configuration *CertRotatorConfiguration) (factory ICertRotatorFactory) {
	return &CertRotatorFactory{
		configuration: configuration,
	}
}

// CreateCertRotator Creates new cert rotator
func (factory *CertRotatorFactory) CreateCertRotator() (certRotator *rotator.CertRotator) {
	certSetupFinished := make(chan struct{})
	dnsName := factory.getDnsName()
	return &rotator.CertRotator{
		SecretKey: types.NamespacedName{
			Namespace: factory.configuration.Namespace,
			Name:      factory.configuration.SecretName,
		},
		CertDir:        factory.configuration.CertDir,
		CAName:         factory.configuration.CaName,
		CAOrganization: factory.configuration.CaOrganization,
		DNSName:        dnsName,
		IsReady:        certSetupFinished,
		Webhooks:       []rotator.WebhookInfo{{Name: factory.configuration.WebhookName, Type: rotator.Mutating}},
	}
}

// getDnsName returns the dns name of the server .
// It matches the MutatingWebhookConfiguration webhooks name.
// The format of the dns sever is <ServiceName>.<Namespace>.svc
func (factory *CertRotatorFactory) getDnsName() (dnsName string) {
	return fmt.Sprintf("%s.%s.svc", factory.configuration.ServiceName, factory.configuration.Namespace)
}
