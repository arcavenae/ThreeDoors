package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestConfigShowHumanOutput_AllFields(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion:      2,
		Provider:           "textfile",
		NoteTitle:          "My Tasks",
		Theme:              "modern",
		DevDispatchEnabled: true,
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)
	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "KEY\tVALUE\n")
	_, _ = fmt.Fprintf(tw, "schema_version\t%d\n", cfg.SchemaVersion)
	_, _ = fmt.Fprintf(tw, "provider\t%s\n", cfg.Provider)
	_, _ = fmt.Fprintf(tw, "note_title\t%s\n", cfg.NoteTitle)
	_, _ = fmt.Fprintf(tw, "theme\t%s\n", cfg.Theme)
	_, _ = fmt.Fprintf(tw, "dev_dispatch_enabled\t%t\n", cfg.DevDispatchEnabled)
	_ = tw.Flush()

	output := buf.String()
	for _, want := range []string{
		"KEY", "VALUE",
		"schema_version", "2",
		"provider", "textfile",
		"note_title", "My Tasks",
		"theme", "modern",
		"dev_dispatch_enabled", "true",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("table output missing %q, got:\n%s", want, output)
		}
	}
}

func TestConfigShowJSON_AllFieldsPresent(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion:      2,
		Provider:           "textfile",
		NoteTitle:          "ThreeDoors Tasks",
		Theme:              "classic",
		DevDispatchEnabled: true,
	}

	m := configToMap(cfg)

	tests := []struct {
		key  string
		want string
	}{
		{"schema_version", "2"},
		{"provider", "textfile"},
		{"note_title", "ThreeDoors Tasks"},
		{"theme", "classic"},
		{"dev_dispatch_enabled", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			got, ok := m[tt.key]
			if !ok {
				t.Errorf("configToMap missing key %q", tt.key)
				return
			}
			if got != tt.want {
				t.Errorf("configToMap[%q] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestConfigGetJSON_ValidKeys(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion:      2,
		Provider:           "textfile",
		NoteTitle:          "My Tasks",
		Theme:              "modern",
		DevDispatchEnabled: false,
	}

	tests := []struct {
		key  string
		want string
	}{
		{"provider", "textfile"},
		{"note_title", "My Tasks"},
		{"theme", "modern"},
		{"dev_dispatch_enabled", "false"},
		{"schema_version", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			value := getConfigValue(cfg, tt.key)
			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, true)
			if err := formatter.WriteJSON("config get", map[string]string{"key": tt.key, "value": value}, nil); err != nil {
				t.Fatalf("WriteJSON: %v", err)
			}

			var env JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if env.Command != "config get" {
				t.Errorf("command = %q, want %q", env.Command, "config get")
			}

			data, ok := env.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("data not a map: %T", env.Data)
			}
			if data["key"] != tt.key {
				t.Errorf("key = %v, want %q", data["key"], tt.key)
			}
			if data["value"] != tt.want {
				t.Errorf("value = %v, want %q", data["value"], tt.want)
			}
		})
	}
}

func TestConfigSetJSON_SuccessEnvelope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"set provider", "provider", "applenotes"},
		{"set theme", "theme", "scifi"},
		{"set note_title", "note_title", "Work"},
		{"set dev_dispatch true", "dev_dispatch_enabled", "true"},
		{"set schema_version", "schema_version", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &core.ProviderConfig{SchemaVersion: 2, Provider: "textfile"}
			if err := setConfigValue(cfg, tt.key, tt.value); err != nil {
				t.Fatalf("setConfigValue: %v", err)
			}

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, true)
			if err := formatter.WriteJSON("config set", map[string]string{"key": tt.key, "value": tt.value}, nil); err != nil {
				t.Fatalf("WriteJSON: %v", err)
			}

			var env JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if env.Command != "config set" {
				t.Errorf("command = %q, want %q", env.Command, "config set")
			}
			if env.Error != nil {
				t.Errorf("unexpected error: %v", env.Error)
			}
		})
	}
}

func TestConfigSetJSON_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		value   string
		wantMsg string
	}{
		{"invalid boolean", "dev_dispatch_enabled", "notabool", "invalid boolean"},
		{"invalid integer", "schema_version", "abc", "invalid integer"},
		{"unknown key", "nonexistent", "val", "unknown config key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &core.ProviderConfig{}
			err := setConfigValue(cfg, tt.key, tt.value)
			if err == nil {
				t.Fatal("expected error")
			}

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, true)
			_ = formatter.WriteJSONError("config set", ExitValidation, err.Error(), "")

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
			if !strings.Contains(env.Error.Message, tt.wantMsg) {
				t.Errorf("error message = %q, want containing %q", env.Error.Message, tt.wantMsg)
			}
		})
	}
}

func TestConfigRoundTrip_AllKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &core.ProviderConfig{
		SchemaVersion: 2,
		Provider:      "textfile",
	}

	updates := []struct {
		key   string
		value string
		check func(*core.ProviderConfig) string
	}{
		{"provider", "applenotes", func(c *core.ProviderConfig) string { return c.Provider }},
		{"note_title", "Work Tasks", func(c *core.ProviderConfig) string { return c.NoteTitle }},
		{"theme", "scifi", func(c *core.ProviderConfig) string { return c.Theme }},
	}

	for _, u := range updates {
		t.Run(u.key, func(t *testing.T) {
			if err := setConfigValue(cfg, u.key, u.value); err != nil {
				t.Fatalf("setConfigValue(%q, %q): %v", u.key, u.value, err)
			}
			if err := core.SaveProviderConfig(configPath, cfg); err != nil {
				t.Fatalf("save: %v", err)
			}
			loaded, err := core.LoadProviderConfig(configPath)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			got := u.check(loaded)
			if got != u.value {
				t.Errorf("after round-trip, %s = %q, want %q", u.key, got, u.value)
			}
			cfg = loaded
		})
	}
}

func TestConfigShowHumanOutput_DefaultValues(t *testing.T) {
	t.Parallel()

	origHome := os.Getenv("HOME")
	t.Cleanup(func() {
		_ = os.Setenv("HOME", origHome)
	})

	dir := t.TempDir()
	_ = os.Setenv("HOME", dir)

	_, cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)
	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "KEY\tVALUE\n")
	_, _ = fmt.Fprintf(tw, "provider\t%s\n", cfg.Provider)
	_ = tw.Flush()

	output := buf.String()
	if !strings.Contains(output, "textfile") {
		t.Errorf("default provider should be textfile, got:\n%s", output)
	}
}
