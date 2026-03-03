package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

const syncLogPageSize = 20

// SyncLogView displays a scrollable list of sync log entries.
type SyncLogView struct {
	entries []tasks.SyncLogEntry
	offset  int
	width   int
}

// NewSyncLogView creates a new SyncLogView with the given entries.
func NewSyncLogView(entries []tasks.SyncLogEntry) *SyncLogView {
	// Show newest first
	reversed := make([]tasks.SyncLogEntry, len(entries))
	for i, e := range entries {
		reversed[len(entries)-1-i] = e
	}
	return &SyncLogView{
		entries: reversed,
	}
}

// SetWidth sets the terminal width for rendering.
func (sv *SyncLogView) SetWidth(w int) {
	sv.width = w
}

// Update handles key presses for scrolling and closing.
func (sv *SyncLogView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		case "j", "down":
			if sv.offset+syncLogPageSize < len(sv.entries) {
				sv.offset++
			}
		case "k", "up":
			if sv.offset > 0 {
				sv.offset--
			}
		case "pgdown", " ":
			sv.offset += syncLogPageSize
			if sv.offset+syncLogPageSize > len(sv.entries) {
				sv.offset = max(0, len(sv.entries)-syncLogPageSize)
			}
		case "pgup":
			sv.offset -= syncLogPageSize
			if sv.offset < 0 {
				sv.offset = 0
			}
		}
	}
	return nil
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

	end := sv.offset + syncLogPageSize
	if end > len(sv.entries) {
		end = len(sv.entries)
	}
	visible := sv.entries[sv.offset:end]

	for _, entry := range visible {
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

	fmt.Fprintf(&s, "\n  Showing %d-%d of %d entries", sv.offset+1, end, len(sv.entries))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("j/k scroll | PgUp/PgDn page | q/Esc return"))

	return s.String()
}
