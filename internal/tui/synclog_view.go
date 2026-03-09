package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// syncLogHeaderHeight is the number of lines consumed by the title header above the viewport.
const syncLogHeaderHeight = 2

// syncLogFooterHeight is the number of lines consumed by the footer below the viewport.
const syncLogFooterHeight = 3

// SyncLogView displays a scrollable list of sync log entries using bubbles/viewport.
type SyncLogView struct {
	entries  []core.SyncLogEntry
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewSyncLogView creates a new SyncLogView with the given entries.
func NewSyncLogView(entries []core.SyncLogEntry) *SyncLogView {
	// Show newest first
	reversed := make([]core.SyncLogEntry, len(entries))
	for i, e := range entries {
		reversed[len(entries)-1-i] = e
	}
	sv := &SyncLogView{
		entries: reversed,
		width:   80,
		height:  24,
	}
	sv.initViewport()
	return sv
}

// SetWidth sets the terminal width and re-renders content.
func (sv *SyncLogView) SetWidth(w int) {
	sv.width = w
	sv.viewport.Width = w
	sv.viewport.SetContent(sv.renderContent())
}

// SetHeight sets the terminal height and adjusts viewport.
func (sv *SyncLogView) SetHeight(h int) {
	sv.height = h
	vpHeight := h - syncLogHeaderHeight - syncLogFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	sv.viewport.Height = vpHeight
}

// initViewport creates and configures the viewport.
func (sv *SyncLogView) initViewport() {
	vpHeight := sv.height - syncLogHeaderHeight - syncLogFooterHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	sv.viewport = NewScrollableView(sv.width, vpHeight)
	sv.viewport.SetContent(sv.renderContent())
	sv.ready = true
}

// renderContent pre-computes all sync log content as a single string.
func (sv *SyncLogView) renderContent() string {
	if len(sv.entries) == 0 {
		return ""
	}

	var s strings.Builder
	for _, entry := range sv.entries {
		ts := syncLogTimestampStyle.Render(entry.Timestamp.Format("2006-01-02 15:04:05"))

		var line string
		switch entry.Operation {
		case "sync":
			line = syncLogEntryStyle.Render(fmt.Sprintf("[%s] %s", entry.Provider, entry.Summary))
		case "conflict_resolved":
			line = syncLogEntryStyle.Render(fmt.Sprintf("[%s] Conflict: %s → %s", entry.Provider, entry.TaskText, entry.Resolution))
		case "error":
			line = syncLogErrorStyle.Render(fmt.Sprintf("[%s] ERROR: %s", entry.Provider, entry.Error))
		default:
			line = syncLogEntryStyle.Render(fmt.Sprintf("[%s] %s", entry.Provider, entry.Summary))
		}

		fmt.Fprintf(&s, "  %s  %s\n", ts, line)
	}
	return s.String()
}

// Update handles key presses for scrolling and closing.
func (sv *SyncLogView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}

	var cmd tea.Cmd
	sv.viewport, cmd = sv.viewport.Update(msg)
	return cmd
}

// View renders the sync log.
func (sv *SyncLogView) View() string {
	var s strings.Builder

	s.WriteString(syncLogHeaderStyle.Render("Sync Log"))
	s.WriteString("\n\n")

	if len(sv.entries) == 0 {
		s.WriteString(helpStyle.Render("No sync operations recorded yet."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("q/Esc to return"))
		return s.String()
	}

	s.WriteString(sv.viewport.View())

	fmt.Fprintf(&s, "\n\n  %3.f%%", sv.viewport.ScrollPercent()*100) //nolint:mnd
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("j/k scroll | PgUp/PgDn page | q/Esc return"))

	return s.String()
}
