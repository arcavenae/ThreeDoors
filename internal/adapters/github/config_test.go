package github

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
		errMsg   string
		check    func(t *testing.T, cfg *GitHubConfig)
	}{
		{
			name: "minimal valid config",
			settings: map[string]string{
				"repos": "owner/repo",
			},
			check: func(t *testing.T, cfg *GitHubConfig) {
				t.Helper()
				if len(cfg.Repos) != 1 || cfg.Repos[0] != "owner/repo" {
					t.Errorf("repos = %v, want [owner/repo]", cfg.Repos)
				}
				if cfg.Assignee != DefaultAssignee {
					t.Errorf("assignee = %q, want %q", cfg.Assignee, DefaultAssignee)
				}
				if cfg.PollInterval != DefaultPollInterval {
					t.Errorf("poll_interval = %v, want %v", cfg.PollInterval, DefaultPollInterval)
				}
				if cfg.InProgressLabel != DefaultInProgress {
					t.Errorf("in_progress_label = %q, want %q", cfg.InProgressLabel, DefaultInProgress)
				}
			},
		},
		{
			name: "multiple repos",
			settings: map[string]string{
				"repos": "owner/repo1, owner/repo2, org/other",
			},
			check: func(t *testing.T, cfg *GitHubConfig) {
				t.Helper()
				if len(cfg.Repos) != 3 {
					t.Errorf("repos count = %d, want 3", len(cfg.Repos))
				}
				if cfg.Repos[2] != "org/other" {
					t.Errorf("repos[2] = %q, want %q", cfg.Repos[2], "org/other")
				}
			},
		},
		{
			name: "custom settings",
			settings: map[string]string{
				"repos":             "owner/repo",
				"token":             "config-token",
				"assignee":          "testuser",
				"poll_interval":     "10m",
				"in_progress_label": "wip",
			},
			check: func(t *testing.T, cfg *GitHubConfig) {
				t.Helper()
				if cfg.Token != "config-token" {
					t.Errorf("token = %q, want %q", cfg.Token, "config-token")
				}
				if cfg.Assignee != "testuser" {
					t.Errorf("assignee = %q, want %q", cfg.Assignee, "testuser")
				}
				if cfg.PollInterval != 10*time.Minute {
					t.Errorf("poll_interval = %v, want 10m", cfg.PollInterval)
				}
				if cfg.InProgressLabel != "wip" {
					t.Errorf("in_progress_label = %q, want %q", cfg.InProgressLabel, "wip")
				}
			},
		},
		{
			name: "priority labels",
			settings: map[string]string{
				"repos":                    "owner/repo",
				"priority_label.critical":  "high",
				"priority_label.low":       "low",
				"priority_label.":          "ignored",
				"priority_label.something": "",
			},
			check: func(t *testing.T, cfg *GitHubConfig) {
				t.Helper()
				if len(cfg.PriorityLabels) != 2 {
					t.Errorf("priority_labels count = %d, want 2", len(cfg.PriorityLabels))
				}
				if cfg.PriorityLabels["critical"] != "high" {
					t.Errorf("priority_labels[critical] = %q, want %q", cfg.PriorityLabels["critical"], "high")
				}
			},
		},
		{
			name:     "no repos",
			settings: map[string]string{},
			wantErr:  true,
			errMsg:   "repos is required",
		},
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  true,
			errMsg:   "repos is required",
		},
		{
			name: "invalid repo format - no slash",
			settings: map[string]string{
				"repos": "justarepo",
			},
			wantErr: true,
			errMsg:  "must be in owner/repo format",
		},
		{
			name: "invalid repo format - empty owner",
			settings: map[string]string{
				"repos": "/repo",
			},
			wantErr: true,
			errMsg:  "must be in owner/repo format",
		},
		{
			name: "invalid repo format - empty repo",
			settings: map[string]string{
				"repos": "owner/",
			},
			wantErr: true,
			errMsg:  "must be in owner/repo format",
		},
		{
			name: "invalid poll_interval",
			settings: map[string]string{
				"repos":         "owner/repo",
				"poll_interval": "not-a-duration",
			},
			wantErr: true,
			errMsg:  "invalid poll_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := ParseConfig(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestParseConfigEnvTokenOverride(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env-token")

	cfg, err := ParseConfig(map[string]string{
		"repos": "owner/repo",
		"token": "config-token",
	})
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.Token != "env-token" {
		t.Errorf("token = %q, want %q (env override)", cfg.Token, "env-token")
	}
}

func TestValidateRepoFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repo    string
		wantErr bool
	}{
		{"valid", "owner/repo", false},
		{"no slash", "noslash", true},
		{"too many slashes", "a/b/c", true},
		{"empty owner", "/repo", true},
		{"empty repo", "owner/", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateRepoFormat(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRepoFormat(%q) error = %v, wantErr %v", tt.repo, err, tt.wantErr)
			}
		})
	}
}

func TestSplitRepos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"single", "owner/repo", 1},
		{"multiple", "a/b, c/d, e/f", 3},
		{"with extra spaces", " a/b , c/d ", 2},
		{"trailing comma", "a/b,", 1},
		{"empty parts", "a/b,,c/d", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitRepos(tt.input)
			if len(got) != tt.want {
				t.Errorf("splitRepos(%q) = %v (len %d), want len %d", tt.input, got, len(got), tt.want)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
