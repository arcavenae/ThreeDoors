package linear

import (
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	// Cannot use t.Parallel() because subtests use t.Setenv

	tests := []struct {
		name     string
		settings map[string]string
		envKey   string
		envVal   string
		wantErr  bool
		check    func(t *testing.T, cfg *LinearConfig)
	}{
		{
			name: "valid minimal config",
			settings: map[string]string{
				"api_key":  "lin_api_test123",
				"team_ids": "TEAM-1",
			},
			check: func(t *testing.T, cfg *LinearConfig) {
				t.Helper()
				if cfg.APIKey != "lin_api_test123" {
					t.Errorf("APIKey = %q, want %q", cfg.APIKey, "lin_api_test123")
				}
				if len(cfg.TeamIDs) != 1 || cfg.TeamIDs[0] != "TEAM-1" {
					t.Errorf("TeamIDs = %v, want [TEAM-1]", cfg.TeamIDs)
				}
				if cfg.PollInterval != DefaultPollInterval {
					t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, DefaultPollInterval)
				}
				if cfg.Assignee != "" {
					t.Errorf("Assignee = %q, want empty", cfg.Assignee)
				}
			},
		},
		{
			name: "multiple team IDs",
			settings: map[string]string{
				"api_key":  "lin_api_test123",
				"team_ids": "TEAM-1, TEAM-2, TEAM-3",
			},
			check: func(t *testing.T, cfg *LinearConfig) {
				t.Helper()
				if len(cfg.TeamIDs) != 3 {
					t.Errorf("TeamIDs len = %d, want 3", len(cfg.TeamIDs))
				}
				if cfg.TeamIDs[1] != "TEAM-2" {
					t.Errorf("TeamIDs[1] = %q, want %q", cfg.TeamIDs[1], "TEAM-2")
				}
			},
		},
		{
			name: "custom poll interval and assignee",
			settings: map[string]string{
				"api_key":       "lin_api_test123",
				"team_ids":      "TEAM-1",
				"poll_interval": "10m",
				"assignee":      "@me",
			},
			check: func(t *testing.T, cfg *LinearConfig) {
				t.Helper()
				if cfg.PollInterval != 10*time.Minute {
					t.Errorf("PollInterval = %v, want 10m", cfg.PollInterval)
				}
				if cfg.Assignee != "@me" {
					t.Errorf("Assignee = %q, want @me", cfg.Assignee)
				}
			},
		},
		{
			name: "env var overrides config api_key",
			settings: map[string]string{
				"api_key":  "config_key",
				"team_ids": "TEAM-1",
			},
			envKey: "LINEAR_API_KEY",
			envVal: "env_key_override",
			check: func(t *testing.T, cfg *LinearConfig) {
				t.Helper()
				if cfg.APIKey != "env_key_override" {
					t.Errorf("APIKey = %q, want env_key_override", cfg.APIKey)
				}
			},
		},
		{
			name: "env var provides api_key when config missing",
			settings: map[string]string{
				"team_ids": "TEAM-1",
			},
			envKey: "LINEAR_API_KEY",
			envVal: "env_only_key",
			check: func(t *testing.T, cfg *LinearConfig) {
				t.Helper()
				if cfg.APIKey != "env_only_key" {
					t.Errorf("APIKey = %q, want env_only_key", cfg.APIKey)
				}
			},
		},
		{
			name:     "missing api_key",
			settings: map[string]string{"team_ids": "TEAM-1"},
			wantErr:  true,
		},
		{
			name:     "missing team_ids",
			settings: map[string]string{"api_key": "lin_api_test123"},
			wantErr:  true,
		},
		{
			name: "invalid poll_interval",
			settings: map[string]string{
				"api_key":       "lin_api_test123",
				"team_ids":      "TEAM-1",
				"poll_interval": "invalid",
			},
			wantErr: true,
		},
		{
			name:     "nil settings with no env var",
			settings: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envVal)
			} else {
				// Clear env to avoid interference
				t.Setenv("LINEAR_API_KEY", "")
			}

			cfg, err := ParseConfig(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil && cfg != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestParseConfigEnvVarCleared(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "")

	_, err := ParseConfig(map[string]string{
		"team_ids": "TEAM-1",
	})
	if err == nil {
		t.Fatal("expected error when api_key missing and no env var")
	}
}
