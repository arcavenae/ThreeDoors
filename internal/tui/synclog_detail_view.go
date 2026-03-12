package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// syncLogDetailHeaderHeight is the number of lines consumed by the title header above the viewport.
const syncLogDetailHeaderHeight = 2

// syncLogDetailFooterHeight is the number of lines consumed by the footer below the viewport.
const syncLogDetailFooterHeight = 3

// syncLogDetailLimit is the maximum number of events to load.
const syncLogDetailLimit = 200

var (
	syncLogSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#5fff00", ANSI256: "82", ANSI: "2"})

	syncLogWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.CompleteColor{TrueColor: "#ffaf00", ANSI256: "214", ANSI: "3"})
)

// SyncLogDetailView displays a scrollable list of sync events for a specific connection.
type SyncLogDetailView struct {
	connectionID   string
	connectionName string
	events         []connection.SyncEvent
	viewport       viewport.Model
	width          int
	height         int
}

// NewSyncLogDetailView creates a new SyncLogDetailView for a connection.
func NewSyncLogDetailView(connectionID, connectionName string, events []connection.SyncEvent) *SyncLogDetailView {
	sv := &SyncLogDetailView{
		connectionID:   connectionID,
		connectionName: connectionName,
		events:         events,
		width:          80,
		height:         24,
	}
	sv.initViewport()
	return sv
}

// SetWidth sets the terminal width and re-renders content.
func (sv *SyncLogDetailView) SetWidth(w int) {
	sv.width = w
	sv.viewport.Width = w
	sv.viewport.SetContent(sv.renderContent())
}

// SetHeight sets the terminal height and adjusts viewport.
func (sv *SyncLogDetailView) SetHeight(h int) {
	sv.height = h
	vpHeight := h - syncLogDetailHeaderHeight - syncLogDetailFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	sv.viewport.Height = vpHeight
}

// initViewport creates and configures the viewport.
func (sv *SyncLogDetailView) initViewport() {
	vpHeight := sv.height - syncLogDetailHeaderHeight - syncLogDetailFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	sv.viewport = NewScrollableView(sv.width, vpHeight)
	sv.viewport.SetContent(sv.renderContent())
}

// renderContent pre-computes all sync event content as a single string.
func (sv *SyncLogDetailView) renderContent() string {
	if len(sv.events) == 0 {
		return ""
	}

	var s strings.Builder
	for _, event := range sv.events {
		ts := syncLogTimestampStyle.Render(event.Timestamp.Format("2006-01-02 15:04:05"))

		var indicator string
		var line string
		switch event.Type {
		case connection.EventSyncComplete:
			indicator = syncLogSuccessStyle.Render("✓")
			line = syncLogEntryStyle.Render(fmt.Sprintf(
				"Sync complete: +%d -%d ~%d",
				event.Added, event.Removed, event.Updated,
			))
		case connection.EventSyncError:
			indicator = syncLogErrorStyle.Render("✗")
			line = syncLogErrorStyle.Render(event.Error)
		case connection.EventConflict:
			indicator = syncLogWarningStyle.Render("⚠")
			resolution := event.Resolution
			if resolution == "" {
				resolution = "unresolved"
			}
			line = syncLogWarningStyle.Render(fmt.Sprintf(
				"Conflict on '%s' — %s",
				event.ConflictTaskText, resolution,
			))
		case connection.EventReauthRequired:
			indicator = syncLogWarningStyle.Render("⚠")
			line = syncLogWarningStyle.Render("Re-authentication required")
			if event.Error != "" {
				line = syncLogWarningStyle.Render(fmt.Sprintf("Re-auth required: %s", event.Error))
			}
		case connection.EventStateChange:
			indicator = syncLogEntryStyle.Render("→")
			line = syncLogEntryStyle.Render(fmt.Sprintf(
				"State: %s → %s",
				event.FromState, event.ToState,
			))
		case connection.EventSyncStart:
			indicator = syncLogEntryStyle.Render("⟳")
			line = syncLogEntryStyle.Render("Sync started")
		default:
			indicator = syncLogEntryStyle.Render("·")
			line = syncLogEntryStyle.Render(event.Summary)
		}

		fmt.Fprintf(&s, "  %s  %s  %s\n", ts, indicator, line)
	}
	return s.String()
}

// Update handles key presses for scrolling and closing.
func (sv *SyncLogDetailView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Navigate back to source detail if it exists, otherwise sources dashboard.
			return func() tea.Msg {
				return ShowSourceDetailMsg{ConnectionID: sv.connectionID}
			}
		}
	}

	var cmd tea.Cmd
	sv.viewport, cmd = sv.viewport.Update(msg)
	return cmd
}

// View renders the sync log detail view.
func (sv *SyncLogDetailView) View() string {
	var s strings.Builder

	title := "Sync Log"
	if sv.connectionName != "" {
		title = fmt.Sprintf("Sync Log — %s", sv.connectionName)
	}
	s.WriteString(syncLogHeaderStyle.Render(title))
	s.WriteString("\n\n")

	if len(sv.events) == 0 {
		s.WriteString(helpStyle.Render("No sync events yet"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Esc to return"))
		return s.String()
	}

	s.WriteString(sv.viewport.View())

	fmt.Fprintf(&s, "\n\n  %3.f%%", sv.viewport.ScrollPercent()*100) //nolint:mnd
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("j/k scroll | PgUp/PgDn page | Esc return"))

	return s.String()
}
