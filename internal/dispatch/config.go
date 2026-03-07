package dispatch

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DispatchConfig holds configuration for the dev dispatch pipeline.
type DispatchConfig struct {
	// RequireStory controls whether story files are generated before dispatching.
	// Defaults to false — when false, workers receive raw task descriptions.
	RequireStory bool `yaml:"require_story" json:"require_story"`
}

// DevDispatchConfig holds guardrail settings for the dev dispatch pipeline.
type DevDispatchConfig struct {
	Enabled         bool `yaml:"enabled"`
	MaxConcurrent   int  `yaml:"max_concurrent"`
	AutoDispatch    bool `yaml:"auto_dispatch"`
	CooldownMinutes int  `yaml:"cooldown_minutes"`
	DailyLimit      int  `yaml:"daily_limit"`
	RequireStory    bool `yaml:"require_story"`
}

// DefaultDevDispatchConfig returns config with safe defaults.
func DefaultDevDispatchConfig() DevDispatchConfig {
	return DevDispatchConfig{
		Enabled:         false,
		MaxConcurrent:   2,
		AutoDispatch:    false,
		CooldownMinutes: 5,
		DailyLimit:      10,
		RequireStory:    false,
	}
}

// configFile is the top-level structure for ~/.threedoors/config.yaml
// that includes the dev_dispatch section.
type configFile struct {
	DevDispatch *DevDispatchConfig `yaml:"dev_dispatch,omitempty"`
}

// LoadDevDispatchConfig reads the dev_dispatch section from a config.yaml file.
// Returns defaults if the file does not exist or has no dev_dispatch section.
func LoadDevDispatchConfig(path string) (DevDispatchConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultDevDispatchConfig(), nil
		}
		return DevDispatchConfig{}, fmt.Errorf("read dispatch config: %w", err)
	}

	if len(data) == 0 {
		return DefaultDevDispatchConfig(), nil
	}

	var cf configFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return DevDispatchConfig{}, fmt.Errorf("parse dispatch config: %w", err)
	}

	if cf.DevDispatch == nil {
		return DefaultDevDispatchConfig(), nil
	}

	cfg := *cf.DevDispatch
	applyDefaults(&cfg)
	return cfg, nil
}

// applyDefaults fills zero-value fields with safe defaults.
func applyDefaults(cfg *DevDispatchConfig) {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 2
	}
	if cfg.CooldownMinutes <= 0 {
		cfg.CooldownMinutes = 5
	}
	if cfg.DailyLimit <= 0 {
		cfg.DailyLimit = 10
	}
}
