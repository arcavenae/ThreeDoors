package core

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// OnboardingConfig holds onboarding state persisted to config.yaml.
type OnboardingConfig struct {
	OnboardingComplete bool `yaml:"onboarding_complete"`
}

// IsFirstRun checks whether onboarding has been completed by reading config.yaml.
// Returns true if onboarding has not been completed (or config doesn't exist).
func IsFirstRun(configDir string) bool {
	data, err := ReadFileWithLimit(configDir+"/config.yaml", MaxConfigFileSize)
	if err != nil {
		return true
	}
	if len(data) == 0 {
		return true
	}

	var cfg OnboardingConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return true
	}
	return !cfg.OnboardingComplete
}

// MarkOnboardingComplete persists onboarding_complete: true to config.yaml.
// It preserves existing config fields by reading, merging, and rewriting.
func MarkOnboardingComplete(configDir string) error {
	configPath := configDir + "/config.yaml"

	// Read existing config as a generic map to preserve all fields
	existing := make(map[string]interface{})
	data, err := ReadFileWithLimit(configPath, MaxConfigFileSize)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read config: %w", err)
	}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	}

	existing["onboarding_complete"] = true

	out, err := yaml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmpPath := configPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
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
