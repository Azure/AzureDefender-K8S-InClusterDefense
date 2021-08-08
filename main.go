package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/util"
	"log"
)

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	// Load configuration
	managerConfiguration := getManagerConfiguration()
	certRotatorConfig := getCertRotatorConfiguration()
	serverConfiguration := getServerConfiguration()

	// Create factories
	managerFactory := webhook.NewManagerFactory(managerConfiguration, nil)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfig)
	serverFactory := webhook.NewServerFactory(serverConfiguration, managerFactory, certRotatorFactory, nil)
	// Create Server
	server, err := serverFactory.CreateServer()
	if err != nil {
		log.Fatal(err)
	}
	// Run server
	if err = server.Run(); err != nil {
		log.Fatal(err)
	}
}

//TODO All three methods below will be deleted once Or finishes the configuration
func getServerConfiguration() (configuration *webhook.ServerConfiguration) {
	return &webhook.ServerConfiguration{
		Path:               "/mutate",
		RunOnDryRunMode:    false,
		EnableCertRotation: true,
	}
}

func getCertRotatorConfiguration() (configuration *webhook.CertRotatorConfiguration) {
	return &webhook.CertRotatorConfiguration{
		Namespace:      util.GetNamespace(),
		SecretName:     "azure-defender-proxy-cert",                           // matches the Secret name
		ServiceName:    "azure-defender-proxy-service",                        // matches the Service name
		WebhookName:    "azure-defender-proxy-mutating-webhook-configuration", // matches the MutatingWebhookConfiguration name
		CaName:         "azure-defender-proxy-ca",
		CaOrganization: "azure-defender-proxy",
		CertDir:        "/certs",
	}
}

func getManagerConfiguration() (configuration *webhook.ManagerConfiguration) {
	return &webhook.ManagerConfiguration{
		Port:    8000,
		CertDir: "/certs",
	}
}
