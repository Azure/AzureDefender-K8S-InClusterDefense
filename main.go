package main

import (
	"log"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/util"
)

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	// Load configuration
	managerConfiguration := getManagerConfiguration()
	certRotatorConfig := getCertRotatorConfiguration()
	serverConfiguration := getServerConfiguration()
	handlerConfiguration := gerHandlerConfiguration()

	// Create factories
	managerFactory := webhook.NewManagerFactory(managerConfiguration, nil)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfig)

	azdSecInfoProvider := azdsecinfo.NewAzdSecInfoProvider()
	handler := webhook.NewHandler(azdSecInfoProvider, handlerConfiguration, nil)
	serverFactory := webhook.NewServerFactory(serverConfiguration, managerFactory, certRotatorFactory, handler, nil)
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
func gerHandlerConfiguration() *webhook.HandlerConfiguration {
	return &webhook.HandlerConfiguration{
		DryRun: false,
	}
}
func getServerConfiguration() *webhook.ServerConfiguration {
	return &webhook.ServerConfiguration{
		Path:               "/mutate",
		EnableCertRotation: true,
	}
}

func getCertRotatorConfiguration() *webhook.CertRotatorConfiguration {
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

func getManagerConfiguration() *webhook.ManagerConfiguration {
	return &webhook.ManagerConfiguration{
		Port:    8000,
		CertDir: "/certs",
	}
}
