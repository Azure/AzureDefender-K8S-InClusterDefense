package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/cmd/webhook"
)

// main is the entrypoint to azdproxy.
func main() {
	webhook.StartServer()
}
