package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/lipgloss"
)

// OfflineInfo holds connectivity and queue information for TUI display (AC6).
type OfflineInfo struct {
	Online        bool
	Probing       bool
	UnpushedCount int
	LastSyncTime  time.Time
}

// RenderSyncStatusBar renders a compact sync status bar for all tracked providers.
// Returns an empty string if no providers are registered.
func RenderSyncStatusBar(tracker *core.SyncStatusTracker) string {
	return RenderSyncStatusBarFull(tracker, nil)
}

// RenderSyncStatusBarFull renders the sync status bar with optional offline info.
func RenderSyncStatusBarFull(tracker *core.SyncStatusTracker, info *OfflineInfo) string {
	if tracker == nil || tracker.Count() == 0 {
		if info == nil {
			return ""
		}
	}

	var parts []string

	// Render connectivity indicator (AC6)
	if info != nil {
		parts = append(parts, renderOfflineInfo(*info))
	}

	if tracker != nil && tracker.Count() > 0 {
		statuses := tracker.All()
		// Sort by name for deterministic rendering
		sort.Slice(statuses, func(i, j int) bool {
			return statuses[i].Name < statuses[j].Name
		})

		for _, s := range statuses {
			parts = append(parts, renderProviderStatus(s))
		}

		bar := strings.Join(parts, syncStatusSeparator)

		// Append WAL pending summary if any provider has pending items
		walLine := renderWALPending(statuses)
		if walLine != "" {
			bar += "\n" + walLine
		}

		return syncStatusBarStyle.Render(bar)
	}

	return syncStatusBarStyle.Render(strings.Join(parts, syncStatusSeparator))
}

// renderOfflineInfo renders the online/offline indicator with queue depth.
func renderOfflineInfo(info OfflineInfo) string {
	var icon, label string

	switch {
	case info.Probing:
		icon = syncStatusHalfOpenStyle.Render("↻")
		label = syncStatusLabelStyle.Render("probing")
	case info.Online:
		icon = syncStatusSyncedStyle.Render("●")
		label = syncStatusLabelStyle.Render("online")
	default:
		icon = syncStatusErrorStyle.Render("○")
		label = syncStatusLabelStyle.Render("offline")
	}

	result := fmt.Sprintf("%s %s", icon, label)

	if info.UnpushedCount > 0 {
		result += " " + syncStatusPendingStyle.Render(fmt.Sprintf("(%d queued)", info.UnpushedCount))
	}

	if !info.LastSyncTime.IsZero() {
		result += " " + syncStatusDetailStyle.Render(formatSyncAge(info.LastSyncTime))
	}

	return result
}

// renderProviderStatus renders a single provider's status with appropriate styling.
func renderProviderStatus(s core.ProviderSyncStatus) string {
	icon := s.Icon()
	styledIcon := styleIcon(icon, s)
	label := syncStatusLabelStyle.Render(s.Name)
	detail := renderDetail(s)

	if detail != "" {
		return fmt.Sprintf("%s %s %s", styledIcon, label, detail)
	}
	return fmt.Sprintf("%s %s", styledIcon, label)
}

// styleIcon applies the appropriate color to the icon based on circuit and phase state.
func styleIcon(icon string, s core.ProviderSyncStatus) string {
	switch s.CircuitState {
	case core.CircuitOpen:
		return syncStatusErrorStyle.Render(icon)
	case core.CircuitHalfOpen:
		return syncStatusHalfOpenStyle.Render(icon)
	default:
		switch s.Phase {
		case core.SyncPhaseSynced:
			return syncStatusSyncedStyle.Render(icon)
		case core.SyncPhaseSyncing:
			return syncStatusSyncingStyle.Render(icon)
		case core.SyncPhasePending:
			return syncStatusPendingStyle.Render(icon)
		case core.SyncPhaseError:
			return syncStatusErrorStyle.Render(icon)
		default:
			return icon
		}
	}
}

// renderDetail renders extra information based on circuit and sync phase.
func renderDetail(s core.ProviderSyncStatus) string {
	// Circuit state takes priority
	switch s.CircuitState {
	case core.CircuitOpen:
		if s.RetryIn > 0 {
			return syncStatusDetailStyle.Render(fmt.Sprintf("error (retry in %s)", core.FormatDuration(s.RetryIn)))
		}
		return syncStatusDetailStyle.Render("error")
	case core.CircuitHalfOpen:
		return syncStatusDetailStyle.Render("probing...")
	}

	// Staleness indicator
	if !s.StaleSince.IsZero() {
		age := time.Since(s.StaleSince)
		return syncStatusStaleStyle.Render(fmt.Sprintf("stale %s", core.FormatDuration(age)))
	}

	switch s.Phase {
	case core.SyncPhasePending:
		return syncStatusDetailStyle.Render(fmt.Sprintf("(%d)", s.PendingCount))
	case core.SyncPhaseSynced:
		if !s.LastSyncTime.IsZero() {
			return syncStatusDetailStyle.Render(formatSyncAge(s.LastSyncTime))
		}
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

// renderWALPending returns a WAL pending line if any provider has pending items.
func renderWALPending(statuses []core.ProviderSyncStatus) string {
	totalPending := 0
	var oldestTime time.Time
	for _, s := range statuses {
		totalPending += s.PendingCount
		if !s.OldestPending.IsZero() && (oldestTime.IsZero() || s.OldestPending.Before(oldestTime)) {
			oldestTime = s.OldestPending
		}
	}
	if totalPending == 0 {
		return ""
	}
	if !oldestTime.IsZero() {
		age := time.Since(oldestTime)
		return syncStatusPendingStyle.Render(
			fmt.Sprintf("WAL pending (%d items, oldest %s)", totalPending, core.FormatDuration(age)),
		)
	}
	return syncStatusPendingStyle.Render(
		fmt.Sprintf("WAL pending (%d items)", totalPending),
	)
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

	syncStatusHalfOpenStyle = lipgloss.NewStyle().
				Foreground(colorInProgress)

	syncStatusDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	syncStatusStaleStyle = lipgloss.NewStyle().
				Foreground(colorInProgress).
				Bold(true)

	syncStatusSeparator = "  "
)
