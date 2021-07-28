package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/pkg/infra/instrumentation"
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
