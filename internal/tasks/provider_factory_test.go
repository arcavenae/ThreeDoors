package tasks

import (
	"testing"
)

// setupTestRegistry resets the default registry and registers built-in adapters.
func setupTestRegistry(t *testing.T) {
	t.Helper()
	// Replace the default registry with a fresh one for test isolation
	orig := defaultRegistry
	defaultRegistry = NewRegistry()
	RegisterBuiltinAdapters(defaultRegistry)
	t.Cleanup(func() {
		defaultRegistry = orig
	})
}

func TestNewProviderFromConfig_TextFile(t *testing.T) {
	setupTestRegistry(t)

	cfg := &ProviderConfig{Provider: "textfile", NoteTitle: "ThreeDoors Tasks"}

	provider := NewProviderFromConfig(cfg)
	if provider == nil {
		t.Fatal("NewProviderFromConfig() returned nil")
	}

	_, ok := provider.(*TextFileProvider)
	if !ok {
		t.Errorf("expected *TextFileProvider, got %T", provider)
	}
}

func TestNewProviderFromConfig_AppleNotes(t *testing.T) {
	setupTestRegistry(t)

	cfg := &ProviderConfig{Provider: "applenotes", NoteTitle: "My Tasks"}

	provider := NewProviderFromConfig(cfg)
	if provider == nil {
		t.Fatal("NewProviderFromConfig() returned nil")
	}

	_, ok := provider.(*FallbackProvider)
	if !ok {
		t.Errorf("expected *FallbackProvider wrapping AppleNotes, got %T", provider)
	}
}

func TestNewProviderFromConfig_EmptyProvider_DefaultsToTextFile(t *testing.T) {
	setupTestRegistry(t)

	cfg := &ProviderConfig{Provider: "", NoteTitle: "ThreeDoors Tasks"}

	provider := NewProviderFromConfig(cfg)
	if provider == nil {
		t.Fatal("NewProviderFromConfig() returned nil")
	}

	_, ok := provider.(*TextFileProvider)
	if !ok {
		t.Errorf("expected *TextFileProvider for empty provider, got %T", provider)
	}
}

func TestNewProviderFromConfig_UnknownProvider_DefaultsToTextFile(t *testing.T) {
	setupTestRegistry(t)

	cfg := &ProviderConfig{Provider: "unknown", NoteTitle: "ThreeDoors Tasks"}

	provider := NewProviderFromConfig(cfg)
	if provider == nil {
		t.Fatal("NewProviderFromConfig() returned nil")
	}

	_, ok := provider.(*TextFileProvider)
	if !ok {
		t.Errorf("expected *TextFileProvider for unknown provider, got %T", provider)
	}
}
