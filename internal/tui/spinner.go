package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// spinnerThreshold is the minimum duration before a spinner is shown.
// Operations completing in under this time show no spinner to avoid UI flash.
const spinnerThreshold = 100 * time.Millisecond

// SyncSpinner wraps a bubbles/spinner with ThreeDoors-specific lifecycle:
// start/stop tracking, 100ms display threshold, and provider name context.
type SyncSpinner struct {
	model        spinner.Model
	active       bool
	providerName string
	startTime    time.Time
}

// NewSyncSpinner creates a spinner with ThreeDoors default styling.
func NewSyncSpinner() *SyncSpinner {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(colorInProgress)
	return &SyncSpinner{
		model: s,
	}
}

// Start activates the spinner for the given provider.
func (s *SyncSpinner) Start(provider string) {
	s.active = true
	s.providerName = provider
	s.startTime = time.Now().UTC()
}

// Stop deactivates the spinner.
func (s *SyncSpinner) Stop() {
	s.active = false
	s.providerName = ""
}

// Active returns whether the spinner is currently running.
func (s *SyncSpinner) Active() bool {
	return s.active
}

// ProviderName returns the name of the provider being synced.
func (s *SyncSpinner) ProviderName() string {
	return s.providerName
}

// ThresholdElapsed returns true if enough time has passed since Start
// to warrant showing the spinner (avoids flash for instant operations).
func (s *SyncSpinner) ThresholdElapsed() bool {
	if !s.active {
		return false
	}
	return time.Since(s.startTime) >= spinnerThreshold
}

// View returns the spinner animation frame, or empty string if
// the spinner is inactive or the display threshold has not elapsed.
func (s *SyncSpinner) View() string {
	if !s.active || !s.ThresholdElapsed() {
		return ""
	}
	return s.model.View()
}

// Update processes a spinner tick message and returns the next command.
func (s *SyncSpinner) Update(msg tea.Msg) tea.Cmd {
	if !s.active {
		return nil
	}
	var cmd tea.Cmd
	s.model, cmd = s.model.Update(msg)
	return cmd
}

// Tick returns the initial tick command to start spinner animation.
func (s *SyncSpinner) Tick() tea.Cmd {
	return s.model.Tick
}

// RenderSyncStatusBarWithSpinner renders the sync status bar, replacing
// the static syncing icon with an animated spinner for the active provider.
func RenderSyncStatusBarWithSpinner(tracker *core.SyncStatusTracker, sp *SyncSpinner) string {
	if tracker == nil || tracker.Count() == 0 {
		return ""
	}

	// Fall back to regular rendering when spinner is nil or inactive
	if sp == nil || !sp.Active() || !sp.ThresholdElapsed() {
		return RenderSyncStatusBar(tracker)
	}

	statuses := tracker.All()
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	var parts []string
	for _, s := range statuses {
		if s.Name == sp.ProviderName() && s.Phase == core.SyncPhaseSyncing {
			// Replace the static syncing icon with the animated spinner
			label := syncStatusLabelStyle.Render(s.Name)
			parts = append(parts, fmt.Sprintf("%s %s", sp.View(), label))
		} else {
			parts = append(parts, renderProviderStatus(s))
		}
	}

	bar := strings.Join(parts, syncStatusSeparator)

	walLine := renderWALPending(statuses)
	if walLine != "" {
		bar += "\n" + walLine
	}

	return syncStatusBarStyle.Render(bar)
}
