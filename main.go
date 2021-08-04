package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
)

// main is the entrypoint to azdproxy.
func main() {
	webhook.StartServer()
}
