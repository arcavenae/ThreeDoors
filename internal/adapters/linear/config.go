package linear

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Default values for optional Linear configuration fields.
const (
	DefaultPollInterval = 5 * time.Minute
)

// LinearConfig holds parsed and validated Linear integration settings.
type LinearConfig struct {
	APIKey       string
	TeamIDs      []string
	Assignee     string
	PollInterval time.Duration
}

// ParseConfig creates a LinearConfig from a settings map with environment variable
// fallback. LINEAR_API_KEY env var takes precedence over config file api_key.
//
// Required: api_key (via config or env var), team_ids (at least one).
func ParseConfig(settings map[string]string) (*LinearConfig, error) {
	cfg := &LinearConfig{
		PollInterval: DefaultPollInterval,
	}

	if settings != nil {
		cfg.APIKey = settings["api_key"]

		if v := settings["team_ids"]; v != "" {
			ids := strings.Split(v, ",")
			for _, id := range ids {
				trimmed := strings.TrimSpace(id)
				if trimmed != "" {
					cfg.TeamIDs = append(cfg.TeamIDs, trimmed)
				}
			}
		}

		if v := settings["assignee"]; v != "" {
			cfg.Assignee = v
		}

		if v := settings["poll_interval"]; v != "" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("linear config: invalid poll_interval %q: %w", v, err)
			}
			cfg.PollInterval = d
		}
	}

	// Environment variable overrides config file (AC4)
	if v := os.Getenv("LINEAR_API_KEY"); v != "" {
		cfg.APIKey = v
	}

	// Validate required fields
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("linear config: api_key is required (set in config or LINEAR_API_KEY env var)")
	}
	if len(cfg.TeamIDs) == 0 {
		return nil, fmt.Errorf("linear config: team_ids is required (at least one team ID)")
	}

	return cfg, nil
}
