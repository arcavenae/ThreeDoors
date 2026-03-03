package core

import (
	"fmt"
	"os"
)

// NewProviderFromConfig creates a TaskProvider based on the given configuration.
// It first attempts to use the new providers list (Story 7.2).
// If no providers list is present, it falls back to the legacy flat provider field.
// If the registry has a matching factory, it uses that; otherwise it falls back
// to the "textfile" adapter from the registry for backward compatibility.
func NewProviderFromConfig(config *ProviderConfig) TaskProvider {
	reg := DefaultRegistry()

	provider, err := ResolveActiveProvider(config, reg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v, using textfile\n", err)
		fallback, fallbackErr := reg.InitProvider("textfile", config)
		if fallbackErr != nil {
			fmt.Fprintf(os.Stderr, "Error: textfile fallback failed: %v\n", fallbackErr)
			return nil
		}
		return fallback
	}

	return provider
}
