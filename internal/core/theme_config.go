package core

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SaveThemeConfig persists the selected theme to config.yaml.
// It preserves existing config fields by reading, merging, and rewriting
// using atomic write (write to .tmp, sync, rename).
func SaveThemeConfig(configDir string, themeName string) error {
	configPath := configDir + "/config.yaml"

	existing := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read config: %w", err)
	}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	}

	existing["theme"] = themeName

	out, err := yaml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmpPath := configPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := f.Write(out); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync temp file: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, configPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
