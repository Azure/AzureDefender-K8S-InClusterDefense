package webhook

import (
	"fmt"
	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	"k8s.io/apimachinery/pkg/types"
)

type ICertRotatorFactory interface {
	// CreateCertRotator Creates new cert rotator
	CreateCertRotator(certDir string) (certRotator *rotator.CertRotator)
}

type CertRotatorFactory struct {
	configuration *CertRotatorConfiguration // configuration is the configuration of the rotator.CertRotator
}

type CertRotatorConfiguration struct {
	Namespace          string // Namespace is the namespace that the pod is running.
	SecretName         string // SecretName matches the Secret name.
	ServiceName        string // SecretName matches the Service name.
	WebhookName        string // WebhookName matches the MutatingWebhookConfiguration name.
	CaName             string // CaName is the Ca name.
	CaOrganization     string // CaOrganization
	EnableCertRotation bool   // EnableCertRotation is flag that indicates whether cert rotator should run
}

// NewCertRotatorFactory Creates new cert rotator factory
func NewCertRotatorFactory(configuration *CertRotatorConfiguration) (factory ICertRotatorFactory) {
	return &CertRotatorFactory{
		configuration: configuration,
	}
}

// CreateCertRotator Creates new cert rotator
func (factory *CertRotatorFactory) CreateCertRotator(certDir string) (certRotator *rotator.CertRotator) {
	certSetupFinished := make(chan struct{})
	dnsName := fmt.Sprintf("%s.%s.svc", factory.configuration.ServiceName, factory.configuration.Namespace) // matches the MutatingWebhookConfiguration webhooks name
	return &rotator.CertRotator{
		SecretKey: types.NamespacedName{
			Namespace: factory.configuration.Namespace,
			Name:      factory.configuration.SecretName,
		},
		CertDir:        certDir,
		CAName:         factory.configuration.CaName,
		CAOrganization: factory.configuration.CaOrganization,
		DNSName:        dnsName,
		IsReady:        certSetupFinished,
		Webhooks:       []rotator.WebhookInfo{{Name: factory.configuration.WebhookName, Type: rotator.Mutating}},
	}
}
