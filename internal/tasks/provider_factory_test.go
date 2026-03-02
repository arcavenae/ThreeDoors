package tasks

import (
	"testing"
)

func TestNewProviderFromConfig_TextFile(t *testing.T) {
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
