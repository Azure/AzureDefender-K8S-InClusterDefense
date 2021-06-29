// Package util is holds all the utilities that azdproxy is need.
package util

import "os"

// GetNamespace this function checks if there is POD_NAMESPACE as env variable, and if not, returns the default ns - kube-system
func GetNamespace() string {
	//TODO Maybe fetch this var from config-map/app parameter (performance)
	ns, found := os.LookupEnv("POD_NAMESPACE")
	if !found {
		return "kube-system"
	}
	return ns
}
