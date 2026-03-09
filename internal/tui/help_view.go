package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const helpPageSize = 20

// helpEntry represents a single key/command → description mapping.
type helpEntry struct {
	Key         string
	Description string
}

// helpSection groups related help entries under a header.
type helpSection struct {
	Title   string
	Entries []helpEntry
}

// helpContent defines all help sections as structured data.
var helpContent = []helpSection{
	{
		Title: "Navigation",
		Entries: []helpEntry{
			{"a / Left", "Select left door (again to deselect)"},
			{"w / Up", "Select center door (again to deselect)"},
			{"d / Right", "Select right door (again to deselect)"},
			{"s / Down", "Re-roll doors (new random selection)"},
			{"Enter / Space", "Open selected door (detail view)"},
			{"j / k", "Scroll down/up (in scrollable views)"},
			{"PgDn / Space", "Page down (in scrollable views)"},
			{"PgUp", "Page up (in scrollable views)"},
			{"Esc", "Return to previous view"},
		},
	},
	{
		Title: "Task Actions",
		Entries: []helpEntry{
			{"C", "Complete task (in detail view)"},
			{"B", "Mark task as blocked (in detail view)"},
			{"I", "Mark task as in-progress (in detail view)"},
			{"E", "Expand task — create subtask (in detail view)"},
			{"F", "Fork task — create variant (in detail view)"},
			{"P", "Mark as procrastinated (in detail view)"},
			{"R", "Rework task (in detail view)"},
			{"N", "Give door feedback (from doors view)"},
			{"L", "Link/unlink tasks (in detail view)"},
			{"X", "Show cross-references (in detail view)"},
			{"x", "Dispatch task to dev queue (in detail view)"},
		},
	},
	{
		Title: "Commands",
		Entries: []helpEntry{
			{":add <text>", "Add a new task"},
			{":add-ctx", "Add task with context (why it matters)"},
			{":add --why", "Add task with context prompt"},
			{":tag", "Edit task categories/tags"},
			{":theme", "Open theme picker"},
			{":mood [mood]", "Record current mood"},
			{":stats", "Show session statistics"},
			{":dashboard", "Open insights dashboard"},
			{":insights", "Show pattern insights"},
			{":health", "Run health check"},
			{":synclog", "View sync operation log"},
			{":devqueue", "View dev dispatch queue"},
			{":suggestions", "View AI task suggestions"},
			{":dispatch", "Dev dispatch info"},
			{":goals [edit]", "View or edit values/goals"},
			{":quit / :exit", "Quit application"},
			{":help", "Show this help screen"},
		},
	},
	{
		Title: "Search",
		Entries: []helpEntry{
			{"/", "Open search"},
			{":", "Open command palette"},
			{"Up / Down", "Navigate search results"},
			{"Enter", "Select search result or run command"},
			{"Esc", "Close search"},
		},
	},
	{
		Title: "Global",
		Entries: []helpEntry{
			{"?", "Show this help screen"},
			{"S", "Open suggestions (from doors view)"},
			{"M", "Record mood (from doors view)"},
			{"q", "Quit (from most views)"},
		},
	},
}

// HelpView displays a scrollable categorized help screen.
type HelpView struct {
	lines  []string
	offset int
	width  int
}

// NewHelpView creates a new HelpView with pre-rendered content lines.
func NewHelpView() *HelpView {
	hv := &HelpView{
		width: 80,
	}
	hv.renderLines()
	return hv
}

// SetWidth sets the terminal width and re-renders content.
func (hv *HelpView) SetWidth(w int) {
	hv.width = w
	hv.renderLines()
}

// renderLines pre-computes all output lines from structured help content.
func (hv *HelpView) renderLines() {
	hv.lines = nil

	keyWidth := 20
	// Description column width: total width minus key column minus padding/gutters
	descWidth := hv.width - keyWidth - 6
	if descWidth < 20 {
		descWidth = 20
	}

	for i, section := range helpContent {
		if i > 0 {
			hv.lines = append(hv.lines, "")
		}
		hv.lines = append(hv.lines, syncLogHeaderStyle.Render(section.Title))
		hv.lines = append(hv.lines, "")

		for _, entry := range section.Entries {
			key := headerStyle.Render(fmt.Sprintf("  %-*s", keyWidth, entry.Key))
			wrapped := wordWrap(entry.Description, descWidth)
			parts := strings.Split(wrapped, "\n")
			hv.lines = append(hv.lines, fmt.Sprintf("%s  %s", key, parts[0]))
			for _, cont := range parts[1:] {
				hv.lines = append(hv.lines, fmt.Sprintf("  %-*s  %s", keyWidth, "", cont))
			}
		}
	}
}

// wordWrap wraps text to fit within maxWidth columns.
func wordWrap(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	var lines []string
	words := strings.Fields(text)
	var current strings.Builder

	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
			continue
		}
		if current.Len()+1+len(word) > maxWidth {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		} else {
			current.WriteByte(' ')
			current.WriteString(word)
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return strings.Join(lines, "\n")
}

// Update handles key presses for scrolling and dismissal.
func (hv *HelpView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		case "j", "down":
			if hv.offset < len(hv.lines)-1 {
				hv.offset++
			}
		case "k", "up":
			if hv.offset > 0 {
				hv.offset--
			}
		case "pgdown", " ":
			hv.offset += helpPageSize
			if hv.offset > len(hv.lines)-1 {
				hv.offset = max(0, len(hv.lines)-1)
			}
		case "pgup":
			hv.offset -= helpPageSize
			if hv.offset < 0 {
				hv.offset = 0
			}
		}
	}
	return nil
}

// View renders the help screen.
func (hv *HelpView) View() string {
	var s strings.Builder

	s.WriteString(syncLogHeaderStyle.Render("Help"))
	s.WriteString("\n\n")

	if len(hv.lines) == 0 {
		s.WriteString(helpStyle.Render("No help content available."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("q/Esc to return"))
		return s.String()
	}

	// Calculate visible window
	visibleLines := helpPageSize
	end := hv.offset + visibleLines
	if end > len(hv.lines) {
		end = len(hv.lines)
	}
	visible := hv.lines[hv.offset:end]

	for _, line := range visible {
		s.WriteString(line)
		s.WriteString("\n")
	}

	fmt.Fprintf(&s, "\n  Showing lines %d-%d of %d", hv.offset+1, end, len(hv.lines))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("j/k scroll | PgUp/PgDn page | q/Esc return"))

	return s.String()
}
