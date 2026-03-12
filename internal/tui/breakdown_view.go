package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/intelligence/services"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BreakdownView displays proposed subtasks from LLM breakdown for selection and import.
type BreakdownView struct {
	parentTask  *core.Task
	subtasks    []services.Subtask
	selected    []bool
	cursorIndex int
	width       int
	loading     bool
	errorMsg    string
}

// NewBreakdownView creates a BreakdownView for the given breakdown result.
func NewBreakdownView(parentTask *core.Task, result *services.BreakdownResult) *BreakdownView {
	selected := make([]bool, len(result.Subtasks))
	for i := range selected {
		selected[i] = true // all selected by default
	}
	return &BreakdownView{
		parentTask: parentTask,
		subtasks:   result.Subtasks,
		selected:   selected,
	}
}

// NewBreakdownViewLoading creates a BreakdownView in loading state.
func NewBreakdownViewLoading(parentTask *core.Task) *BreakdownView {
	return &BreakdownView{
		parentTask: parentTask,
		loading:    true,
	}
}

// SetWidth sets the terminal width for rendering.
func (bv *BreakdownView) SetWidth(w int) {
	bv.width = w
}

// SetResult populates the view with breakdown results.
func (bv *BreakdownView) SetResult(result *services.BreakdownResult) {
	bv.loading = false
	bv.subtasks = result.Subtasks
	bv.selected = make([]bool, len(result.Subtasks))
	for i := range bv.selected {
		bv.selected[i] = true
	}
	bv.cursorIndex = 0
}

// SetError sets the error state.
func (bv *BreakdownView) SetError(errMsg string) {
	bv.loading = false
	bv.errorMsg = errMsg
}

// SelectedCount returns the number of currently selected subtasks.
func (bv *BreakdownView) SelectedCount() int {
	count := 0
	for _, s := range bv.selected {
		if s {
			count++
		}
	}
	return count
}

// SelectedSubtasks returns the subtasks that are currently selected.
func (bv *BreakdownView) SelectedSubtasks() []services.Subtask {
	var result []services.Subtask
	for i, s := range bv.selected {
		if s {
			result = append(result, bv.subtasks[i])
		}
	}
	return result
}

// Update handles key input in the breakdown view.
func (bv *BreakdownView) Update(msg tea.Msg) tea.Cmd {
	if bv.loading {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				return func() tea.Msg { return BreakdownCancelMsg{} }
			}
		}
		return nil
	}

	if bv.errorMsg != "" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			return func() tea.Msg { _ = msg; return BreakdownCancelMsg{} }
		}
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return bv.handleKey(msg)
	}
	return nil
}

func (bv *BreakdownView) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		return func() tea.Msg { return BreakdownCancelMsg{} }
	case "up", "k":
		if bv.cursorIndex > 0 {
			bv.cursorIndex--
		}
	case "down", "j":
		if bv.cursorIndex < len(bv.subtasks)-1 {
			bv.cursorIndex++
		}
	case " ":
		if bv.cursorIndex >= 0 && bv.cursorIndex < len(bv.selected) {
			bv.selected[bv.cursorIndex] = !bv.selected[bv.cursorIndex]
		}
	case "a":
		allSelected := bv.SelectedCount() == len(bv.subtasks)
		for i := range bv.selected {
			bv.selected[i] = !allSelected
		}
	case "enter":
		selectedSubs := bv.SelectedSubtasks()
		if len(selectedSubs) == 0 {
			return nil
		}
		parentTask := bv.parentTask
		return func() tea.Msg {
			return BreakdownImportMsg{
				ParentTask: parentTask,
				Subtasks:   selectedSubs,
			}
		}
	}
	return nil
}

// View renders the breakdown view.
func (bv *BreakdownView) View() string {
	var s strings.Builder

	w := bv.width - 6
	if w < 40 {
		w = 40
	}

	s.WriteString(breakdownHeaderStyle.Render("TASK BREAKDOWN"))
	s.WriteString("\n\n")

	// Parent task context
	parentText := bv.parentTask.Text
	if len(parentText) > 80 {
		parentText = parentText[:77] + "..."
	}
	fmt.Fprintf(&s, "Parent: %s\n\n", breakdownParentStyle.Render(parentText))

	if bv.loading {
		s.WriteString(breakdownLoadingStyle.Render("Breaking down task..."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("[Esc] Cancel"))
		return detailBorder.Width(w).Render(s.String())
	}

	if bv.errorMsg != "" {
		fmt.Fprintf(&s, "%s\n\n", breakdownErrorStyle.Render(bv.errorMsg))
		s.WriteString(helpStyle.Render("Press any key to return"))
		return detailBorder.Width(w).Render(s.String())
	}

	fmt.Fprintf(&s, "Subtasks (%d/%d selected):\n\n", bv.SelectedCount(), len(bv.subtasks))

	for i, st := range bv.subtasks {
		cursor := "  "
		if i == bv.cursorIndex {
			cursor = "> "
		}
		checkbox := "[ ]"
		if bv.selected[i] {
			checkbox = "[x]"
		}
		effortBadge := ""
		if st.EffortEstimate != "" {
			effortBadge = " " + breakdownEffortStyle.Render(st.EffortEstimate)
		}
		fmt.Fprintf(&s, "%s%s %s%s\n", cursor, checkbox, st.Text, effortBadge)
	}

	s.WriteString("\n")
	s.WriteString(separatorStyle.Render("─────────────────────────────────"))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("[Space] Toggle  [A]ll  [Enter] Import  [Esc] Cancel"))

	return detailBorder.Width(w).Render(s.String())
}

// Styles for the breakdown view.
var (
	breakdownHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	breakdownParentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Bold(true)

	breakdownLoadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Italic(true)

	breakdownErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	breakdownEffortStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)
)
