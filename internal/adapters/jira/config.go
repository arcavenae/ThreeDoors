package jira

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Default values for optional Jira configuration fields.
const (
	DefaultJQL          = "assignee = currentUser() AND statusCategory != Done"
	DefaultMaxResults   = 50
	DefaultPollInterval = 30 * time.Second
)

// JiraConfig holds parsed and validated Jira integration settings.
type JiraConfig struct {
	URL          string
	AuthType     AuthType
	Email        string
	APIToken     string `yaml:"-"`
	JQL          string
	MaxResults   int
	PollInterval time.Duration
}

// ParseConfig creates a JiraConfig from a settings map with environment variable
// fallback. Environment variables take precedence over config file settings.
//
// Required: url, auth_type (from config or env).
// Env vars: JIRA_URL, JIRA_EMAIL, JIRA_API_TOKEN.
func ParseConfig(settings map[string]string) (*JiraConfig, error) {
	cfg := &JiraConfig{
		JQL:          DefaultJQL,
		MaxResults:   DefaultMaxResults,
		PollInterval: DefaultPollInterval,
	}

	// Read from settings map first
	if settings != nil {
		cfg.URL = settings["url"]
		cfg.Email = settings["email"]
		cfg.APIToken = settings["api_token"]

		if v := settings["auth_type"]; v != "" {
			cfg.AuthType = AuthType(v)
		}
		if v := settings["jql"]; v != "" {
			cfg.JQL = v
		}
		if v := settings["max_results"]; v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("jira config: invalid max_results %q: %w", v, err)
			}
			cfg.MaxResults = n
		}
		if v := settings["poll_interval"]; v != "" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("jira config: invalid poll_interval %q: %w", v, err)
			}
			cfg.PollInterval = d
		}
	}

	// Environment variables override config file settings
	if v := os.Getenv("JIRA_URL"); v != "" {
		cfg.URL = v
	}
	if v := os.Getenv("JIRA_EMAIL"); v != "" {
		cfg.Email = v
	}
	if v := os.Getenv("JIRA_API_TOKEN"); v != "" {
		cfg.APIToken = v
	}

	// Validate required fields
	if cfg.URL == "" {
		return nil, fmt.Errorf("jira config: url is required (set in config or JIRA_URL env var)")
	}
	if err := validateURL(cfg.URL); err != nil {
		return nil, fmt.Errorf("jira config: invalid url: %w", err)
	}

	if cfg.AuthType == "" {
		return nil, fmt.Errorf("jira config: auth_type is required (must be %q or %q)", AuthBasic, AuthPAT)
	}
	if cfg.AuthType != AuthBasic && cfg.AuthType != AuthPAT {
		return nil, fmt.Errorf("jira config: invalid auth_type %q (must be %q or %q)", cfg.AuthType, AuthBasic, AuthPAT)
	}

	return cfg, nil
}

// validateURL checks that a URL has both a scheme and a host.
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("missing scheme (e.g. https://)")
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	if !strings.HasPrefix(u.Scheme, "http") {
		return fmt.Errorf("scheme must be http or https, got %q", u.Scheme)
	}
	return nil
}
