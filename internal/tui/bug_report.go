package tui

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/cli"
)

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
