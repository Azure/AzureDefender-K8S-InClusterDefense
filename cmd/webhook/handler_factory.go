package webhook

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// IHandlerFactory Factory to create admission.Handler
type IHandlerFactory interface {
	// CreateHandler Initialize the manager object of the service - this object is manages the creation and registration
	// of the controllers of the server
	CreateHandler() (handler admission.Handler, err error)
}

// HandlerFactory Factory to create admission.Handler from configuration
type HandlerFactory struct {
	// instrumentationProviderFactory
	instrumentationProviderFactory instrumentation.IInstrumentationProviderFactory
	// Configuration is the manager configuration
	Configuration *HandlerConfiguration
}

// NewHandlerFactory Constructor for HandlerFactory
func NewHandlerFactory(configuration *HandlerConfiguration, instrumentationProviderFactory instrumentation.IInstrumentationProviderFactory) (factory IHandlerFactory) {
	return &HandlerFactory{
		Configuration:                  configuration,
		instrumentationProviderFactory: instrumentationProviderFactory,
	}
}

// HandlerConfiguration Factory configuration to create an admission.Handler
type HandlerConfiguration struct {
	// DryRun is flag that if it's true, it handles request but doesn't mutate the pod spec.
	DryRun bool
}

// CreateHandler is creating handler
// of the controllers of the server
func (factory *HandlerFactory) CreateHandler() (handler admission.Handler, err error) {
	provider, err := factory.instrumentationProviderFactory.CreateInstrumentationProvider()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create handler")
	}
	return NewHandler(factory.Configuration.DryRun, provider), nil
}
