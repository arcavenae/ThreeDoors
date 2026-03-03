package tasks

import (
	"fmt"
	"os"
)

// NewProviderFromConfig creates a TaskProvider based on the given configuration.
// It first attempts to look up the provider in the default registry.
// If the registry has a matching factory, it uses that; otherwise it falls back
// to built-in defaults for backward compatibility.
func NewProviderFromConfig(config *ProviderConfig) TaskProvider {
	reg := DefaultRegistry()

	name := config.Provider
	if name == "" {
		name = "textfile"
	}

	if reg.IsRegistered(name) {
		provider, err := reg.InitProvider(name, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: adapter %q failed to initialize: %v, using textfile\n", name, err)
			return NewTextFileProvider()
		}
		return provider
	}

	fmt.Fprintf(os.Stderr, "Warning: unknown provider %q, using textfile\n", name)
	return NewTextFileProvider()
}
