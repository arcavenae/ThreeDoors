package todoist

import (
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings map[string]string
		wantErr  bool
		check    func(t *testing.T, cfg *TodoistConfig)
	}{
		{
			name: "valid with api_token",
			settings: map[string]string{
				"api_token": "test-token-123",
			},
			check: func(t *testing.T, cfg *TodoistConfig) {
				t.Helper()
				if cfg.APIToken != "test-token-123" {
					t.Errorf("expected api_token test-token-123, got %s", cfg.APIToken)
				}
				if cfg.PollInterval != DefaultPollInterval {
					t.Errorf("expected default poll interval, got %s", cfg.PollInterval)
				}
			},
		},
		{
			name: "valid with project_ids",
			settings: map[string]string{
				"api_token":   "tok",
				"project_ids": "111, 222, 333",
			},
			check: func(t *testing.T, cfg *TodoistConfig) {
				t.Helper()
				if len(cfg.ProjectIDs) != 3 {
					t.Fatalf("expected 3 project IDs, got %d", len(cfg.ProjectIDs))
				}
				if cfg.ProjectIDs[0] != "111" || cfg.ProjectIDs[1] != "222" || cfg.ProjectIDs[2] != "333" {
					t.Errorf("unexpected project IDs: %v", cfg.ProjectIDs)
				}
			},
		},
		{
			name: "valid with filter",
			settings: map[string]string{
				"api_token": "tok",
				"filter":    "today | overdue",
			},
			check: func(t *testing.T, cfg *TodoistConfig) {
				t.Helper()
				if cfg.Filter != "today | overdue" {
					t.Errorf("expected filter 'today | overdue', got %s", cfg.Filter)
				}
			},
		},
		{
			name: "valid with custom poll_interval",
			settings: map[string]string{
				"api_token":     "tok",
				"poll_interval": "2m",
			},
			check: func(t *testing.T, cfg *TodoistConfig) {
				t.Helper()
				if cfg.PollInterval != 2*time.Minute {
					t.Errorf("expected 2m poll interval, got %s", cfg.PollInterval)
				}
			},
		},
		{
			name:     "missing api_token",
			settings: map[string]string{},
			wantErr:  true,
		},
		{
			name:     "nil settings no env",
			settings: nil,
			wantErr:  true,
		},
		{
			name: "mutually exclusive project_ids and filter",
			settings: map[string]string{
				"api_token":   "tok",
				"project_ids": "111",
				"filter":      "today",
			},
			wantErr: true,
		},
		{
			name: "invalid poll_interval",
			settings: map[string]string{
				"api_token":     "tok",
				"poll_interval": "not-a-duration",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

func TestParseConfigEnvVarOverridesConfig(t *testing.T) {
	t.Setenv("TODOIST_API_TOKEN", "env-token")

	cfg, err := ParseConfig(map[string]string{
		"api_token": "config-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if cfg.APIToken != "env-token" {
		t.Errorf("expected env-token, got %s", cfg.APIToken)
	}
}

func TestParseConfigEnvVarProvidesToken(t *testing.T) {
	t.Setenv("TODOIST_API_TOKEN", "env-only-token")

	cfg, err := ParseConfig(map[string]string{
		"project_ids": "123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if cfg.APIToken != "env-only-token" {
		t.Errorf("expected env-only-token, got %s", cfg.APIToken)
	}
}

func TestParseConfigEmptyProjectIDs(t *testing.T) {
	t.Parallel()

	cfg, err := ParseConfig(map[string]string{
		"api_token":   "tok",
		"project_ids": ", , ,",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if len(cfg.ProjectIDs) != 0 {
		t.Errorf("expected 0 project IDs from whitespace-only list, got %d: %v", len(cfg.ProjectIDs), cfg.ProjectIDs)
	}
}

func TestAuthConfigFrom(t *testing.T) {
	t.Parallel()

	cfg := &TodoistConfig{APIToken: "my-token"}
	auth := AuthConfigFrom(cfg)
	if auth.APIToken != "my-token" {
		t.Errorf("expected my-token, got %s", auth.APIToken)
	}
}
