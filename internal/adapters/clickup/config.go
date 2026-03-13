package clickup

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// Default values for optional ClickUp configuration fields.
const (
	DefaultPollInterval = 30 * time.Second
)

// DefaultStatusMapping maps common ClickUp status strings to ThreeDoors statuses.
var DefaultStatusMapping = map[string]core.TaskStatus{
	"to do":       core.StatusTodo,
	"open":        core.StatusTodo,
	"in progress": core.StatusInProgress,
	"complete":    core.StatusComplete,
	"closed":      core.StatusComplete,
	"done":        core.StatusComplete,
}

// DefaultReverseStatusMapping maps ThreeDoors statuses back to ClickUp status strings
// for bidirectional sync write-back operations.
var DefaultReverseStatusMapping = map[core.TaskStatus]string{
	core.StatusTodo:       "to do",
	core.StatusInProgress: "in progress",
	core.StatusComplete:   "complete",
	core.StatusBlocked:    "blocked",
}

// ClickUpConfig holds parsed and validated ClickUp integration settings.
type ClickUpConfig struct {
	APIToken      string                     `yaml:"-"`
	TeamID        string                     `yaml:"team_id"`
	SpaceIDs      []string                   `yaml:"space_ids"`
	ListIDs       []string                   `yaml:"list_ids"`
	Assignee      string                     `yaml:"assignee"`
	PollInterval  time.Duration              `yaml:"poll_interval"`
	StatusMapping map[string]core.TaskStatus `yaml:"status_mapping"`
	DoneStatus    string                     `yaml:"done_status"`    // ClickUp status name for "done" write-back
	BlockedStatus string                     `yaml:"blocked_status"` // ClickUp status name for "blocked" write-back
}

// ParseConfig creates a ClickUpConfig from a settings map with environment variable
// fallback. Environment variable CLICKUP_API_TOKEN takes precedence over config file
// api_token when both are set.
//
// Required: api_token (from config or env), team_id.
// At least one of space_ids or list_ids must be specified.
func ParseConfig(settings map[string]string) (*ClickUpConfig, error) {
	cfg := &ClickUpConfig{
		PollInterval: DefaultPollInterval,
	}

	// Read from settings map first
	if settings != nil {
		cfg.APIToken = settings["api_token"]
		cfg.TeamID = settings["team_id"]
		cfg.Assignee = settings["assignee"]

		if v := settings["space_ids"]; v != "" {
			cfg.SpaceIDs = splitAndTrim(v)
		}
		if v := settings["list_ids"]; v != "" {
			cfg.ListIDs = splitAndTrim(v)
		}
		if v := settings["poll_interval"]; v != "" {
			d, err := time.ParseDuration(v)
			if err != nil {
				// Try parsing as seconds
				secs, secErr := strconv.Atoi(v)
				if secErr != nil {
					return nil, fmt.Errorf("clickup config: invalid poll_interval %q: %w", v, err)
				}
				d = time.Duration(secs) * time.Second
			}
			cfg.PollInterval = d
		}
	}

	// Environment variable takes precedence over config file
	if v := os.Getenv("CLICKUP_API_TOKEN"); v != "" {
		cfg.APIToken = v
	}

	// Parse done/blocked status overrides
	if v := settings["done_status"]; v != "" {
		cfg.DoneStatus = v
	}
	if v := settings["blocked_status"]; v != "" {
		cfg.BlockedStatus = v
	}

	// Validate required fields
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("clickup config: api_token is required (set in config or CLICKUP_API_TOKEN env var)")
	}
	if cfg.TeamID == "" {
		return nil, fmt.Errorf("clickup config: team_id is required")
	}
	if len(cfg.SpaceIDs) == 0 && len(cfg.ListIDs) == 0 {
		return nil, fmt.Errorf("clickup config: at least one of space_ids or list_ids must be specified")
	}

	return cfg, nil
}

// splitAndTrim splits a comma-separated string and trims whitespace from each element.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
