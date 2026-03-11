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
		return
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
		return
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
		return
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
		return
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

// --- Story 33.3: SeasonalThemes config field ---

func TestSeasonalThemesEnabled_NilDefaultsTrue(t *testing.T) {
	t.Parallel()
	cfg := &ProviderConfig{}
	if !cfg.SeasonalThemesEnabled() {
		t.Error("nil SeasonalThemes should default to true")
	}
}

func TestSeasonalThemesEnabled_ExplicitTrue(t *testing.T) {
	t.Parallel()
	v := true
	cfg := &ProviderConfig{SeasonalThemes: &v}
	if !cfg.SeasonalThemesEnabled() {
		t.Error("explicit true should return true")
	}
}

func TestSeasonalThemesEnabled_ExplicitFalse(t *testing.T) {
	t.Parallel()
	v := false
	cfg := &ProviderConfig{SeasonalThemes: &v}
	if cfg.SeasonalThemesEnabled() {
		t.Error("explicit false should return false")
	}
}

func TestLoadProviderConfig_SeasonalThemes_Roundtrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	v := false
	cfg := &ProviderConfig{
		SchemaVersion:  CurrentSchemaVersion,
		Provider:       "textfile",
		SeasonalThemes: &v,
	}
	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("load: %v", err)
		return
	}
	if loaded.SeasonalThemesEnabled() {
		t.Error("loaded config should have seasonal_themes: false")
	}
}

func TestLoadProviderConfig_SeasonalThemes_Absent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("provider: textfile\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	loaded, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("load: %v", err)
		return
	}
	if !loaded.SeasonalThemesEnabled() {
		t.Error("absent seasonal_themes should default to enabled")
	}
}

// --- Story 43.3 Tests: Config Schema v3 with Connections ---

func TestMigrateConfig_V2ToV3_WithProvidersList(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: 2,
		Providers: []ProviderEntry{
			{Name: "jira", Settings: map[string]string{"url": "https://a.atlassian.net"}},
			{Name: "todoist", Settings: map[string]string{"filter": "today"}},
		},
	}

	MigrateConfig(cfg)

	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
	if len(cfg.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].ID != "legacy-jira" {
		t.Errorf("Connections[0].ID = %q, want %q", cfg.Connections[0].ID, "legacy-jira")
	}
	if cfg.Connections[0].Provider != "jira" {
		t.Errorf("Connections[0].Provider = %q, want %q", cfg.Connections[0].Provider, "jira")
	}
	if cfg.Connections[0].Settings["url"] != "https://a.atlassian.net" {
		t.Errorf("Connections[0].Settings[url] = %q, want %q", cfg.Connections[0].Settings["url"], "https://a.atlassian.net")
	}
	if cfg.Connections[1].ID != "legacy-todoist" {
		t.Errorf("Connections[1].ID = %q, want %q", cfg.Connections[1].ID, "legacy-todoist")
	}
}

func TestMigrateConfig_V2ToV3_FlatProviderOnly(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: 2,
		Provider:      "jira",
	}

	MigrateConfig(cfg)

	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 connection from flat provider, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].Provider != "jira" {
		t.Errorf("Connections[0].Provider = %q, want %q", cfg.Connections[0].Provider, "jira")
	}
}

func TestMigrateConfig_V2ToV3_TextfileOnly_NoConnection(t *testing.T) {
	t.Parallel()

	// textfile-only configs don't get a connection entry (it's the default)
	cfg := &ProviderConfig{
		SchemaVersion: 2,
		Provider:      "textfile",
	}

	MigrateConfig(cfg)

	if len(cfg.Connections) != 0 {
		t.Errorf("expected 0 connections for textfile-only config, got %d", len(cfg.Connections))
	}
}

func TestMigrateConfig_V3_NoDoubleConversion(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: 3,
		Connections: []ConnectionConfig{
			{ID: "existing", Provider: "jira", Label: "Work Jira"},
		},
		Providers: []ProviderEntry{
			{Name: "jira"},
		},
	}

	MigrateConfig(cfg)

	// Should NOT create new connections since schema is already v3.
	if len(cfg.Connections) != 1 {
		t.Errorf("expected 1 connection (no migration), got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].ID != "existing" {
		t.Errorf("Connections[0].ID = %q, want %q", cfg.Connections[0].ID, "existing")
	}
}

func TestMigrateConfig_V1_MigratesThrough(t *testing.T) {
	t.Parallel()

	// v1 config with providers list should migrate through to v3
	cfg := &ProviderConfig{
		SchemaVersion: 1,
		Providers: []ProviderEntry{
			{Name: "obsidian", Settings: map[string]string{"vault_path": "/v"}},
		},
	}

	MigrateConfig(cfg)

	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].Provider != "obsidian" {
		t.Errorf("Connections[0].Provider = %q, want %q", cfg.Connections[0].Provider, "obsidian")
	}
}

func TestLoadProviderConfig_V3_WithConnections(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`schema_version: 3
provider: textfile
connections:
  - id: work-jira
    provider: jira
    label: Work Jira
    settings:
      url: https://company.atlassian.net
  - id: personal-todoist
    provider: todoist
    label: Personal Todoist
    settings:
      filter: today
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	if len(cfg.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].ID != "work-jira" {
		t.Errorf("Connections[0].ID = %q, want %q", cfg.Connections[0].ID, "work-jira")
	}
	if cfg.Connections[0].Label != "Work Jira" {
		t.Errorf("Connections[0].Label = %q, want %q", cfg.Connections[0].Label, "Work Jira")
	}
	if cfg.Connections[1].ID != "personal-todoist" {
		t.Errorf("Connections[1].ID = %q, want %q", cfg.Connections[1].ID, "personal-todoist")
	}
}

func TestLoadProviderConfig_V2_AutoMigratesToV3(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`schema_version: 2
providers:
  - name: jira
    settings:
      url: https://company.atlassian.net
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	if cfg.SchemaVersion != 3 {
		t.Errorf("SchemaVersion = %d, want 3", cfg.SchemaVersion)
	}
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 migrated connection, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].Settings["url"] != "https://company.atlassian.net" {
		t.Errorf("migrated connection settings lost: %v", cfg.Connections[0].Settings)
	}
}

func TestProviderConfig_AddConnection(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{SchemaVersion: 3}

	err := cfg.AddConnection(ConnectionConfig{
		ID:       "work-jira",
		Provider: "jira",
		Label:    "Work Jira",
		Settings: map[string]string{"url": "https://a.atlassian.net"},
	})
	if err != nil {
		t.Fatalf("AddConnection() unexpected error: %v", err)
	}

	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].ID != "work-jira" {
		t.Errorf("ID = %q, want %q", cfg.Connections[0].ID, "work-jira")
	}
}

func TestProviderConfig_AddConnection_DuplicateID(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: 3,
		Connections: []ConnectionConfig{
			{ID: "work-jira", Provider: "jira", Label: "Work Jira"},
		},
	}

	err := cfg.AddConnection(ConnectionConfig{
		ID:       "work-jira",
		Provider: "jira",
		Label:    "Other Jira",
	})
	if err == nil {
		t.Error("AddConnection() expected error for duplicate ID, got nil")
	}
}

func TestProviderConfig_AddConnection_EmptyID(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{SchemaVersion: 3}
	err := cfg.AddConnection(ConnectionConfig{Provider: "jira", Label: "x"})
	if err == nil {
		t.Error("AddConnection() expected error for empty ID, got nil")
	}
}

func TestProviderConfig_AddConnection_EmptyProvider(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{SchemaVersion: 3}
	err := cfg.AddConnection(ConnectionConfig{ID: "x", Label: "x"})
	if err == nil {
		t.Error("AddConnection() expected error for empty provider, got nil")
	}
}

func TestProviderConfig_RemoveConnection(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: 3,
		Connections: []ConnectionConfig{
			{ID: "a", Provider: "jira", Label: "A"},
			{ID: "b", Provider: "todoist", Label: "B"},
		},
	}

	if err := cfg.RemoveConnection("a"); err != nil {
		t.Fatalf("RemoveConnection() unexpected error: %v", err)
	}

	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 connection after removal, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].ID != "b" {
		t.Errorf("remaining connection ID = %q, want %q", cfg.Connections[0].ID, "b")
	}
}

func TestProviderConfig_RemoveConnection_NotFound(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{SchemaVersion: 3}
	err := cfg.RemoveConnection("nonexistent")
	if err == nil {
		t.Error("RemoveConnection() expected error for missing ID, got nil")
	}
}

func TestProviderConfig_GetConnection(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: 3,
		Connections: []ConnectionConfig{
			{ID: "a", Provider: "jira", Label: "A"},
			{ID: "b", Provider: "todoist", Label: "B"},
		},
	}

	conn := cfg.GetConnection("b")
	if conn == nil {
		t.Fatal("GetConnection() returned nil for existing ID")
		return
	}
	if conn.Provider != "todoist" {
		t.Errorf("Provider = %q, want %q", conn.Provider, "todoist")
	}
}

func TestProviderConfig_GetConnection_NotFound(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{SchemaVersion: 3}
	conn := cfg.GetConnection("nonexistent")
	if conn != nil {
		t.Errorf("GetConnection() expected nil for missing ID, got %+v", conn)
	}
}

func TestProviderConfig_UpdateConnection(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: 3,
		Connections: []ConnectionConfig{
			{ID: "a", Provider: "jira", Label: "Old Label"},
		},
	}

	err := cfg.UpdateConnection(ConnectionConfig{
		ID:       "a",
		Provider: "jira",
		Label:    "New Label",
		Settings: map[string]string{"url": "https://new.atlassian.net"},
	})
	if err != nil {
		t.Fatalf("UpdateConnection() unexpected error: %v", err)
	}

	if cfg.Connections[0].Label != "New Label" {
		t.Errorf("Label = %q, want %q", cfg.Connections[0].Label, "New Label")
	}
	if cfg.Connections[0].Settings["url"] != "https://new.atlassian.net" {
		t.Errorf("Settings[url] = %q, want %q", cfg.Connections[0].Settings["url"], "https://new.atlassian.net")
	}
}

func TestProviderConfig_UpdateConnection_NotFound(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{SchemaVersion: 3}
	err := cfg.UpdateConnection(ConnectionConfig{ID: "missing", Provider: "jira"})
	if err == nil {
		t.Error("UpdateConnection() expected error for missing ID, got nil")
	}
}

func TestProviderConfig_MultipleConnectionsSameProvider(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{SchemaVersion: 3}

	_ = cfg.AddConnection(ConnectionConfig{
		ID: "work-jira", Provider: "jira", Label: "Work Jira",
		Settings: map[string]string{"url": "https://work.atlassian.net"},
	})
	_ = cfg.AddConnection(ConnectionConfig{
		ID: "personal-jira", Provider: "jira", Label: "Personal Jira",
		Settings: map[string]string{"url": "https://personal.atlassian.net"},
	})

	if len(cfg.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].ID == cfg.Connections[1].ID {
		t.Error("connections should have distinct IDs")
	}
	if cfg.Connections[0].Label == cfg.Connections[1].Label {
		t.Error("connections should have distinct labels")
	}
}

func TestConnectionConfig_RoundTrip_SaveLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &ProviderConfig{
		SchemaVersion: 3,
		Provider:      "textfile",
		NoteTitle:     "ThreeDoors Tasks",
		Connections: []ConnectionConfig{
			{
				ID:       "work-jira",
				Provider: "jira",
				Label:    "Work Jira",
				Settings: map[string]string{"url": "https://a.atlassian.net"},
			},
			{
				ID:       "personal-todoist",
				Provider: "todoist",
				Label:    "Personal Todoist",
				Settings: map[string]string{"filter": "today"},
			},
		},
	}

	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveProviderConfig() unexpected error: %v", err)
	}

	loaded, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig() unexpected error: %v", err)
	}

	if len(loaded.Connections) != 2 {
		t.Fatalf("expected 2 connections after round-trip, got %d", len(loaded.Connections))
	}
	if loaded.Connections[0].ID != "work-jira" {
		t.Errorf("Connections[0].ID = %q, want %q", loaded.Connections[0].ID, "work-jira")
	}
	if loaded.Connections[0].Settings["url"] != "https://a.atlassian.net" {
		t.Errorf("Connections[0].Settings[url] lost after round-trip")
	}
	if loaded.Connections[1].ID != "personal-todoist" {
		t.Errorf("Connections[1].ID = %q, want %q", loaded.Connections[1].ID, "personal-todoist")
	}
}

func TestConnectionConfig_NoCredentialsInSavedYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &ProviderConfig{
		SchemaVersion: 3,
		Connections: []ConnectionConfig{
			{
				ID:       "work-jira",
				Provider: "jira",
				Label:    "Work Jira",
				Settings: map[string]string{"url": "https://a.atlassian.net"},
			},
		},
	}

	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	content := string(data)
	for _, keyword := range []string{"password", "secret", "token", "api_key", "credential"} {
		if strings.Contains(strings.ToLower(content), keyword) {
			t.Errorf("saved YAML contains credential keyword %q", keyword)
		}
	}
}

func TestProviderConfig_ExistingFieldsPreserved_AfterV3Migration(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`schema_version: 2
provider: textfile
note_title: My Tasks
theme: scifi
dev_dispatch_enabled: true
providers:
  - name: jira
    settings:
      url: https://a.atlassian.net
llm:
  backend: claude
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// All existing fields should be preserved after migration.
	if cfg.Provider != "textfile" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "textfile")
	}
	if cfg.NoteTitle != "My Tasks" {
		t.Errorf("NoteTitle = %q, want %q", cfg.NoteTitle, "My Tasks")
	}
	if cfg.Theme != "scifi" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "scifi")
	}
	if !cfg.DevDispatchEnabled {
		t.Error("DevDispatchEnabled should be true")
	}
	if cfg.LLM.Backend != "claude" {
		t.Errorf("LLM.Backend = %q, want %q", cfg.LLM.Backend, "claude")
	}
	if len(cfg.Providers) != 1 {
		t.Errorf("Providers should be preserved, got %d", len(cfg.Providers))
	}
	// And connections should be migrated
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 migrated connection, got %d", len(cfg.Connections))
	}
}
