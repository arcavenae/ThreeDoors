package clickup

import (
	"testing"
	"time"
)

func TestParseConfig_Valid(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"api_token": "pk_test_token",
		"team_id":   "team123",
		"space_ids": "space1,space2",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.APIToken != "pk_test_token" {
		t.Errorf("APIToken = %q, want %q", cfg.APIToken, "pk_test_token")
	}
	if cfg.TeamID != "team123" {
		t.Errorf("TeamID = %q, want %q", cfg.TeamID, "team123")
	}
	if len(cfg.SpaceIDs) != 2 || cfg.SpaceIDs[0] != "space1" || cfg.SpaceIDs[1] != "space2" {
		t.Errorf("SpaceIDs = %v, want [space1 space2]", cfg.SpaceIDs)
	}
}

func TestParseConfig_ListIDsOnly(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"api_token": "pk_test_token",
		"team_id":   "team123",
		"list_ids":  "list1, list2, list3",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if len(cfg.SpaceIDs) != 0 {
		t.Errorf("SpaceIDs = %v, want empty", cfg.SpaceIDs)
	}
	if len(cfg.ListIDs) != 3 {
		t.Fatalf("ListIDs length = %d, want 3", len(cfg.ListIDs))
	}
	if cfg.ListIDs[1] != "list2" {
		t.Errorf("ListIDs[1] = %q, want %q (should be trimmed)", cfg.ListIDs[1], "list2")
	}
}

func TestParseConfig_Defaults(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"api_token": "pk_test_token",
		"team_id":   "team123",
		"list_ids":  "list1",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.PollInterval != DefaultPollInterval {
		t.Errorf("PollInterval = %v, want default %v", cfg.PollInterval, DefaultPollInterval)
	}
	if cfg.Assignee != "" {
		t.Errorf("Assignee = %q, want empty", cfg.Assignee)
	}
}

func TestParseConfig_CustomOptionals(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"api_token":     "pk_test_token",
		"team_id":       "team123",
		"space_ids":     "space1",
		"assignee":      "user123",
		"poll_interval": "1m",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.Assignee != "user123" {
		t.Errorf("Assignee = %q, want %q", cfg.Assignee, "user123")
	}
	if cfg.PollInterval != time.Minute {
		t.Errorf("PollInterval = %v, want 1m", cfg.PollInterval)
	}
}

func TestParseConfig_EnvVarOverride(t *testing.T) {
	t.Setenv("CLICKUP_API_TOKEN", "env-token")

	settings := map[string]string{
		"api_token": "config-token",
		"team_id":   "team123",
		"space_ids": "space1",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.APIToken != "env-token" {
		t.Errorf("APIToken = %q, want env override %q", cfg.APIToken, "env-token")
	}
}

func TestParseConfig_EnvVarOnly(t *testing.T) {
	t.Setenv("CLICKUP_API_TOKEN", "env-only-token")

	settings := map[string]string{
		"team_id":  "team123",
		"list_ids": "list1",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if cfg.APIToken != "env-only-token" {
		t.Errorf("APIToken = %q, want %q", cfg.APIToken, "env-only-token")
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
			name:     "missing api_token",
			settings: map[string]string{"team_id": "t", "space_ids": "s"},
			wantErr:  "api_token is required",
		},
		{
			name:     "missing team_id",
			settings: map[string]string{"api_token": "t", "space_ids": "s"},
			wantErr:  "team_id is required",
		},
		{
			name:     "missing space_ids and list_ids",
			settings: map[string]string{"api_token": "t", "team_id": "team1"},
			wantErr:  "at least one of space_ids or list_ids",
		},
		{
			name:     "invalid poll_interval",
			settings: map[string]string{"api_token": "t", "team_id": "team1", "space_ids": "s", "poll_interval": "xyz"},
			wantErr:  "invalid poll_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseConfig(tt.settings)
			if err == nil {
				t.Fatalf("ParseConfig() expected error containing %q, got nil", tt.wantErr)
			}
			if !containsStr(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseConfig_NilSettings(t *testing.T) {
	t.Setenv("CLICKUP_API_TOKEN", "env-token")

	_, err := ParseConfig(nil)
	if err == nil {
		t.Fatal("ParseConfig(nil) should fail when team_id is missing")
	}
}

func TestParseConfig_EmptySettings(t *testing.T) {
	t.Parallel()

	_, err := ParseConfig(map[string]string{})
	if err == nil {
		t.Fatal("ParseConfig(empty) should fail")
	}
}

func TestParseConfig_BothSpaceAndListIDs(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"api_token": "pk_test_token",
		"team_id":   "team123",
		"space_ids": "space1",
		"list_ids":  "list1",
	}

	cfg, err := ParseConfig(settings)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if len(cfg.SpaceIDs) != 1 {
		t.Errorf("SpaceIDs length = %d, want 1", len(cfg.SpaceIDs))
	}
	if len(cfg.ListIDs) != 1 {
		t.Errorf("ListIDs length = %d, want 1", len(cfg.ListIDs))
	}
}

func TestSplitAndTrim(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple", "a,b,c", []string{"a", "b", "c"}},
		{"with spaces", " a , b , c ", []string{"a", "b", "c"}},
		{"single", "only", []string{"only"}},
		{"empty elements", "a,,b", []string{"a", "b"}},
		{"trailing comma", "a,b,", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitAndTrim(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitAndTrim(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("splitAndTrim(%q)[%d] = %q, want %q", tt.input, i, v, tt.want[i])
				}
			}
		})
	}
}
