package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestNewConfigCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := newConfigCmd()
	if cmd.Use != "config" {
		t.Errorf("Use = %q, want %q", cmd.Use, "config")
	}

	subCmds := cmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subCmds {
		names[sub.Name()] = true
	}

	for _, want := range []string{"show", "get", "set"} {
		if !names[want] {
			t.Errorf("missing %q subcommand", want)
		}
	}
}

func TestGetConfigValue(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion:      2,
		Provider:           "textfile",
		NoteTitle:          "My Tasks",
		Theme:              "modern",
		DevDispatchEnabled: true,
	}

	tests := []struct {
		key  string
		want string
	}{
		{"provider", "textfile"},
		{"note_title", "My Tasks"},
		{"theme", "modern"},
		{"dev_dispatch_enabled", "true"},
		{"schema_version", "2"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			got := getConfigValue(cfg, tt.key)
			if got != tt.want {
				t.Errorf("getConfigValue(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSetConfigValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		value   string
		check   func(*core.ProviderConfig) bool
		wantErr bool
	}{
		{
			name:  "set provider",
			key:   "provider",
			value: "applenotes",
			check: func(cfg *core.ProviderConfig) bool { return cfg.Provider == "applenotes" },
		},
		{
			name:  "set note_title",
			key:   "note_title",
			value: "Work Tasks",
			check: func(cfg *core.ProviderConfig) bool { return cfg.NoteTitle == "Work Tasks" },
		},
		{
			name:  "set theme",
			key:   "theme",
			value: "scifi",
			check: func(cfg *core.ProviderConfig) bool { return cfg.Theme == "scifi" },
		},
		{
			name:  "set dev_dispatch_enabled true",
			key:   "dev_dispatch_enabled",
			value: "true",
			check: func(cfg *core.ProviderConfig) bool { return cfg.DevDispatchEnabled },
		},
		{
			name:    "set dev_dispatch_enabled invalid",
			key:     "dev_dispatch_enabled",
			value:   "notabool",
			wantErr: true,
		},
		{
			name:  "set schema_version",
			key:   "schema_version",
			value: "3",
			check: func(cfg *core.ProviderConfig) bool { return cfg.SchemaVersion == 3 },
		},
		{
			name:    "set schema_version invalid",
			key:     "schema_version",
			value:   "abc",
			wantErr: true,
		},
		{
			name:    "set unknown key",
			key:     "nonexistent",
			value:   "val",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &core.ProviderConfig{}
			err := setConfigValue(cfg, tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("setConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(cfg) {
				t.Error("setConfigValue() did not set expected value")
			}
		})
	}
}

func TestConfigToMap(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion:      2,
		Provider:           "textfile",
		NoteTitle:          "ThreeDoors Tasks",
		Theme:              "classic",
		DevDispatchEnabled: false,
	}

	m := configToMap(cfg)

	if m["provider"] != "textfile" {
		t.Errorf("provider = %q, want %q", m["provider"], "textfile")
	}
	if m["schema_version"] != "2" {
		t.Errorf("schema_version = %q, want %q", m["schema_version"], "2")
	}
	if m["theme"] != "classic" {
		t.Errorf("theme = %q, want %q", m["theme"], "classic")
	}
}

func TestValidConfigKeys_UnknownReturnsError(t *testing.T) {
	t.Parallel()

	if validConfigKeys["bogus_key"] {
		t.Error("expected bogus_key to not be valid")
	}
	if !validConfigKeys["provider"] {
		t.Error("expected provider to be valid")
	}
	if !validConfigKeys["theme"] {
		t.Error("expected theme to be valid")
	}
}

func TestConfigSetAndLoad_RoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &core.ProviderConfig{
		SchemaVersion: 2,
		Provider:      "textfile",
		NoteTitle:     "ThreeDoors Tasks",
	}

	if err := setConfigValue(cfg, "theme", "modern"); err != nil {
		t.Fatalf("setConfigValue: %v", err)
	}

	if err := core.SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveProviderConfig: %v", err)
	}

	loaded, err := core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProviderConfig: %v", err)
	}

	if loaded.Theme != "modern" {
		t.Errorf("Theme = %q, want %q", loaded.Theme, "modern")
	}
}

func TestConfigShowJSON(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion:      2,
		Provider:           "textfile",
		NoteTitle:          "ThreeDoors Tasks",
		Theme:              "classic",
		DevDispatchEnabled: false,
	}

	m := configToMap(cfg)

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	if err := formatter.WriteJSON("config show", m, nil); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if env.Command != "config show" {
		t.Errorf("command = %q, want %q", env.Command, "config show")
	}
	if env.Error != nil {
		t.Errorf("unexpected error in envelope: %v", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map, got %T", env.Data)
	}
	if data["provider"] != "textfile" {
		t.Errorf("data.provider = %v, want textfile", data["provider"])
	}
}

func TestConfigGetJSON_UnknownKey(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	_ = formatter.WriteJSONError("config get", ExitValidation, "unknown config key: bogus", "")

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error in envelope")
	}
	if env.Error.Code != ExitValidation {
		t.Errorf("error code = %d, want %d", env.Error.Code, ExitValidation)
	}
}

func TestConfigSetPersists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Write initial config
	initial := &core.ProviderConfig{
		SchemaVersion: 2,
		Provider:      "textfile",
		Theme:         "classic",
	}
	if err := core.SaveProviderConfig(configPath, initial); err != nil {
		t.Fatalf("save initial: %v", err)
	}

	// Load, modify, save
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := setConfigValue(cfg, "theme", "scifi"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := core.SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify
	reloaded, err := core.LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Theme != "scifi" {
		t.Errorf("Theme = %q, want %q", reloaded.Theme, "scifi")
	}
}

func TestLoadConfig_MissingDir(t *testing.T) {
	t.Parallel()

	// loadConfig uses GetConfigDirPath which uses XDG or ~/.threedoors
	// We can't easily control this in tests, but we can verify the function
	// signature works. This is more of a smoke test.
	origHome := os.Getenv("HOME")
	t.Cleanup(func() {
		_ = os.Setenv("HOME", origHome)
	})

	dir := t.TempDir()
	_ = os.Setenv("HOME", dir)

	// loadConfig should succeed even if dir doesn't have a config file
	// (defaults are returned)
	path, cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty config path")
	}
	if cfg.Provider != "textfile" {
		t.Errorf("Provider = %q, want default %q", cfg.Provider, "textfile")
	}
}
