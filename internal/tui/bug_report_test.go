package tui

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCollectEnvironment_AllFieldsPopulated(t *testing.T) {
	t.Parallel()

	env := CollectEnvironment(120, 40, "Doors", "modern", 15, 3, 5*time.Minute)

	if env.Version == "" {
		t.Error("Version should not be empty")
	}
	if env.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if env.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", env.OS, runtime.GOOS)
	}
	if env.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q, want %q", env.Arch, runtime.GOARCH)
	}
	if env.TerminalWidth != 120 {
		t.Errorf("TerminalWidth = %d, want 120", env.TerminalWidth)
	}
	if env.TerminalHeight != 40 {
		t.Errorf("TerminalHeight = %d, want 40", env.TerminalHeight)
	}
	if env.CurrentView != "Doors" {
		t.Errorf("CurrentView = %q, want %q", env.CurrentView, "Doors")
	}
	if env.ThemeName != "modern" {
		t.Errorf("ThemeName = %q, want %q", env.ThemeName, "modern")
	}
	if env.TaskCount != 15 {
		t.Errorf("TaskCount = %d, want 15", env.TaskCount)
	}
	if env.ProviderCount != 3 {
		t.Errorf("ProviderCount = %d, want 3", env.ProviderCount)
	}
	if env.SessionDuration != 5*time.Minute {
		t.Errorf("SessionDuration = %v, want 5m0s", env.SessionDuration)
	}
}

func TestCollectEnvironment_SessionDurationTruncated(t *testing.T) {
	t.Parallel()

	env := CollectEnvironment(80, 24, "Detail", "classic", 0, 0, 5*time.Minute+123*time.Millisecond)

	if env.SessionDuration != 5*time.Minute {
		t.Errorf("SessionDuration = %v, want truncated to 5m0s", env.SessionDuration)
	}
}

func TestBugReport_FormatMarkdown_AllSections(t *testing.T) {
	t.Parallel()

	report := &BugReport{
		Description: "The app crashes when I press Enter",
		Environment: EnvironmentInfo{
			Version:         "1.0.0",
			Commit:          "abc1234",
			BuildDate:       "2025-01-15",
			GoVersion:       "go1.25.4",
			OS:              "darwin",
			Arch:            "arm64",
			TerminalWidth:   120,
			TerminalHeight:  40,
			CurrentView:     "Doors",
			ThemeName:       "modern",
			TaskCount:       5,
			ProviderCount:   2,
			SessionDuration: 10 * time.Minute,
		},
		Breadcrumbs: "2025-01-15T10:00:00Z [Doors] view:Doors\n2025-01-15T10:00:05Z [Detail] view:Detail",
		Timestamp:   time.Date(2025, 1, 15, 10, 0, 30, 0, time.UTC),
	}

	md := report.FormatMarkdown()

	checks := []struct {
		name   string
		substr string
	}{
		{"header", "## Bug Report"},
		{"timestamp", "2025-01-15T10:00:30Z"},
		{"description header", "### Description"},
		{"description text", "The app crashes when I press Enter"},
		{"environment header", "### Environment"},
		{"version", "1.0.0"},
		{"commit", "abc1234"},
		{"build date", "2025-01-15"},
		{"go version", "go1.25.4"},
		{"os/arch", "darwin/arm64"},
		{"terminal", "120x40"},
		{"view", "Doors"},
		{"theme", "modern"},
		{"task count", "| 5 |"},
		{"provider count", "| 2 |"},
		{"session duration", "10m0s"},
		{"breadcrumb header", "### Navigation Trail"},
		{"breadcrumb content", "view:Doors"},
	}

	for _, tc := range checks {
		if !strings.Contains(md, tc.substr) {
			t.Errorf("FormatMarkdown missing %s: expected substring %q", tc.name, tc.substr)
		}
	}
}

func TestBugReport_FormatMarkdown_NoBreadcrumbs(t *testing.T) {
	t.Parallel()

	report := &BugReport{
		Description: "Something broke",
		Environment: EnvironmentInfo{
			Version: "dev",
			OS:      "linux",
			Arch:    "amd64",
		},
		Timestamp: time.Now().UTC(),
	}

	md := report.FormatMarkdown()

	if strings.Contains(md, "### Navigation Trail") {
		t.Error("FormatMarkdown should omit Navigation Trail when breadcrumbs are empty")
	}
}

func TestBugReport_FormatMarkdown_IsValidGFM(t *testing.T) {
	t.Parallel()

	report := &BugReport{
		Description: "Test",
		Environment: EnvironmentInfo{
			Version: "dev",
			OS:      "darwin",
			Arch:    "arm64",
		},
		Timestamp: time.Now().UTC(),
	}

	md := report.FormatMarkdown()

	// Verify GFM table structure
	if !strings.Contains(md, "| Field | Value |") {
		t.Error("missing table header")
	}
	if !strings.Contains(md, "|-------|-------|") {
		t.Error("missing table separator")
	}
}

func TestScrubHomePath(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"replaces home dir", home + "/Documents/file.txt", "~/Documents/file.txt"},
		{"no home dir present", "/tmp/file.txt", "/tmp/file.txt"},
		{"empty string", "", ""},
		{"multiple occurrences", home + "/a " + home + "/b", "~/a ~/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ScrubHomePath(tt.input)
			if got != tt.want {
				t.Errorf("ScrubHomePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBugReport_PrivacyBlocklist(t *testing.T) {
	t.Parallel()

	// Verify that the environment info struct and FormatMarkdown do not include
	// any blocklisted data categories.
	report := &BugReport{
		Description: "test",
		Environment: CollectEnvironment(80, 24, "Doors", "modern", 5, 1, time.Minute),
		Timestamp:   time.Now().UTC(),
	}

	md := report.FormatMarkdown()

	// The output should never contain the user's home directory
	home, _ := os.UserHomeDir()
	if home != "" && strings.Contains(md, home) {
		t.Errorf("FormatMarkdown contains home directory %q — privacy violation", home)
	}
}

func TestEnvironmentInfo_NoTaskContent(t *testing.T) {
	t.Parallel()

	// EnvironmentInfo should only hold task COUNT, not any task text/names.
	env := CollectEnvironment(80, 24, "Doors", "modern", 42, 3, time.Minute)

	if env.TaskCount != 42 {
		t.Errorf("TaskCount = %d, want 42", env.TaskCount)
	}
	// Verify there's no field for task text — this is a compile-time check
	// enforced by the struct definition, but we validate the format too.
	report := &BugReport{
		Description: "test",
		Environment: env,
		Timestamp:   time.Now().UTC(),
	}
	md := report.FormatMarkdown()
	if !strings.Contains(md, "| Task Count | 42 |") {
		t.Error("expected task count in markdown output")
	}
}
