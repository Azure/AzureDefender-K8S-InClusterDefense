package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	config "github.com/Azure/AzureDefender-K8S-InClusterDefense/configs"
	"log"
)

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	AppConfig, err := config.NewConfiguration("AppConfig","yaml",
		"/configs", true)
	if err != nil {
		log.Fatal(err)
	}

	// Create Configuration objects for each factory
	managerConfiguration := new(webhook.ManagerConfiguration)
	certRotatorConfiguration := new(webhook.CertRotatorConfiguration)
	serverConfiguration := new(webhook.ServerConfiguration)

	// Unmarshal the relevant parts of appConfig's data to each of the configuration objects
	CreateSubConfiguration(AppConfig, "webhook.ManagerConfiguration", managerConfiguration)
	CreateSubConfiguration(AppConfig, "webhook.CertRotatorConfiguration", certRotatorConfiguration)
	CreateSubConfiguration(AppConfig, "webhook.ServerConfiguration", serverConfiguration)

	// Create factories
	managerFactory := webhook.NewManagerFactory(managerConfiguration, nil)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfiguration)
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

// CreateSubConfiguration Create new configuration object for each resource,
// based on it's values in the main configuration file
func CreateSubConfiguration(mainConfiguration *config.Configuration, subConfigHierarchy string, configuration interface{}){
	ConfigValues := mainConfiguration.SubConfig(subConfigHierarchy)
	err := ConfigValues.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode the %v into struct, %v", subConfigHierarchy, err)
	}
}