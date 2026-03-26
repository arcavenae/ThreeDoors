package tui

import (
	"fmt"
	"strings"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// AvoidancePromptView displays a gentle prompt for tasks with 10+ bypasses.
type AvoidancePromptView struct {
	task   *core.Task
	count  int
	width  int
	height int
}

// NewAvoidancePromptView creates an avoidance prompt for the given task.
func NewAvoidancePromptView(task *core.Task, bypassCount int) *AvoidancePromptView {
	return &AvoidancePromptView{
		task:  task,
		count: bypassCount,
	}
}

// SetWidth sets the terminal width for rendering.
func (v *AvoidancePromptView) SetWidth(w int) {
	v.width = w
}

// SetHeight sets the terminal height for layout decisions.
func (v *AvoidancePromptView) SetHeight(h int) {
	v.height = h
}

// Update handles key input for the avoidance prompt.
func (v *AvoidancePromptView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			return func() tea.Msg {
				return AvoidanceActionMsg{Task: v.task, Action: "reconsider"}
			}
		case "b", "B":
			return func() tea.Msg {
				return AvoidanceActionMsg{Task: v.task, Action: "breakdown"}
			}
		case "d", "D":
			return func() tea.Msg {
				return AvoidanceActionMsg{Task: v.task, Action: "defer"}
			}
		case "a", "A":
			return func() tea.Msg {
				return AvoidanceActionMsg{Task: v.task, Action: "archive"}
			}
		case "esc", "escape", "q":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}
	return nil
}

// View renders the avoidance prompt.
func (v *AvoidancePromptView) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("ThreeDoors"))
	s.WriteString("\n\n")

	taskText := v.task.Text
	if len(taskText) > 60 {
		taskText = taskText[:57] + "..."
	}

	fmt.Fprintf(&s, "  %s\n\n", taskText)
	fmt.Fprintf(&s, "  This task has appeared %d times.\n", v.count)
	s.WriteString("  What would you like to do?\n\n")
	s.WriteString("  [R] Reconsider - take it on now\n")
	s.WriteString("  [B] Break down - look at it closer\n")
	s.WriteString("  [D] Defer - set aside for later\n")
	s.WriteString("  [A] Archive - remove from your list\n")
	s.WriteString("  [Esc] Dismiss\n")

	return s.String()
}
