package core

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// InlineHintsConfig holds the inline key hints configuration.
type InlineHintsConfig struct {
	ShowInlineHints bool `yaml:"show_inline_hints"`
	SessionCount    int  `yaml:"inline_hints_session_count"`
	FadeThreshold   int  `yaml:"inline_hints_fade_threshold"`
}

// DefaultInlineHintsConfig returns the default inline hints configuration.
func DefaultInlineHintsConfig() *InlineHintsConfig {
	return &InlineHintsConfig{
		ShowInlineHints: true,
		SessionCount:    0,
		FadeThreshold:   5,
	}
}

// ResolveInlineHintState determines whether inline hints should be shown and
// whether fade mode is active based on the current configuration.
// Returns (enabled, fade). When session count >= threshold, enabled is false.
// When session count == threshold-1, fade is true (visual cue before auto-disable).
func ResolveInlineHintState(cfg *InlineHintsConfig) (enabled bool, fade bool) {
	if !cfg.ShowInlineHints {
		return false, false
	}
	if cfg.FadeThreshold > 0 && cfg.SessionCount >= cfg.FadeThreshold {
		return false, false
	}
	if cfg.FadeThreshold > 0 && cfg.SessionCount == cfg.FadeThreshold-1 {
		return true, true
	}
	return true, false
}

// LoadInlineHintsConfig reads inline hints configuration from config.yaml.
// Returns default config if the file does not exist or lacks inline hints fields.
func LoadInlineHintsConfig(configDir string) (*InlineHintsConfig, error) {
	configPath := configDir + "/config.yaml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultInlineHintsConfig(), nil
		}
		return nil, fmt.Errorf("read config for inline hints: %w", err)
	}

	if len(data) == 0 {
		return DefaultInlineHintsConfig(), nil
	}

	raw := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config for inline hints: %w", err)
	}

	cfg := DefaultInlineHintsConfig()

	if v, ok := raw["show_inline_hints"]; ok {
		if b, ok := v.(bool); ok {
			cfg.ShowInlineHints = b
		}
	}
	if v, ok := raw["inline_hints_session_count"]; ok {
		if n, ok := v.(int); ok {
			cfg.SessionCount = n
		}
	}
	if v, ok := raw["inline_hints_fade_threshold"]; ok {
		if n, ok := v.(int); ok {
			cfg.FadeThreshold = n
		}
	}

	return cfg, nil
}

// SaveInlineHintsConfig persists inline hints configuration to config.yaml
// using atomic write. Preserves existing config fields.
func SaveInlineHintsConfig(configDir string, cfg *InlineHintsConfig) error {
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

	existing["show_inline_hints"] = cfg.ShowInlineHints
	existing["inline_hints_session_count"] = cfg.SessionCount
	existing["inline_hints_fade_threshold"] = cfg.FadeThreshold

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
