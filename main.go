package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
)

// main is the entrypoint to azdproxy.
func main() {
	// Creates instrumentation
	instrumentationFactory := instrumentation.NewInstrumentationFactory()
	serverInstrumentation, err := instrumentationFactory.CreateInstrumentation()
	if err != nil {
		//TODO Error flow
		return
	}
	server := webhook.NewServer(serverInstrumentation)
	// Run server
	server.Run()
}
