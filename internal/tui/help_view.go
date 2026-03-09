package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

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

// headerHeight is the number of lines consumed by the title header above the viewport.
const headerHeight = 2

// footerHeight is the number of lines consumed by the footer below the viewport.
const footerHeight = 3

// HelpView displays a scrollable categorized help screen using bubbles/viewport.
type HelpView struct {
	viewport viewport.Model
	content  string
	width    int
	height   int
	ready    bool
}

// NewHelpView creates a new HelpView with pre-rendered content.
func NewHelpView() *HelpView {
	hv := &HelpView{
		width:  80,
		height: 24,
	}
	hv.renderContent()
	hv.initViewport()
	return hv
}

// SetWidth sets the terminal width and re-renders content.
func (hv *HelpView) SetWidth(w int) {
	hv.width = w
	hv.renderContent()
	hv.viewport.Width = w
	hv.viewport.SetContent(hv.content)
}

// SetHeight sets the terminal height and adjusts viewport.
func (hv *HelpView) SetHeight(h int) {
	hv.height = h
	vpHeight := h - headerHeight - footerHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	hv.viewport.Height = vpHeight
}

// initViewport creates and configures the viewport.
func (hv *HelpView) initViewport() {
	vpHeight := hv.height - headerHeight - footerHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	hv.viewport = NewScrollableView(hv.width, vpHeight)
	hv.viewport.SetContent(hv.content)
	hv.ready = true
}

// renderContent pre-computes all help content as a single string.
func (hv *HelpView) renderContent() {
	keyWidth := 20
	descWidth := hv.width - keyWidth - 6
	if descWidth < 20 {
		descWidth = 20
	}

	var s strings.Builder

	for i, section := range helpContent {
		if i > 0 {
			s.WriteString("\n")
		}
		s.WriteString(syncLogHeaderStyle.Render(section.Title))
		s.WriteString("\n\n")

		for _, entry := range section.Entries {
			key := headerStyle.Render(fmt.Sprintf("  %-*s", keyWidth, entry.Key))
			wrapped := wordWrap(entry.Description, descWidth)
			parts := strings.Split(wrapped, "\n")
			fmt.Fprintf(&s, "%s  %s\n", key, parts[0])
			for _, cont := range parts[1:] {
				fmt.Fprintf(&s, "  %-*s  %s\n", keyWidth, "", cont)
			}
		}
	}

	hv.content = s.String()
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
		}
	}

	var cmd tea.Cmd
	hv.viewport, cmd = hv.viewport.Update(msg)
	return cmd
}

// View renders the help screen.
func (hv *HelpView) View() string {
	var s strings.Builder

	s.WriteString(syncLogHeaderStyle.Render("Help"))
	s.WriteString("\n\n")

	if len(hv.content) == 0 {
		s.WriteString(helpStyle.Render("No help content available."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("q/Esc to return"))
		return s.String()
	}

	s.WriteString(hv.viewport.View())

	fmt.Fprintf(&s, "\n\n  %3.f%%", hv.viewport.ScrollPercent()*100) //nolint:mnd
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("j/k scroll | PgUp/PgDn page | q/Esc return"))

	return s.String()
}
