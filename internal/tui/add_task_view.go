package tui

import (
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// AddTaskView handles inline task creation when :add is used without arguments.
type AddTaskView struct {
	textInput textinput.Model
	width     int
}

// NewAddTaskView creates a new AddTaskView with a focused text input.
func NewAddTaskView() *AddTaskView {
	ti := textinput.New()
	ti.Placeholder = "Enter task text..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 40

	return &AddTaskView{
		textInput: ti,
	}
}

// SetWidth sets the terminal width for rendering.
func (av *AddTaskView) SetWidth(w int) {
	av.width = w
	if w > 4 {
		av.textInput.Width = w - 4
	}
}

// Update handles messages for the add task view.
func (av *AddTaskView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return func() tea.Msg { return ReturnToDoorsMsg{} }

		case tea.KeyEnter:
			text := strings.TrimSpace(av.textInput.Value())
			if text == "" {
				return func() tea.Msg {
					return FlashMsg{Text: "Task text cannot be empty"}
				}
			}
			newTask := tasks.NewTask(text)
			return func() tea.Msg {
				return TaskAddedMsg{Task: newTask}
			}
		}
	}

	var cmd tea.Cmd
	av.textInput, cmd = av.textInput.Update(msg)
	return cmd
}

// View renders the add task view.
func (av *AddTaskView) View() string {
	s := strings.Builder{}

	s.WriteString(headerStyle.Render("ThreeDoors - Add Task"))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Enter a new task:"))
	s.WriteString("\n\n")
	s.WriteString(av.textInput.View())
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Enter submit | Esc cancel"))

	return s.String()
}
