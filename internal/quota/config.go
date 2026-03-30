package quota

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLConfig is the on-disk representation of the warning engine configuration.
// It maps directly to YAML config files and environment variable overrides.
type YAMLConfig struct {
	PlanType string     `yaml:"plan_type"` // e.g., "max_5x", "max_20x"
	Tiers    []YAMLTier `yaml:"tiers"`
	Peak     YAMLPeak   `yaml:"peak"`
	Notify   string     `yaml:"notify"` // "cli" or "multiclaude"
}

// YAMLTier represents a single threshold tier in YAML configuration.
type YAMLTier struct {
	Percent    float64 `yaml:"percent"`
	Label      string  `yaml:"label"`
	Suggestion string  `yaml:"suggestion"`
}

// YAMLPeak represents peak hour configuration in YAML.
type YAMLPeak struct {
	StartHour   int     `yaml:"start_hour"`
	EndHour     int     `yaml:"end_hour"`
	ShiftFactor float64 `yaml:"shift_factor"`
}

// LoadConfigFromFile reads a YAML configuration file and returns a ThresholdConfig.
func LoadConfigFromFile(path string) (ThresholdConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ThresholdConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}
	return ParseConfig(data)
}

// ParseConfig parses YAML bytes into a ThresholdConfig.
// Missing fields fall back to defaults.
func ParseConfig(data []byte) (ThresholdConfig, error) {
	var yc YAMLConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return ThresholdConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return yc.ToThresholdConfig(), nil
}

// ToThresholdConfig converts a YAMLConfig to a ThresholdConfig, applying defaults
// for any missing values.
func (yc YAMLConfig) ToThresholdConfig() ThresholdConfig {
	cfg := DefaultThresholdConfig()

	if len(yc.Tiers) > 0 {
		cfg.Tiers = make([]Tier, len(yc.Tiers))
		for i, yt := range yc.Tiers {
			cfg.Tiers[i] = Tier(yt)
		}
	}

	if yc.Peak.StartHour > 0 {
		cfg.PeakStartHour = yc.Peak.StartHour
	}
	if yc.Peak.EndHour > 0 {
		cfg.PeakEndHour = yc.Peak.EndHour
	}
	if yc.Peak.ShiftFactor > 0 {
		cfg.PeakShiftFactor = yc.Peak.ShiftFactor
	}
	if yc.Notify == "cli" {
		cfg.NotifyViaCLI = true
	}

	return cfg
}

// LoadConfigFromEnv reads configuration from environment variables.
// Supported variables:
//   - QUOTA_PLAN_TYPE: plan type (max_5x, max_20x)
//   - QUOTA_NOTIFY: notification method (cli, multiclaude)
//   - QUOTA_PEAK_SHIFT: peak shift factor (float)
//
// For full tier customization, use YAML config file.
func LoadConfigFromEnv() ThresholdConfig {
	cfg := DefaultThresholdConfig()

	if notify := os.Getenv("QUOTA_NOTIFY"); notify == "cli" {
		cfg.NotifyViaCLI = true
	}

	if shift := os.Getenv("QUOTA_PEAK_SHIFT"); shift != "" {
		var f float64
		if _, err := fmt.Sscanf(shift, "%f", &f); err == nil && f > 0 && f <= 1 {
			cfg.PeakShiftFactor = f
		}
	}

	return cfg
}
