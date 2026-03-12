package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConflictView renders a side-by-side conflict resolution interface.
type ConflictView struct {
	conflictSet *core.ConflictSet
	syncLog     *core.SyncLog
	width       int
	height      int
}

// NewConflictView creates a new ConflictView for the given conflict set.
func NewConflictView(cs *core.ConflictSet, syncLog *core.SyncLog) *ConflictView {
	return &ConflictView{
		conflictSet: cs,
		syncLog:     syncLog,
	}
}

// SetWidth sets the terminal width for rendering.
func (cv *ConflictView) SetWidth(w int) {
	cv.width = w
}

// SetHeight sets the terminal height for layout decisions.
func (cv *ConflictView) SetHeight(h int) {
	cv.height = h
}

// Update handles key presses for conflict resolution.
func (cv *ConflictView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l", "L":
			return cv.resolve(core.ChoiceKeepLocal)
		case "r", "R":
			return cv.resolve(core.ChoiceKeepRemote)
		case "b", "B":
			return cv.resolve(core.ChoiceKeepBoth)
		case "q", "esc":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}
	return nil
}

func (cv *ConflictView) resolve(choice core.ConflictChoice) tea.Cmd {
	current := cv.conflictSet.CurrentConflict()
	if current == nil {
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	}

	if cv.syncLog != nil {
		_ = cv.syncLog.LogConflictResolution(
			cv.conflictSet.Provider,
			current.Conflict.LocalTask.ID,
			current.Conflict.LocalTask.Text,
			string(choice),
		)
	}

	if err := cv.conflictSet.Resolve(choice); err != nil {
		return func() tea.Msg {
			return FlashMsg{Text: fmt.Sprintf("Error resolving conflict: %v", err)}
		}
	}

	if cv.conflictSet.AllResolved() {
		cs := cv.conflictSet
		return func() tea.Msg {
			return ConflictResolvedMsg{ConflictSet: cs}
		}
	}

	return nil
}

// View renders the conflict resolution interface.
func (cv *ConflictView) View() string {
	var s strings.Builder

	ic := cv.conflictSet.CurrentConflict()
	if ic == nil {
		s.WriteString(conflictHeaderStyle.Render("All conflicts resolved!"))
		s.WriteString("\n\nPress any key to return.")
		return s.String()
	}

	total := len(cv.conflictSet.Conflicts)
	current := cv.conflictSet.Current + 1

	s.WriteString(conflictHeaderStyle.Render(fmt.Sprintf("⚠ Sync Conflict (%d/%d)", current, total)))
	s.WriteString("\n\n")

	local := ic.Conflict.LocalTask
	remote := ic.Conflict.RemoteTask

	// Side-by-side comparison
	colWidth := 35
	if cv.width > 20 {
		colWidth = (cv.width - 8) / 2
		if colWidth < 20 {
			colWidth = 20
		}
	}

	localContent := formatTaskForConflict(local, "LOCAL")
	remoteContent := formatTaskForConflict(remote, "REMOTE")

	// Highlight differences
	var diffLines []string
	if local.Text != remote.Text {
		diffLines = append(diffLines, conflictDiffStyle.Render("Text differs"))
	}
	if local.Status != remote.Status {
		diffLines = append(diffLines, conflictDiffStyle.Render(fmt.Sprintf("Status: %s vs %s", local.Status, remote.Status)))
	}

	localBox := conflictLocalStyle.Width(colWidth).Render(localContent)
	remoteBox := conflictRemoteStyle.Width(colWidth).Render(remoteContent)

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, localBox, "  ", remoteBox))

	if len(diffLines) > 0 {
		s.WriteString("\n\n")
		s.WriteString(conflictDiffStyle.Render("Differences:"))
		s.WriteString("\n")
		for _, line := range diffLines {
			fmt.Fprintf(&s, "  %s\n", line)
		}
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("L keep local | R keep remote | B keep both | q/Esc cancel"))

	return s.String()
}

func formatTaskForConflict(t *core.Task, label string) string {
	var s strings.Builder
	fmt.Fprintf(&s, "%s\n", label)
	fmt.Fprintf(&s, "─────────\n")
	fmt.Fprintf(&s, "Text: %s\n", t.Text)
	fmt.Fprintf(&s, "Status: %s\n", t.Status)
	fmt.Fprintf(&s, "Updated: %s", t.UpdatedAt.Format("2006-01-02 15:04"))
	if len(t.Notes) > 0 {
		fmt.Fprintf(&s, "\nNotes: %d", len(t.Notes))
	}
	return s.String()
}
