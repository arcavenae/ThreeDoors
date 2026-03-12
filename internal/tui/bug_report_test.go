package tui

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

func newTestReport() *BugReport {
	return &BugReport{
		Description: "The app crashes when I press Enter",
		Environment: EnvironmentInfo{
			Version:   "1.0.0",
			Commit:    "abc1234",
			OS:        "darwin",
			Arch:      "arm64",
			TaskCount: 5,
		},
		Breadcrumbs: "trail",
		Timestamp:   time.Date(2025, 1, 15, 10, 0, 30, 0, time.UTC),
	}
}

func TestBuildIssueURL_ContainsRequiredParts(t *testing.T) {
	t.Parallel()

	report := newTestReport()
	issueURL := BuildIssueURL(report)

	checks := []struct {
		name   string
		substr string
	}{
		{"github base", "https://github.com/arcaven/ThreeDoors/issues/new"},
		{"title param", "title="},
		{"body param", "body="},
		{"label param", "labels=type.bug"},
	}

	for _, tc := range checks {
		if !strings.Contains(issueURL, tc.substr) {
			t.Errorf("BuildIssueURL missing %s: expected substring %q in URL", tc.name, tc.substr)
		}
	}
}

func TestBuildIssueURL_EncodesSpecialCharacters(t *testing.T) {
	t.Parallel()

	report := newTestReport()
	report.Description = "crash with special chars: <>&\"'"
	issueURL := BuildIssueURL(report)

	// The URL should be properly encoded — no raw < > & in query string.
	parsed, err := url.Parse(issueURL)
	if err != nil {
		t.Fatalf("BuildIssueURL returned invalid URL: %v", err)
	}

	title := parsed.Query().Get("title")
	if !strings.Contains(title, "crash with special chars") {
		t.Errorf("title should contain description, got %q", title)
	}
}

func TestBuildIssueURL_TruncatesBreadcrumbsWhenTooLong(t *testing.T) {
	t.Parallel()

	report := newTestReport()
	// Create breadcrumbs long enough to exceed the URL limit.
	report.Breadcrumbs = strings.Repeat("2025-01-15T10:00:00Z [Doors] view:Doors\n", 300)

	issueURL := BuildIssueURL(report)

	if len(issueURL) > maxIssueURLLen {
		t.Errorf("URL length %d exceeds max %d after truncation", len(issueURL), maxIssueURLLen)
	}

	// Breadcrumbs should be removed from the body.
	if strings.Contains(issueURL, "Navigation+Trail") || strings.Contains(issueURL, "Navigation%20Trail") {
		t.Error("truncated URL should not contain Navigation Trail section")
	}
}

func TestBuildIssueURL_ShortReportNoTruncation(t *testing.T) {
	t.Parallel()

	report := newTestReport()
	report.Breadcrumbs = "short trail"

	issueURL := BuildIssueURL(report)

	// Short breadcrumbs should be preserved.
	if !strings.Contains(issueURL, url.QueryEscape("short trail")) && !strings.Contains(issueURL, "short+trail") {
		t.Error("short breadcrumbs should be preserved in URL")
	}
}

func TestBuildIssueURL_TitleTruncation(t *testing.T) {
	t.Parallel()

	report := newTestReport()
	report.Description = strings.Repeat("a", 200)

	issueURL := BuildIssueURL(report)
	parsed, err := url.Parse(issueURL)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}

	title := parsed.Query().Get("title")
	// Title should be truncated: "[Bug] " + 80 chars max
	if len(title) > 90 {
		t.Errorf("title too long: %d chars", len(title))
	}
	if !strings.HasSuffix(title, "...") {
		t.Error("truncated title should end with ...")
	}
}

func TestTruncateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world", 8, "hello..."},
		{"very short max", "hello", 2, "he"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestSaveBugReportCmd_CreatesFile(t *testing.T) {
	t.Parallel()

	report := newTestReport()

	// Use a temp directory to avoid writing to the real home dir.
	tmpDir := t.TempDir()

	// Directly test atomicWriteBugReport and file content.
	ts := report.Timestamp.Format(time.RFC3339)
	ts = strings.ReplaceAll(ts, ":", "-")
	filename := fmt.Sprintf("bug-%s.md", ts)
	path := filepath.Join(tmpDir, filename)

	content := []byte(report.FormatMarkdown())
	if err := atomicWriteBugReport(path, content); err != nil {
		t.Fatalf("atomicWriteBugReport: %v", err)
	}

	// Verify file exists and has correct content.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	if !strings.Contains(string(data), "## Bug Report") {
		t.Error("saved file should contain bug report markdown")
	}
	if !strings.Contains(string(data), report.Description) {
		t.Error("saved file should contain the description")
	}

	// Verify file permissions.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat saved file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestAtomicWriteBugReport_NoTempFileOnSuccess(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.md")

	if err := atomicWriteBugReport(path, []byte("test content")); err != nil {
		t.Fatalf("atomicWriteBugReport: %v", err)
	}

	// Temp file should not exist after successful write.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("temp file should be removed after successful write")
	}

	// Final file should exist.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("content = %q, want %q", string(data), "test content")
	}
}

func TestBugReportFilenameFormat(t *testing.T) {
	t.Parallel()

	ts := time.Date(2025, 3, 15, 14, 30, 0, 0, time.UTC)
	formatted := ts.Format(time.RFC3339)
	formatted = strings.ReplaceAll(formatted, ":", "-")

	want := "2025-03-15T14-30-00Z"
	if formatted != want {
		t.Errorf("timestamp = %q, want %q", formatted, want)
	}

	filename := fmt.Sprintf("bug-%s.md", formatted)
	if filename != "bug-2025-03-15T14-30-00Z.md" {
		t.Errorf("filename = %q", filename)
	}
}

func TestHasGitHubToken_Set(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token-123")

	if !hasGitHubToken() {
		t.Error("hasGitHubToken should return true when GITHUB_TOKEN is set")
	}
}

func TestHasGitHubToken_Empty(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	if hasGitHubToken() {
		t.Error("hasGitHubToken should return false when GITHUB_TOKEN is empty")
	}
}

func TestBugReportTarget(t *testing.T) {
	t.Parallel()

	if bugReportTarget != "arcaven/ThreeDoors" {
		t.Errorf("bugReportTarget = %q, want %q", bugReportTarget, "arcaven/ThreeDoors")
	}
}
