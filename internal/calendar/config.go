package calendar

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SourceType identifies a calendar source backend.
type SourceType string

const (
	SourceAppleScript SourceType = "applescript"
	SourceICS         SourceType = "ics"
	SourceCalDAVCache SourceType = "caldav_cache"
)

// SourceConfig describes a single calendar source.
type SourceConfig struct {
	Type SourceType `yaml:"type"`
	Path string     `yaml:"path,omitempty"`
}

// Config holds calendar configuration from config.yaml.
type Config struct {
	Enabled bool           `yaml:"enabled"`
	Sources []SourceConfig `yaml:"sources"`
}

// Validate checks that the configuration is consistent.
func (c *Config) Validate() error {
	for i, src := range c.Sources {
		switch src.Type {
		case SourceAppleScript:
			// No path needed.
		case SourceICS:
			if src.Path == "" {
				return fmt.Errorf("calendar source %d (ics): path is required", i)
			}
		case SourceCalDAVCache:
			// Path is optional; defaults to ~/Library/Calendars/.
		default:
			return fmt.Errorf("calendar source %d: unknown type %q", i, src.Type)
		}
	}
	return nil
}

// fullConfig mirrors the top-level config.yaml structure to extract the calendar section.
type fullConfig struct {
	Calendar Config `yaml:"calendar"`
}

// LoadConfig reads the calendar configuration from a config.yaml file.
// Returns a disabled config if the file does not exist or has no calendar section.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read calendar config %s: %w", path, err)
	}

	if len(data) == 0 {
		return &Config{}, nil
	}

	var fc fullConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse calendar config: %w", err)
	}

	if err := fc.Calendar.Validate(); err != nil {
		return nil, fmt.Errorf("validate calendar config: %w", err)
	}

	return &fc.Calendar, nil
}
