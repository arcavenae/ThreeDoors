package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

var feedbackOptions = []string{
	"Blocked",
	"Not now",
	"Needs breakdown",
	"Other comment",
}

// FeedbackView displays the door feedback dialog.
type FeedbackView struct {
	task        *core.Task
	customInput string
	isCustom    bool
	width       int
}

// NewFeedbackView creates a new feedback view for the given task.
func NewFeedbackView(task *core.Task) *FeedbackView {
	return &FeedbackView{task: task}
}

// SetWidth sets the terminal width.
func (fv *FeedbackView) SetWidth(w int) {
	fv.width = w
}

// Update handles key input for feedback selection.
func (fv *FeedbackView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if fv.isCustom {
			return fv.handleCustomInput(msg)
		}
		return fv.handleFeedbackSelection(msg)
	}
	return nil
}

func (fv *FeedbackView) handleFeedbackSelection(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "1":
		return feedbackCmd(fv.task, "blocked", "")
	case "2":
		return feedbackCmd(fv.task, "not-now", "")
	case "3":
		return feedbackCmd(fv.task, "needs-breakdown", "")
	case "4":
		fv.isCustom = true
		fv.customInput = ""
	}
	return nil
}

func (fv *FeedbackView) handleCustomInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		if fv.customInput != "" {
			return feedbackCmd(fv.task, "other", fv.customInput)
		}
	case "esc":
		fv.isCustom = false
		fv.customInput = ""
	case "backspace":
		if len(fv.customInput) > 0 {
			fv.customInput = fv.customInput[:len(fv.customInput)-1]
		}
	default:
		if len(msg.String()) == 1 && len(fv.customInput) < 200 {
			fv.customInput += msg.String()
		}
	}
	return nil
}

func feedbackCmd(task *core.Task, feedbackType, comment string) tea.Cmd {
	return func() tea.Msg {
		return DoorFeedbackMsg{Task: task, FeedbackType: feedbackType, Comment: comment}
	}
}

// View renders the feedback dialog.
func (fv *FeedbackView) View() string {
	s := strings.Builder{}
	s.WriteString(feedbackHeaderStyle.Render("Door Feedback"))
	s.WriteString("\n\n")

	s.WriteString(helpStyle.Render(fmt.Sprintf("Task: %s", fv.task.Text)))
	s.WriteString("\n\n")

	s.WriteString("Why isn't this task suitable right now?\n\n")

	for i, option := range feedbackOptions {
		fmt.Fprintf(&s, "  %d. %s\n", i+1, option)
	}

	s.WriteString("\n")

	if fv.isCustom {
		s.WriteString("Enter your comment: " + fv.customInput + "_\n")
	} else {
		s.WriteString(helpStyle.Render("Press 1-4 to select | Esc to cancel"))
	}

	w := fv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}
