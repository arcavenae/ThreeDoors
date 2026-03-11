package todoist

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Default values for optional Todoist configuration fields.
const (
	DefaultPollInterval = 30 * time.Second
)

// AuthConfig holds authentication settings for the Todoist client.
type AuthConfig struct {
	APIToken string `yaml:"-"`
}

// TodoistConfig holds parsed and validated Todoist integration settings.
type TodoistConfig struct {
	APIToken     string `yaml:"-"`
	ProjectIDs   []string
	Filter       string
	PollInterval time.Duration
}

// ParseConfig creates a TodoistConfig from a settings map with environment variable
// fallback. The TODOIST_API_TOKEN environment variable takes precedence over config.
//
// ProjectIDs and Filter are mutually exclusive — providing both is an error.
func ParseConfig(settings map[string]string) (*TodoistConfig, error) {
	cfg := &TodoistConfig{
		PollInterval: DefaultPollInterval,
	}

	if settings != nil {
		cfg.APIToken = settings["api_token"]

		if v := settings["project_ids"]; v != "" {
			ids := strings.Split(v, ",")
			for _, id := range ids {
				trimmed := strings.TrimSpace(id)
				if trimmed != "" {
					cfg.ProjectIDs = append(cfg.ProjectIDs, trimmed)
				}
			}
		}

		if v := settings["filter"]; v != "" {
			cfg.Filter = v
		}

		if v := settings["poll_interval"]; v != "" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("todoist config: invalid poll_interval %q: %w", v, err)
			}
			cfg.PollInterval = d
		}
	}

	// Environment variable overrides config file
	if v := os.Getenv("TODOIST_API_TOKEN"); v != "" {
		cfg.APIToken = v
	}

	// Validate required fields
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("todoist config: api_token is required (set in config or TODOIST_API_TOKEN env var)")
	}

	// ProjectIDs and Filter are mutually exclusive
	if len(cfg.ProjectIDs) > 0 && cfg.Filter != "" {
		return nil, fmt.Errorf("todoist config: project_ids and filter are mutually exclusive")
	}

	return cfg, nil
}

// AuthConfigFrom returns an AuthConfig from a TodoistConfig for client construction.
func AuthConfigFrom(cfg *TodoistConfig) AuthConfig {
	return AuthConfig{APIToken: cfg.APIToken}
}
