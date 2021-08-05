package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/util"
	"os"
)

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	// Load configuration
	serverConfiguration := getServerConfiguration()
	managerConfiguration := getManagerConfiguration()
	// Create factories
	managerFactory := webhook.NewManagerFactory(managerConfiguration, nil)
	serverFactory := webhook.NewServerFactory(serverConfiguration, managerFactory, nil)
	// Create Server
	server, err := serverFactory.CreateServer()
	if err != nil {
		os.Exit(1)
	}
	// Run server
	if err = server.Run(); err != nil {
		os.Exit(1)
	}
}

//TODO All three methods below will be deleted once Or finishes the configuration
func getServerConfiguration() (configuration *webhook.ServerConfiguration) {
	certRotatorConfig := getCertRotatorConfiguration()
	return &webhook.ServerConfiguration{
		Path:              "/mutate",
		RunOnDryRunMode:   false,
		CertDir:           "/certs",
		CertRotatorConfig: certRotatorConfig,
	}
}

func getCertRotatorConfiguration() (configuration *webhook.CertRotatorConfiguration) {
	return &webhook.CertRotatorConfiguration{
		Namespace:          util.GetNamespace(),
		SecretName:         "azure-defender-proxy-cert",                           // matches the Secret name
		ServiceName:        "azure-defender-proxy-service",                        // matches the Service name
		WebhookName:        "azure-defender-proxy-mutating-webhook-configuration", // matches the MutatingWebhookConfiguration name
		CaName:             "azure-defender-proxy-ca",
		CaOrganization:     "azure-defender-proxy",
		EnableCertRotation: false,
	}
}

func getManagerConfiguration() (configuration *webhook.ManagerConfiguration) {
	return &webhook.ManagerConfiguration{
		Port:    8000,
		CertDir: "/certs",
	}
}
