package tasks

import (
	"fmt"
	"os"
)

// NewProviderFromConfig creates a TaskProvider based on the given configuration.
// It first attempts to use the new providers list (Story 7.2).
// If no providers list is present, it falls back to the legacy flat provider field.
// If the registry has a matching factory, it uses that; otherwise it falls back
// to built-in defaults for backward compatibility.
func NewProviderFromConfig(config *ProviderConfig) TaskProvider {
	reg := DefaultRegistry()

	provider, err := ResolveActiveProvider(config, reg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v, using textfile\n", err)
		return NewTextFileProvider()
	}

	return provider
}
