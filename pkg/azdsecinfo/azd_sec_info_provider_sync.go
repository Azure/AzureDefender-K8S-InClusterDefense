package azdsecinfo

import (
	"context"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/wrappers"
)

// AzdSecInfoProviderSync contains data members that responsible to synchronize AzdSecInfoProvider's goroutines
type AzdSecInfoProviderSync struct {
	// AzdSecInfoProviderCtx is AzdSecInfoProvider context
	AzdSecInfoProviderCtx context.Context
	// cancelScans is a function that once called, signify that work done on behalf of this context should be canceled
	cancelScans context.CancelFunc
	// vulnerabilitySecInfoChannel is a channel for *wrappers.ContainerVulnerabilityScanInfoWrapper
	vulnerabilitySecInfoChannel chan *wrappers.ContainerVulnerabilityScanInfoWrapper
}

// NewAzdSecInfoProviderSync - AzdSecInfoProviderSync Ctor
func NewAzdSecInfoProviderSync()  *AzdSecInfoProviderSync {
	azdSecInfoProviderSync := &AzdSecInfoProviderSync{
		vulnerabilitySecInfoChannel: make(chan *wrappers.ContainerVulnerabilityScanInfoWrapper),
	}
	azdSecInfoProviderSync.AzdSecInfoProviderCtx, azdSecInfoProviderSync.cancelScans = context.WithCancel(context.Background())
	return azdSecInfoProviderSync
}
