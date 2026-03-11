package core

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSaveThemeConfig_NewFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := SaveThemeConfig(dir, "scifi"); err != nil {
		t.Fatalf("SaveThemeConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
		return
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cfg["theme"] != "scifi" {
		t.Errorf("theme = %v, want scifi", cfg["theme"])
	}
}

func TestSaveThemeConfig_PreservesExistingFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	existing := []byte("provider: textfile\nonboarding_complete: true\n")
	if err := os.WriteFile(configPath, existing, 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	if err := SaveThemeConfig(dir, "classic"); err != nil {
		t.Fatalf("SaveThemeConfig: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
		return
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cfg["theme"] != "classic" {
		t.Errorf("theme = %v, want classic", cfg["theme"])
	}
	if cfg["provider"] != "textfile" {
		t.Errorf("provider = %v, want textfile", cfg["provider"])
	}
	if cfg["onboarding_complete"] != true {
		t.Errorf("onboarding_complete = %v, want true", cfg["onboarding_complete"])
	}
}

func TestSaveThemeConfig_OverwritesPreviousTheme(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	existing := []byte("theme: modern\n")
	if err := os.WriteFile(configPath, existing, 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	if err := SaveThemeConfig(dir, "shoji"); err != nil {
		t.Fatalf("SaveThemeConfig: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
		return
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cfg["theme"] != "shoji" {
		t.Errorf("theme = %v, want shoji", cfg["theme"])
	}
}
