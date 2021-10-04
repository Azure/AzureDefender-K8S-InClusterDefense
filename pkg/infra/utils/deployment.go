package utils

import "errors"

var (
	// _singleton is the only instance of deployment struct.
	_singleton *deployment = nil
)

// deployment is struct the helps to manage all fields of deployment using global DEPLOYMENT.
// In order to use it, you should use DEPLOYMENT instance only
type deployment struct {
	// configuration holds instance of deploymentConfiguration.
	configuration *DeploymentConfiguration
}

// DeploymentConfiguration is deployment configuration
type DeploymentConfiguration struct {
	// IsLocalDevelopment is boolean that indicate if we are in local development or not.
	IsLocalDevelopment bool
	// Namespace  is the Namespace where the server is running
	Namespace string
}

// NewDeployment creates new deployment in case of singleton is not initialized yet.
// returns error when there is already initialized instance.
func NewDeployment(configuration *DeploymentConfiguration) (*deployment, error) {
	if _singleton != nil {
		return nil, errors.New("can't create another instance of Deployment")
	}
	_singleton = &deployment{configuration: configuration}
	return _singleton, nil
}

// GetDeploymentInstance returns the singleton instance of deployment.
//in case that the deployment is not initialized, returns nil.
func GetDeploymentInstance() *deployment {
	return _singleton
}

// IsLocalDevelopment returns boolean that indicate if we are in local deployment or not.
func (d *deployment) IsLocalDevelopment() bool {
	return d.configuration.IsLocalDevelopment
}

// GetNamespace returns the Namespace where the server is running
func (d *deployment) GetNamespace() string {
	return d.configuration.Namespace
}

// UpdateDeploymentForTests is used for update the singleton for tests purpose.
// ------ Shouldn't be used in production code ---------------
func UpdateDeploymentForTests(configuration *DeploymentConfiguration) {
	_singleton = &deployment{configuration}
}
