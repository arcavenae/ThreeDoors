package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/charmbracelet/lipgloss"
)

// RenderSyncStatusBar renders a compact sync status bar for all tracked providers.
// Returns an empty string if no providers are registered.
func RenderSyncStatusBar(tracker *tasks.SyncStatusTracker) string {
	if tracker == nil || tracker.Count() == 0 {
		return ""
	}

	statuses := tracker.All()
	// Sort by name for deterministic rendering
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	var parts []string
	for _, s := range statuses {
		parts = append(parts, renderProviderStatus(s))
	}

	bar := strings.Join(parts, syncStatusSeparator)
	return syncStatusBarStyle.Render(bar)
}

// renderProviderStatus renders a single provider's status with appropriate styling.
func renderProviderStatus(s tasks.ProviderSyncStatus) string {
	icon := s.Icon()
	var styledIcon string

	switch s.Phase {
	case tasks.SyncPhaseSynced:
		styledIcon = syncStatusSyncedStyle.Render(icon)
	case tasks.SyncPhaseSyncing:
		styledIcon = syncStatusSyncingStyle.Render(icon)
	case tasks.SyncPhasePending:
		styledIcon = syncStatusPendingStyle.Render(icon)
	case tasks.SyncPhaseError:
		styledIcon = syncStatusErrorStyle.Render(icon)
	default:
		styledIcon = icon
	}

	label := syncStatusLabelStyle.Render(s.Name)
	detail := renderDetail(s)

	if detail != "" {
		return fmt.Sprintf("%s %s %s", styledIcon, label, detail)
	}
	return fmt.Sprintf("%s %s", styledIcon, label)
}

// renderDetail renders extra information based on sync phase.
func renderDetail(s tasks.ProviderSyncStatus) string {
	switch s.Phase {
	case tasks.SyncPhasePending:
		return syncStatusDetailStyle.Render(fmt.Sprintf("(%d)", s.PendingCount))
	case tasks.SyncPhaseSynced:
		if !s.LastSyncTime.IsZero() {
			return syncStatusDetailStyle.Render(formatSyncAge(s.LastSyncTime))
		}
	case tasks.SyncPhaseError:
		// Don't show error details in the compact bar
	}
	return ""
}

// formatSyncAge returns a human-readable age string for the last sync time.
func formatSyncAge(t time.Time) string {
	age := time.Since(t)
	switch {
	case age < time.Minute:
		return "just now"
	case age < time.Hour:
		m := int(age.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case age < 24*time.Hour:
		h := int(age.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	default:
		d := int(age.Hours() / 24)
		if d == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", d)
	}
}

// Sync status styles
var (
	syncStatusBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	syncStatusSyncedStyle = lipgloss.NewStyle().
				Foreground(colorComplete)

	syncStatusSyncingStyle = lipgloss.NewStyle().
				Foreground(colorInProgress)

	syncStatusPendingStyle = lipgloss.NewStyle().
				Foreground(colorInProgress)

	syncStatusErrorStyle = lipgloss.NewStyle().
				Foreground(colorBlocked).
				Bold(true)

	syncStatusLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250"))

	syncStatusDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	syncStatusSeparator = "  "
)
