package tui

import (
	"fmt"
	"strings"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/charmbracelet/lipgloss"
)

// Known provider badge labels.
var providerLabels = map[string]string{
	"textfile":   "TXT",
	"obsidian":   "OBS",
	"applenotes": "NOTES",
	"todoist":    "TDT",
	"linear":     "LN",
}

// Provider badge colors.
var providerColors = map[string]lipgloss.Color{
	"textfile":   lipgloss.Color("243"), // gray
	"obsidian":   lipgloss.Color("141"), // purple
	"applenotes": lipgloss.Color("220"), // yellow
	"todoist":    lipgloss.Color("196"), // red (Todoist brand)
	"linear":     lipgloss.Color("99"),  // purple-blue (Linear brand)
}

// SourceBadgeLabel returns the short label for a provider name.
// Known providers get predefined labels; unknown providers get an
// uppercase abbreviation (up to 4 characters).
func SourceBadgeLabel(provider string) string {
	if provider == "" {
		return ""
	}
	if label, ok := providerLabels[provider]; ok {
		return label
	}
	upper := strings.ToUpper(provider)
	if len(upper) > 4 {
		return upper[:4]
	}
	return upper
}

// SourceBadge returns a styled badge string for display in TUI views.
func SourceBadge(provider string) string {
	label := SourceBadgeLabel(provider)
	if label == "" {
		return ""
	}

	color, ok := providerColors[provider]
	if !ok {
		color = lipgloss.Color("243") // gray fallback
	}

	style := lipgloss.NewStyle().Foreground(color)
	return style.Render("[" + label + "]")
}

// DevDispatchBadge returns a styled badge showing PR status for a dispatched task.
// Returns empty string if the task has no DevDispatch or no PR number.
func DevDispatchBadge(task *core.Task) string {
	if task.DevDispatch == nil {
		return ""
	}
	dd := task.DevDispatch
	if dd.PRNumber > 0 {
		if dd.PRStatus == "merged" {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
			return style.Render("[MERGED]")
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		return style.Render(fmt.Sprintf("[PR #%d]", dd.PRNumber))
	}
	if dd.Queued {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
		return style.Render("[QUEUED]")
	}
	return ""
}

// DuplicateIndicator returns a styled indicator for potential duplicates.
func DuplicateIndicator() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Faint(true)
	return style.Render("Possible duplicate")
}
