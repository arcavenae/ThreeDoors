package tasks

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// ProviderConfig holds configuration for which task provider to use.
type ProviderConfig struct {
	Provider  string `yaml:"provider"`
	NoteTitle string `yaml:"note_title"`
}

// defaultProviderConfig returns the default configuration.
func defaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Provider:  "textfile",
		NoteTitle: "ThreeDoors Tasks",
	}
}

// LoadProviderConfig reads provider configuration from a YAML file.
// Returns defaults if the file does not exist.
func LoadProviderConfig(path string) (*ProviderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultProviderConfig(), nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return defaultProviderConfig(), nil
	}

	cfg := defaultProviderConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Ensure defaults for empty fields
	if cfg.Provider == "" {
		cfg.Provider = "textfile"
	}
	if cfg.NoteTitle == "" {
		cfg.NoteTitle = "ThreeDoors Tasks"
	}

	return cfg, nil
}
