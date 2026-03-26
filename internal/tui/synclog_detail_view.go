package tui

import (
	"fmt"
	"strings"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// syncLogDetailHeaderHeight is the number of lines consumed by the title header above the viewport.
const syncLogDetailHeaderHeight = 2

// syncLogDetailFooterHeight is the number of lines consumed by the footer below the viewport.
const syncLogDetailFooterHeight = 3

// SyncLogDetailView displays a scrollable list of sync events for a specific connection.
type SyncLogDetailView struct {
	connectionID string
	events       []connection.SyncEvent
	viewport     viewport.Model
	width        int
	height       int
	ready        bool
}

// NewSyncLogDetailView creates a SyncLogDetailView for the given connection's events.
// Events should already be in reverse chronological order (newest first).
func NewSyncLogDetailView(connectionID string, events []connection.SyncEvent) *SyncLogDetailView {
	v := &SyncLogDetailView{
		connectionID: connectionID,
		events:       events,
		width:        80,
		height:       24,
	}
	v.initViewport()
	return v
}

// SetWidth sets the terminal width and re-renders content.
func (v *SyncLogDetailView) SetWidth(w int) {
	v.width = w
	v.viewport.Width = w
	v.viewport.SetContent(v.renderContent())
}

// SetHeight sets the terminal height and adjusts viewport.
func (v *SyncLogDetailView) SetHeight(h int) {
	v.height = h
	vpHeight := h - syncLogDetailHeaderHeight - syncLogDetailFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	v.viewport.Height = vpHeight
}

func (v *SyncLogDetailView) initViewport() {
	vpHeight := v.height - syncLogDetailHeaderHeight - syncLogDetailFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	v.viewport = NewScrollableView(v.width, vpHeight)
	v.viewport.SetContent(v.renderContent())
	v.ready = true
}

func (v *SyncLogDetailView) renderContent() string {
	if len(v.events) == 0 {
		return ""
	}

	var s strings.Builder
	for _, event := range v.events {
		ts := syncLogTimestampStyle.Render(event.Timestamp.Format("2006-01-02 15:04:05"))
		indicator := eventIndicator(event.Type)
		detail := eventDetail(event)
		fmt.Fprintf(&s, "  %s  %s %s\n", ts, indicator, detail)
	}
	return s.String()
}

// eventIndicator returns a styled status indicator for the event type.
func eventIndicator(t connection.SyncEventType) string {
	switch t {
	case connection.EventSyncComplete:
		return syncLogEntryStyle.Render("✓")
	case connection.EventSyncError:
		return syncLogErrorStyle.Render("✗")
	case connection.EventConflict:
		return syncLogEntryStyle.Render("⚠")
	case connection.EventReauthRequired:
		return syncLogErrorStyle.Render("⚠")
	case connection.EventStateChange:
		return syncLogEntryStyle.Render("→")
	case connection.EventSyncStart:
		return syncLogEntryStyle.Render("⟳")
	default:
		return syncLogEntryStyle.Render("·")
	}
}

// eventDetail returns a human-readable detail string for the event.
func eventDetail(e connection.SyncEvent) string {
	switch e.Type {
	case connection.EventSyncComplete:
		detail := fmt.Sprintf("+%d /%d -%d", e.Added, e.Updated, e.Removed)
		return syncLogEntryStyle.Render(detail)
	case connection.EventSyncError:
		return syncLogErrorStyle.Render(e.Error)
	case connection.EventConflict:
		res := fmt.Sprintf("%s → %s", e.ConflictTaskText, e.Resolution)
		return syncLogEntryStyle.Render(res)
	case connection.EventReauthRequired:
		return syncLogErrorStyle.Render(e.Error)
	case connection.EventStateChange:
		return syncLogEntryStyle.Render(fmt.Sprintf("%s → %s", e.FromState, e.ToState))
	case connection.EventSyncStart:
		return syncLogEntryStyle.Render("sync started")
	default:
		return syncLogEntryStyle.Render(e.Summary)
	}
}

// Update handles key presses for scrolling and closing.
func (v *SyncLogDetailView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return func() tea.Msg {
				return ShowSourceDetailMsg{ConnectionID: v.connectionID}
			}
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

// View renders the sync log detail view.
func (v *SyncLogDetailView) View() string {
	var s strings.Builder

	s.WriteString(syncLogHeaderStyle.Render("Sync Log"))
	s.WriteString("\n\n")

	if len(v.events) == 0 {
		s.WriteString(helpStyle.Render("No sync events yet"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Esc to return"))
		return s.String()
	}

	s.WriteString(v.viewport.View())

	fmt.Fprintf(&s, "\n\n  %3.f%%", v.viewport.ScrollPercent()*100) //nolint:mnd
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("j/k scroll | PgUp/PgDn page | Esc return"))

	return s.String()
}
