package provider

import "github.com/hellofresh/janus/pkg/types"

type (
	// Provider defines methods of a provider.
	Provider interface {
		// Provide allows the provider to provide configurations to janus
		// using the given configuration channel.
		Provide(configChan chan<- types.ConfigMessage, configChangeChan chan types.ConfigurationEvent) error
	}
)
