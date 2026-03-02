package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProviderConfig_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("provider: applenotes\nnote_title: My Tasks\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}
	if cfg.Provider != "applenotes" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "applenotes")
	}
	if cfg.NoteTitle != "My Tasks" {
		t.Errorf("NoteTitle = %q, want %q", cfg.NoteTitle, "My Tasks")
	}
}

func TestLoadProviderConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nonexistent.yaml")

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() should not error for missing file, got: %v", err)
	}
	if cfg.Provider != "textfile" {
		t.Errorf("Provider = %q, want default %q", cfg.Provider, "textfile")
	}
	if cfg.NoteTitle != "ThreeDoors Tasks" {
		t.Errorf("NoteTitle = %q, want default %q", cfg.NoteTitle, "ThreeDoors Tasks")
	}
}

func TestLoadProviderConfig_InvalidYAML_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("{{{{invalid yaml content!!!!}")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadProviderConfig(configPath)
	if err == nil {
		t.Error("LoadProviderConfig() expected error for invalid YAML, got nil")
	}
}

func TestLoadProviderConfig_EmptyFile_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}
	if cfg.Provider != "textfile" {
		t.Errorf("Provider = %q, want default %q", cfg.Provider, "textfile")
	}
	if cfg.NoteTitle != "ThreeDoors Tasks" {
		t.Errorf("NoteTitle = %q, want default %q", cfg.NoteTitle, "ThreeDoors Tasks")
	}
}

func TestLoadProviderConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("provider: applenotes\nnote_title: Work Notes\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}
	if cfg.Provider != "applenotes" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "applenotes")
	}
	if cfg.NoteTitle != "Work Notes" {
		t.Errorf("NoteTitle = %q, want %q", cfg.NoteTitle, "Work Notes")
	}
}
