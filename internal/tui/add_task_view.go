package tui

import (
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// addTaskStep tracks which step of the multi-step add flow we're on.
type addTaskStep int

const (
	stepTaskText addTaskStep = iota
	stepContext
)

// AddTaskView handles inline task creation when :add is used without arguments,
// and multi-step context capture when :add-ctx or :add --why is used.
type AddTaskView struct {
	textInput    textinput.Model
	width        int
	withContext  bool
	step         addTaskStep
	capturedText string
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

// NewAddTaskWithContextView creates an AddTaskView that uses the multi-step
// context capture flow (step 1: task text, step 2: why/context).
func NewAddTaskWithContextView() *AddTaskView {
	av := NewAddTaskView()
	av.withContext = true
	return av
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

			if av.step == stepTaskText {
				if text == "" {
					return func() tea.Msg {
						return FlashMsg{Text: "Task text cannot be empty"}
					}
				}
				if av.withContext {
					av.capturedText = text
					av.step = stepContext
					av.textInput.SetValue("")
					av.textInput.Placeholder = "Why does this matter? (Enter to skip)"
					av.textInput.CharLimit = 500
					return nil
				}
				cleanText, tt, ef, loc := tasks.ParseInlineTags(text)
				if cleanText == "" {
					cleanText = text
				}
				newTask := tasks.NewTask(cleanText)
				newTask.Type = tt
				newTask.Effort = ef
				newTask.Location = loc
				return func() tea.Msg {
					return TaskAddedMsg{Task: newTask}
				}
			}

			// step == stepContext
			cleanText, tt, ef, loc := tasks.ParseInlineTags(av.capturedText)
			if cleanText == "" {
				cleanText = av.capturedText
			}
			newTask := tasks.NewTaskWithContext(cleanText, text)
			newTask.Type = tt
			newTask.Effort = ef
			newTask.Location = loc
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

	if av.withContext {
		s.WriteString(headerStyle.Render("ThreeDoors - Add Task with Context"))
	} else {
		s.WriteString(headerStyle.Render("ThreeDoors - Add Task"))
	}
	s.WriteString("\n\n")

	if av.step == stepTaskText {
		s.WriteString(helpStyle.Render("Step 1: What's the task?"))
	} else {
		s.WriteString(helpStyle.Render("Step 2: Why does this matter?"))
	}
	s.WriteString("\n\n")
	s.WriteString(av.textInput.View())
	s.WriteString("\n\n")

	if av.step == stepContext {
		s.WriteString(helpStyle.Render("Enter submit | Enter (empty) skip | Esc cancel"))
	} else {
		s.WriteString(helpStyle.Render("Enter submit | Esc cancel"))
	}

	return s.String()
}
