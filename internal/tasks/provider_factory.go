package tasks

import (
	"fmt"
	"os"
)

// NewProviderFromConfig creates a TaskProvider based on the given configuration.
func NewProviderFromConfig(config *ProviderConfig) TaskProvider {
	switch config.Provider {
	case "applenotes":
		primary := NewAppleNotesProvider(config.NoteTitle)
		fallback := NewTextFileProvider()
		return NewFallbackProvider(primary, fallback)
	case "textfile", "":
		return NewTextFileProvider()
	default:
		fmt.Fprintf(os.Stderr, "Warning: unknown provider %q, using textfile\n", config.Provider)
		return NewTextFileProvider()
	}
}
