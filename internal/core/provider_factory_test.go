package core

import (
	"testing"
)

// setupTestRegistry resets the default registry and registers test adapters.
func setupTestRegistry(t *testing.T) {
	t.Helper()
	orig := defaultRegistry
	defaultRegistry = NewRegistry()
	_ = defaultRegistry.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})
	_ = defaultRegistry.Register("applenotes", func(config *ProviderConfig) (TaskProvider, error) {
		primary := newInMemoryProvider()
		fallback := newInMemoryProvider()
		return NewFallbackProvider(primary, fallback), nil
	})
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
		return
	}

	_, ok := provider.(*inMemoryProvider)
	if !ok {
		t.Errorf("expected *inMemoryProvider, got %T", provider)
	}
}

func TestNewProviderFromConfig_AppleNotes(t *testing.T) {
	setupTestRegistry(t)

	cfg := &ProviderConfig{Provider: "applenotes", NoteTitle: "My Tasks"}

	provider := NewProviderFromConfig(cfg)
	if provider == nil {
		t.Fatal("NewProviderFromConfig() returned nil")
		return
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
		return
	}

	_, ok := provider.(*inMemoryProvider)
	if !ok {
		t.Errorf("expected *inMemoryProvider for empty provider, got %T", provider)
	}
}

func TestNewProviderFromConfig_UnknownProvider_DefaultsToTextFile(t *testing.T) {
	setupTestRegistry(t)

	cfg := &ProviderConfig{Provider: "unknown", NoteTitle: "ThreeDoors Tasks"}

	provider := NewProviderFromConfig(cfg)
	if provider == nil {
		t.Fatal("NewProviderFromConfig() returned nil")
		return
	}

	_, ok := provider.(*inMemoryProvider)
	if !ok {
		t.Errorf("expected *inMemoryProvider for unknown provider, got %T", provider)
	}
}
