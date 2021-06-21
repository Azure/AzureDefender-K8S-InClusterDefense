package main

import (
	"fmt"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/src/AzDProxy/webhook"
)

func main() {
	fmt.Println("AzDProxy is starting...")
	webhook.StartServer()
}
