package tui

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/adapters/github"
	"github.com/arcavenae/ThreeDoors/internal/cli"
	"github.com/arcavenae/ThreeDoors/internal/core/connection/oauth"
	tea "github.com/charmbracelet/bubbletea"
)

// execCommand is a variable for testing clipboard commands.
var execCommand = exec.Command

// EnvironmentInfo holds allowlisted environment data for bug reports.
// Only safe, non-personal data is collected — no task content, file paths,
// credentials, or personal information.
type EnvironmentInfo struct {
	Version         string
	Commit          string
	BuildDate       string
	GoVersion       string
	OS              string
	Arch            string
	TerminalWidth   int
	TerminalHeight  int
	CurrentView     string
	ThemeName       string
	TaskCount       int
	ProviderCount   int
	SessionDuration time.Duration
}

// BugReport holds a complete bug report ready for formatting.
type BugReport struct {
	Description string
	Environment EnvironmentInfo
	Breadcrumbs string
	Timestamp   time.Time
}

// CollectEnvironment gathers allowlisted environment data.
// This function NEVER collects: task names/content, file paths, provider names/config,
// search queries, text input, tag names, username/home directory, credentials,
// mood entries, values/goals.
func CollectEnvironment(width, height int, currentView, themeName string, taskCount, providerCount int, sessionDuration time.Duration) EnvironmentInfo {
	return EnvironmentInfo{
		Version:         cli.Version,
		Commit:          cli.Commit,
		BuildDate:       cli.BuildDate,
		GoVersion:       runtime.Version(),
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		TerminalWidth:   width,
		TerminalHeight:  height,
		CurrentView:     currentView,
		ThemeName:       themeName,
		TaskCount:       taskCount,
		ProviderCount:   providerCount,
		SessionDuration: sessionDuration.Truncate(time.Second),
	}
}

// FormatMarkdown renders the bug report as GitHub-flavored markdown.
func (r *BugReport) FormatMarkdown() string {
	var s strings.Builder

	fmt.Fprintf(&s, "## Bug Report\n\n")
	fmt.Fprintf(&s, "**Submitted:** %s\n\n", r.Timestamp.Format(time.RFC3339))

	fmt.Fprintf(&s, "### Description\n\n")
	fmt.Fprintf(&s, "%s\n\n", r.Description)

	fmt.Fprintf(&s, "### Environment\n\n")
	fmt.Fprintf(&s, "| Field | Value |\n")
	fmt.Fprintf(&s, "|-------|-------|\n")
	fmt.Fprintf(&s, "| Version | %s |\n", r.Environment.Version)
	fmt.Fprintf(&s, "| Commit | %s |\n", r.Environment.Commit)
	fmt.Fprintf(&s, "| Build Date | %s |\n", r.Environment.BuildDate)
	fmt.Fprintf(&s, "| Go Version | %s |\n", r.Environment.GoVersion)
	fmt.Fprintf(&s, "| OS/Arch | %s/%s |\n", r.Environment.OS, r.Environment.Arch)
	fmt.Fprintf(&s, "| Terminal | %dx%d |\n", r.Environment.TerminalWidth, r.Environment.TerminalHeight)
	fmt.Fprintf(&s, "| Current View | %s |\n", r.Environment.CurrentView)
	fmt.Fprintf(&s, "| Theme | %s |\n", r.Environment.ThemeName)
	fmt.Fprintf(&s, "| Task Count | %d |\n", r.Environment.TaskCount)
	fmt.Fprintf(&s, "| Provider Count | %d |\n", r.Environment.ProviderCount)
	fmt.Fprintf(&s, "| Session Duration | %s |\n", r.Environment.SessionDuration)

	if r.Breadcrumbs != "" {
		fmt.Fprintf(&s, "\n### Navigation Trail\n\n")
		fmt.Fprintf(&s, "```\n%s\n```\n", r.Breadcrumbs)
	}

	return s.String()
}

// ScrubHomePath replaces the user's home directory with ~ in a string.
func ScrubHomePath(s string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return s
	}
	return strings.ReplaceAll(s, home, "~")
}

// bugReportTarget is the hardcoded repository for bug reports (D-116).
const bugReportTarget = "arcavenae/ThreeDoors"

// maxIssueURLLen is the conservative browser URL length limit.
const maxIssueURLLen = 7500

// BugReportSubmittedMsg is sent when a bug report is successfully submitted.
type BugReportSubmittedMsg struct {
	Method  string // "browser", "api", "file"
	Details string // URL or file path
}

// BugReportErrorMsg is sent when a submission method fails.
type BugReportErrorMsg struct {
	Method string
	Err    error
}

// BuildIssueURL constructs a GitHub new-issue URL with pre-filled title and body.
// If the URL exceeds maxIssueURLLen, breadcrumbs are truncated from the report body.
func BuildIssueURL(report *BugReport) string {
	title := fmt.Sprintf("[Bug] %s", truncateString(report.Description, 80))

	body := report.FormatMarkdown()

	base := fmt.Sprintf("https://github.com/%s/issues/new", bugReportTarget)
	params := url.Values{}
	params.Set("title", title)
	params.Set("labels", "type.bug")
	params.Set("body", body)

	fullURL := base + "?" + params.Encode()

	if len(fullURL) > maxIssueURLLen && report.Breadcrumbs != "" {
		// Re-render without breadcrumbs to fit within limit.
		trimmed := *report
		trimmed.Breadcrumbs = ""
		body = trimmed.FormatMarkdown()
		params.Set("body", body)
		fullURL = base + "?" + params.Encode()
	}

	return fullURL
}

// truncateString truncates s to maxLen characters, appending "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// openBrowserCmd returns a tea.Cmd that opens the bug report URL in the default browser.
func openBrowserCmd(report *BugReport) tea.Cmd {
	return func() tea.Msg {
		issueURL := BuildIssueURL(report)
		if err := oauth.OpenBrowser(context.Background(), issueURL); err != nil {
			return BugReportErrorMsg{Method: "browser", Err: err}
		}
		return BugReportSubmittedMsg{Method: "browser", Details: issueURL}
	}
}

// submitViaAPICmd returns a tea.Cmd that creates a GitHub issue via the API.
func submitViaAPICmd(report *BugReport, client *github.GitHubClient) tea.Cmd {
	return func() tea.Msg {
		title := fmt.Sprintf("[Bug] %s", truncateString(report.Description, 80))
		body := report.FormatMarkdown()

		parts := strings.SplitN(bugReportTarget, "/", 2)
		issue, err := client.CreateIssue(context.Background(), parts[0], parts[1], title, body)
		if err != nil {
			return BugReportErrorMsg{Method: "api", Err: err}
		}
		return BugReportSubmittedMsg{Method: "api", Details: issue.HTMLURL}
	}
}

// saveBugReportCmd returns a tea.Cmd that saves the report to a local file.
func saveBugReportCmd(report *BugReport) tea.Cmd {
	return func() tea.Msg {
		home, err := os.UserHomeDir()
		if err != nil {
			return BugReportErrorMsg{Method: "file", Err: fmt.Errorf("get home directory: %w", err)}
		}

		dir := filepath.Join(home, ".threedoors", "bug-reports")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return BugReportErrorMsg{Method: "file", Err: fmt.Errorf("create bug reports directory: %w", err)}
		}

		// RFC3339 timestamp with colons replaced by hyphens for filesystem safety.
		ts := report.Timestamp.Format(time.RFC3339)
		ts = strings.ReplaceAll(ts, ":", "-")
		filename := fmt.Sprintf("bug-%s.md", ts)
		path := filepath.Join(dir, filename)

		content := []byte(report.FormatMarkdown())

		if err := atomicWriteBugReport(path, content); err != nil {
			return BugReportErrorMsg{Method: "file", Err: err}
		}

		return BugReportSubmittedMsg{Method: "file", Details: ScrubHomePath(path)}
	}
}

// atomicWriteBugReport writes data to path using write-tmp/fsync/rename.
func atomicWriteBugReport(path string, data []byte) error {
	tmpPath := path + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create temp file %s: %w", tmpPath, err)
	}

	writeOK := false
	defer func() {
		if !writeOK {
			_ = f.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write temp file %s: %w", tmpPath, err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync temp file %s: %w", tmpPath, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file %s: %w", tmpPath, err)
	}

	writeOK = true

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tmpPath, path, err)
	}

	return nil
}

// copyToClipboardCmd returns a tea.Cmd that copies text to the system clipboard.
func copyToClipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		var cmd string
		var args []string

		switch runtime.GOOS {
		case "darwin":
			cmd = "pbcopy"
		case "linux":
			cmd = "xclip"
			args = []string{"-selection", "clipboard"}
		default:
			return BugReportErrorMsg{
				Method: "clipboard",
				Err:    fmt.Errorf("clipboard not supported on %s", runtime.GOOS),
			}
		}

		c := execCommand(cmd, args...)
		c.Stdin = strings.NewReader(text)
		if err := c.Run(); err != nil {
			return BugReportErrorMsg{Method: "clipboard", Err: fmt.Errorf("copy to clipboard: %w", err)}
		}

		return BugReportSubmittedMsg{Method: "clipboard", Details: "URL copied to clipboard"}
	}
}

// hasGitHubToken checks if a GitHub token is available for API submission.
func hasGitHubToken() bool {
	return os.Getenv("GITHUB_TOKEN") != ""
}

// newGitHubClientForBugReport creates a GitHubClient using the GITHUB_TOKEN env var.
func newGitHubClientForBugReport() *github.GitHubClient {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil
	}
	parts := strings.SplitN(bugReportTarget, "/", 2)
	cfg := &github.GitHubConfig{
		Token: token,
		Repos: []string{parts[0] + "/" + parts[1]},
	}
	return github.NewGitHubClient(cfg)
}
