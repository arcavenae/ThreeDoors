package github

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Default values for optional GitHub configuration fields.
const (
	DefaultAssignee     = "@me"
	DefaultPollInterval = 5 * time.Minute
	DefaultInProgress   = "in-progress"
)

// GitHubConfig holds parsed and validated GitHub integration settings.
type GitHubConfig struct {
	Token           string
	Repos           []string
	Assignee        string
	PollInterval    time.Duration
	PriorityLabels  map[string]string
	InProgressLabel string
}

// ParseConfig creates a GitHubConfig from a settings map with environment variable
// fallback. GITHUB_TOKEN env var takes precedence over config file settings.
//
// Required: repos (at least one, in "owner/repo" format).
// Env vars: GITHUB_TOKEN.
func ParseConfig(settings map[string]string) (*GitHubConfig, error) {
	cfg := &GitHubConfig{
		Assignee:        DefaultAssignee,
		PollInterval:    DefaultPollInterval,
		InProgressLabel: DefaultInProgress,
	}

	if settings != nil {
		cfg.Token = settings["token"]

		if v := settings["repos"]; v != "" {
			cfg.Repos = splitRepos(v)
		}
		if v := settings["assignee"]; v != "" {
			cfg.Assignee = v
		}
		if v := settings["poll_interval"]; v != "" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("github config: invalid poll_interval %q: %w", v, err)
			}
			cfg.PollInterval = d
		}
		if v := settings["in_progress_label"]; v != "" {
			cfg.InProgressLabel = v
		}

		cfg.PriorityLabels = parsePriorityLabels(settings)
	}

	// Environment variable overrides config file settings
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		cfg.Token = v
	}

	// Validate required fields
	if len(cfg.Repos) == 0 {
		return nil, fmt.Errorf("github config: repos is required (at least one owner/repo)")
	}
	for _, repo := range cfg.Repos {
		if err := validateRepoFormat(repo); err != nil {
			return nil, fmt.Errorf("github config: invalid repo %q: %w", repo, err)
		}
	}

	return cfg, nil
}

// splitRepos splits a comma-separated list of repos, trimming whitespace.
func splitRepos(s string) []string {
	parts := strings.Split(s, ",")
	repos := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			repos = append(repos, p)
		}
	}
	return repos
}

// validateRepoFormat checks that a repo string is in "owner/repo" format.
func validateRepoFormat(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("must be in owner/repo format")
	}
	return nil
}

// parsePriorityLabels extracts priority_label.* settings into a map.
func parsePriorityLabels(settings map[string]string) map[string]string {
	labels := make(map[string]string)
	for k, v := range settings {
		if strings.HasPrefix(k, "priority_label.") {
			label := strings.TrimPrefix(k, "priority_label.")
			if label != "" && v != "" {
				labels[label] = v
			}
		}
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}
