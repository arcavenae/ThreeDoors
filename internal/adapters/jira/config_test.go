package jira

import (
	"testing"
	"time"
)

func TestParseConfig_ValidBasic(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"url":       "https://company.atlassian.net",
		"auth_type": "basic",
		"email":     "user@example.com",
		"api_token": "tok123",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.URL != "https://company.atlassian.net" {
		t.Errorf("URL = %q, want %q", cfg.URL, "https://company.atlassian.net")
	}
	if cfg.AuthType != AuthBasic {
		t.Errorf("AuthType = %q, want %q", cfg.AuthType, AuthBasic)
	}
	if cfg.Email != "user@example.com" {
		t.Errorf("Email = %q, want %q", cfg.Email, "user@example.com")
	}
	if cfg.APIToken != "tok123" {
		t.Errorf("APIToken = %q, want %q", cfg.APIToken, "tok123")
	}
}

func TestParseConfig_ValidPAT(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"url":       "https://jira.corp.com",
		"auth_type": "pat",
		"api_token": "pat-secret",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.AuthType != AuthPAT {
		t.Errorf("AuthType = %q, want %q", cfg.AuthType, AuthPAT)
	}
}

func TestParseConfig_Defaults(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"url":       "https://company.atlassian.net",
		"auth_type": "basic",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.JQL != DefaultJQL {
		t.Errorf("JQL = %q, want default %q", cfg.JQL, DefaultJQL)
	}
	if cfg.MaxResults != DefaultMaxResults {
		t.Errorf("MaxResults = %d, want default %d", cfg.MaxResults, DefaultMaxResults)
	}
	if cfg.PollInterval != DefaultPollInterval {
		t.Errorf("PollInterval = %v, want default %v", cfg.PollInterval, DefaultPollInterval)
	}
}

func TestParseConfig_CustomOptionals(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"url":           "https://company.atlassian.net",
		"auth_type":     "basic",
		"jql":           "project = FOO",
		"max_results":   "100",
		"poll_interval": "1m",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.JQL != "project = FOO" {
		t.Errorf("JQL = %q, want %q", cfg.JQL, "project = FOO")
	}
	if cfg.MaxResults != 100 {
		t.Errorf("MaxResults = %d, want 100", cfg.MaxResults)
	}
	if cfg.PollInterval != time.Minute {
		t.Errorf("PollInterval = %v, want 1m", cfg.PollInterval)
	}
}

func TestParseConfig_EnvVarOverride(t *testing.T) {
	t.Setenv("JIRA_URL", "https://env.atlassian.net")
	t.Setenv("JIRA_EMAIL", "env@example.com")
	t.Setenv("JIRA_API_TOKEN", "env-token")

	settings := map[string]string{
		"url":       "https://config.atlassian.net",
		"auth_type": "basic",
		"email":     "config@example.com",
		"api_token": "config-token",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.URL != "https://env.atlassian.net" {
		t.Errorf("URL = %q, want env override %q", cfg.URL, "https://env.atlassian.net")
	}
	if cfg.Email != "env@example.com" {
		t.Errorf("Email = %q, want env override %q", cfg.Email, "env@example.com")
	}
	if cfg.APIToken != "env-token" {
		t.Errorf("APIToken = %q, want env override %q", cfg.APIToken, "env-token")
	}
}

func TestParseConfig_EnvVarOnly(t *testing.T) {
	t.Setenv("JIRA_URL", "https://envonly.atlassian.net")
	t.Setenv("JIRA_API_TOKEN", "envonly-token")

	settings := map[string]string{
		"auth_type": "pat",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.URL != "https://envonly.atlassian.net" {
		t.Errorf("URL = %q, want %q", cfg.URL, "https://envonly.atlassian.net")
	}
	if cfg.APIToken != "envonly-token" {
		t.Errorf("APIToken = %q, want %q", cfg.APIToken, "envonly-token")
	}
}

func TestParseConfig_NilSettings(t *testing.T) {
	t.Setenv("JIRA_URL", "https://nil.atlassian.net")

	_, err := ParseConfig(nil)
	if err == nil {
		t.Fatal("ParseConfig(nil) should fail when auth_type is missing")
	}
}

func TestParseConfig_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings map[string]string
		wantErr  string
	}{
		{
			name:     "missing url",
			settings: map[string]string{"auth_type": "basic"},
			wantErr:  "url is required",
		},
		{
			name:     "invalid url no scheme",
			settings: map[string]string{"url": "company.atlassian.net", "auth_type": "basic"},
			wantErr:  "missing scheme",
		},
		{
			name:     "invalid url no host",
			settings: map[string]string{"url": "https://", "auth_type": "basic"},
			wantErr:  "missing host",
		},
		{
			name:     "invalid url scheme",
			settings: map[string]string{"url": "ftp://jira.com", "auth_type": "basic"},
			wantErr:  "scheme must be http or https",
		},
		{
			name:     "missing auth_type",
			settings: map[string]string{"url": "https://company.atlassian.net"},
			wantErr:  "auth_type is required",
		},
		{
			name:     "invalid auth_type",
			settings: map[string]string{"url": "https://company.atlassian.net", "auth_type": "oauth"},
			wantErr:  "invalid auth_type",
		},
		{
			name:     "invalid max_results",
			settings: map[string]string{"url": "https://company.atlassian.net", "auth_type": "basic", "max_results": "abc"},
			wantErr:  "invalid max_results",
		},
		{
			name:     "invalid poll_interval",
			settings: map[string]string{"url": "https://company.atlassian.net", "auth_type": "basic", "poll_interval": "xyz"},
			wantErr:  "invalid poll_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseConfig(tt.settings)
			if err == nil {
				t.Fatalf("ParseConfig() expected error containing %q, got nil", tt.wantErr)
				return
			}
			if got := err.Error(); !contains(got, tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", got, tt.wantErr)
			}
		})
	}
}

func TestParseConfig_EmptySettings(t *testing.T) {
	t.Parallel()

	_, err := ParseConfig(map[string]string{})
	if err == nil {
		t.Fatal("ParseConfig(empty) should fail")
	}
}

func TestValidateURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://company.atlassian.net", false},
		{"valid http", "http://jira.local:8080", false},
		{"valid with path", "https://jira.corp.com/jira", false},
		{"no scheme", "company.atlassian.net", true},
		{"no host", "https://", true},
		{"ftp scheme", "ftp://jira.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

// contains checks if s contains substr (avoids importing strings in tests).
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
