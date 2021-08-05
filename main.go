package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"os"
)

// main is the entrypoint to azdproxy.
func main() {
	server := server.NewServer()
	err := server.Run()
	if err != nil {
		os.Exit(1)
	}
}
