package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// --- Story 7.2 Tests: Config-Driven Provider Selection ---

func TestLoadProviderConfig_WithProvidersList(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`providers:
  - name: textfile
    settings:
      task_file: ~/custom/tasks.yaml
  - name: applenotes
    settings:
      note_title: Work Tasks
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	if len(cfg.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(cfg.Providers))
	}
	if cfg.Providers[0].Name != "textfile" {
		t.Errorf("Providers[0].Name = %q, want %q", cfg.Providers[0].Name, "textfile")
	}
	if cfg.Providers[0].Settings["task_file"] != "~/custom/tasks.yaml" {
		t.Errorf("Providers[0].Settings[task_file] = %q, want %q", cfg.Providers[0].Settings["task_file"], "~/custom/tasks.yaml")
	}
	if cfg.Providers[1].Name != "applenotes" {
		t.Errorf("Providers[1].Name = %q, want %q", cfg.Providers[1].Name, "applenotes")
	}
	if cfg.Providers[1].Settings["note_title"] != "Work Tasks" {
		t.Errorf("Providers[1].Settings[note_title] = %q, want %q", cfg.Providers[1].Settings["note_title"], "Work Tasks")
	}
}

func TestLoadProviderConfig_ProvidersListEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("providers: []\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	if len(cfg.Providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(cfg.Providers))
	}
}

func TestLoadProviderConfig_BackwardCompatibleFlatProvider(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Old-style flat config should still work
	content := []byte("provider: applenotes\nnote_title: My Tasks\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	// Flat provider field should still be parsed
	if cfg.Provider != "applenotes" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "applenotes")
	}
	if cfg.NoteTitle != "My Tasks" {
		t.Errorf("NoteTitle = %q, want %q", cfg.NoteTitle, "My Tasks")
	}
}

func TestLoadProviderConfig_ProviderEntryWithNoSettings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`providers:
  - name: textfile
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	if len(cfg.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(cfg.Providers))
	}
	if cfg.Providers[0].Name != "textfile" {
		t.Errorf("Providers[0].Name = %q, want %q", cfg.Providers[0].Name, "textfile")
	}
	if len(cfg.Providers[0].Settings) != 0 {
		t.Errorf("expected nil or empty settings, got %v", cfg.Providers[0].Settings)
	}
}

func TestProviderEntry_GetSetting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    ProviderEntry
		key      string
		fallback string
		want     string
	}{
		{
			name:     "existing key",
			entry:    ProviderEntry{Name: "test", Settings: map[string]string{"key": "value"}},
			key:      "key",
			fallback: "default",
			want:     "value",
		},
		{
			name:     "missing key returns fallback",
			entry:    ProviderEntry{Name: "test", Settings: map[string]string{"other": "value"}},
			key:      "key",
			fallback: "default",
			want:     "default",
		},
		{
			name:     "nil settings returns fallback",
			entry:    ProviderEntry{Name: "test"},
			key:      "key",
			fallback: "default",
			want:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.entry.GetSetting(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("GetSetting(%q, %q) = %q, want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestGenerateSampleConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	reg := NewRegistry()
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})
	_ = reg.Register("applenotes", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})

	err := GenerateSampleConfig(configPath, reg)
	if err != nil {
		t.Fatalf("GenerateSampleConfig() unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read generated config: %v", err)
	}

	content := string(data)

	// Should contain the active textfile provider
	if !strings.Contains(content, "textfile") {
		t.Error("sample config should mention textfile provider")
	}

	// Should contain commented-out examples
	if !strings.Contains(content, "#") {
		t.Error("sample config should contain comments")
	}
}

func TestGenerateSampleConfig_DoesNotOverwriteExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	existing := []byte("provider: applenotes\n")
	if err := os.WriteFile(configPath, existing, 0o644); err != nil {
		t.Fatalf("failed to write existing config: %v", err)
	}

	reg := NewRegistry()
	err := GenerateSampleConfig(configPath, reg)
	if err != nil {
		t.Fatalf("GenerateSampleConfig() unexpected error: %v", err)
	}

	// Existing file should not be overwritten
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if string(data) != string(existing) {
		t.Errorf("existing config was overwritten: got %q, want %q", string(data), string(existing))
	}
}

func TestResolveActiveProvider_WithProvidersList(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{Name: "textfile", Settings: map[string]string{}},
		},
	}

	provider, err := ResolveActiveProvider(cfg, reg)
	if err != nil {
		t.Fatalf("ResolveActiveProvider() unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("ResolveActiveProvider() returned nil")
	}

	_, ok := provider.(*inMemoryProvider)
	if !ok {
		t.Errorf("expected *inMemoryProvider, got %T", provider)
	}
}

func TestResolveActiveProvider_FallbackToFlatProvider(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})

	// Old-style config with flat provider field, no providers list
	cfg := &ProviderConfig{
		Provider:  "textfile",
		NoteTitle: "ThreeDoors Tasks",
	}

	provider, err := ResolveActiveProvider(cfg, reg)
	if err != nil {
		t.Fatalf("ResolveActiveProvider() unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("ResolveActiveProvider() returned nil")
	}
}

func TestResolveActiveProvider_NoConfig_DefaultsToTextFile(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})

	// Empty config — should default to textfile
	cfg := &ProviderConfig{}

	provider, err := ResolveActiveProvider(cfg, reg)
	if err != nil {
		t.Fatalf("ResolveActiveProvider() unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("ResolveActiveProvider() returned nil")
	}

	_, ok := provider.(*inMemoryProvider)
	if !ok {
		t.Errorf("expected *inMemoryProvider, got %T", provider)
	}
}

func TestResolveActiveProvider_UnknownProvider_ReturnsError(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{Name: "nonexistent"},
		},
	}

	_, err := ResolveActiveProvider(cfg, reg)
	if err == nil {
		t.Error("ResolveActiveProvider() expected error for unknown provider, got nil")
	}
}

func TestResolveActiveProvider_SettingsPassedToFactory(t *testing.T) {
	t.Parallel()

	var receivedSettings map[string]string
	reg := NewRegistry()
	_ = reg.Register("custom", func(config *ProviderConfig) (TaskProvider, error) {
		if len(config.Providers) > 0 {
			receivedSettings = config.Providers[0].Settings
		}
		return newInMemoryProvider(), nil
	})

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{
				Name: "custom",
				Settings: map[string]string{
					"path":     "/custom/path",
					"readonly": "true",
				},
			},
		},
	}

	_, err := ResolveActiveProvider(cfg, reg)
	if err != nil {
		t.Fatalf("ResolveActiveProvider() unexpected error: %v", err)
	}

	if receivedSettings == nil {
		t.Fatal("factory did not receive settings")
	}
	if receivedSettings["path"] != "/custom/path" {
		t.Errorf("settings[path] = %q, want %q", receivedSettings["path"], "/custom/path")
	}
	if receivedSettings["readonly"] != "true" {
		t.Errorf("settings[readonly] = %q, want %q", receivedSettings["readonly"], "true")
	}
}

func TestResolveActiveProvider_FirstProviderUsedAsPrimary(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})
	_ = reg.Register("other", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{Name: "textfile"},
			{Name: "other"},
		},
	}

	provider, err := ResolveActiveProvider(cfg, reg)
	if err != nil {
		t.Fatalf("ResolveActiveProvider() unexpected error: %v", err)
	}

	// First provider should be the primary
	_, ok := provider.(*inMemoryProvider)
	if !ok {
		t.Errorf("expected *inMemoryProvider as primary, got %T", provider)
	}
}

// --- SaveProviderConfig Tests ---

func TestSaveProviderConfig_RoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &ProviderConfig{
		SchemaVersion: 1,
		Provider:      "textfile",
		NoteTitle:     "ThreeDoors Tasks",
		Theme:         "scifi",
	}

	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveProviderConfig() unexpected error: %v", err)
	}

	loaded, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}
	if loaded.Theme != "scifi" {
		t.Errorf("Theme = %q, want %q", loaded.Theme, "scifi")
	}
	if loaded.Provider != "textfile" {
		t.Errorf("Provider = %q, want %q", loaded.Provider, "textfile")
	}
}

func TestSaveProviderConfig_OverwritesExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Write initial config
	cfg := &ProviderConfig{Theme: "classic"}
	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("first save: %v", err)
	}

	// Overwrite with new theme
	cfg.Theme = "modern"
	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("second save: %v", err)
	}

	loaded, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}
	if loaded.Theme != "modern" {
		t.Errorf("Theme = %q, want %q", loaded.Theme, "modern")
	}
}

// --- Story 3.5.3 Tests: Schema Version & Migration Path ---

func TestLoadProviderConfig_SchemaVersionParsed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("schema_version: 1\nprovider: textfile\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d (migrated from 1)", cfg.SchemaVersion, CurrentSchemaVersion)
	}
}

func TestLoadProviderConfig_MissingSchemaVersion_TreatedAsZero(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Pre-versioning config (no schema_version field)
	content := []byte("provider: applenotes\nnote_title: My Tasks\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	// YAML zero value for int is 0 when field is omitempty and absent
	// The config should still load correctly regardless of schema_version
	if cfg.Provider != "applenotes" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "applenotes")
	}
	if cfg.NoteTitle != "My Tasks" {
		t.Errorf("NoteTitle = %q, want %q", cfg.NoteTitle, "My Tasks")
	}
}

func TestLoadProviderConfig_DefaultsHaveSchemaVersion(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nonexistent.yaml")

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("default SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
}

func TestGenerateSampleConfig_IncludesSchemaVersion(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	reg := NewRegistry()
	_ = reg.Register("textfile", func(config *ProviderConfig) (TaskProvider, error) {
		return newInMemoryProvider(), nil
	})

	if err := GenerateSampleConfig(configPath, reg); err != nil {
		t.Fatalf("GenerateSampleConfig() unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read generated config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, fmt.Sprintf("schema_version: %d", CurrentSchemaVersion)) {
		t.Errorf("sample config should include schema_version: %d", CurrentSchemaVersion)
	}
}

func TestLoadProviderConfig_FullConfigWithAllSections(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Simulate a config with all sections present (as a user might have)
	content := []byte(`schema_version: 1
provider: textfile
note_title: ThreeDoors Tasks
providers:
  - name: textfile
    settings:
      task_file: ~/custom/tasks.yaml
  - name: applenotes
    settings:
      note_title: Work Tasks
values:
  - "Stay focused"
  - "Build things"
onboarding_complete: true
calendar:
  enabled: true
llm:
  backend: claude
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	// Provider config sections should be parsed (v1 migrated to current)
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d (migrated)", cfg.SchemaVersion, CurrentSchemaVersion)
	}
	if len(cfg.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(cfg.Providers))
	}
	if cfg.LLM.Backend != "claude" {
		t.Errorf("LLM.Backend = %q, want %q", cfg.LLM.Backend, "claude")
	}

	// Unknown sections (values, onboarding_complete, calendar) should be silently ignored
	// by ProviderConfig — they are loaded by their own loaders
}
