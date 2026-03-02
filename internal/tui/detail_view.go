package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailViewMode tracks the sub-view within the detail view.
type DetailViewMode int

const (
	DetailModeView DetailViewMode = iota
	DetailModeBlockerInput
)

// DetailView displays full task details and status action menu.
type DetailView struct {
	task         *tasks.Task
	mode         DetailViewMode
	blockerInput string
	width        int
	tracker      *tasks.SessionTracker
}

// NewDetailView creates a detail view for the given task.
func NewDetailView(task *tasks.Task, tracker *tasks.SessionTracker) *DetailView {
	if tracker != nil {
		tracker.RecordDetailView()
	}
	return &DetailView{
		task:    task,
		mode:    DetailModeView,
		tracker: tracker,
	}
}

// SetWidth sets the terminal width.
func (dv *DetailView) SetWidth(w int) {
	dv.width = w
}

// Update handles key input in the detail view.
func (dv *DetailView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if dv.mode == DetailModeBlockerInput {
			return dv.handleBlockerInput(msg)
		}
		return dv.handleDetailKeys(msg)
	}
	return nil
}

func (dv *DetailView) handleDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "c", "C":
		if err := dv.task.UpdateStatus(tasks.StatusComplete); err != nil {
			return func() tea.Msg { return FlashMsg{Text: "Cannot complete: " + err.Error()} }
		}
		if dv.tracker != nil {
			dv.tracker.RecordStatusChange()
			dv.tracker.RecordTaskCompleted()
		}
		return func() tea.Msg { return TaskCompletedMsg{Task: dv.task} }
	case "b", "B":
		if tasks.IsValidTransition(dv.task.Status, tasks.StatusBlocked) {
			dv.mode = DetailModeBlockerInput
			dv.blockerInput = ""
		}
	case "i", "I":
		if err := dv.task.UpdateStatus(tasks.StatusInProgress); err != nil {
			return func() tea.Msg { return FlashMsg{Text: "Cannot change status: " + err.Error()} }
		}
		if dv.tracker != nil {
			dv.tracker.RecordStatusChange()
		}
		return func() tea.Msg { return TaskUpdatedMsg{Task: dv.task} }
	case "e", "E":
		// Expand: not yet implemented (deferred UX)
		return func() tea.Msg { return FlashMsg{Text: "Expand not yet implemented"} }
	case "f", "F":
		// Fork: not yet implemented (deferred UX)
		return func() tea.Msg { return FlashMsg{Text: "Fork not yet implemented"} }
	case "p", "P":
		// Procrastinate: just return to doors (task stays in pool)
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "r", "R":
		// Rework: keep in pool, just return
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "m", "M":
		return func() tea.Msg { return ShowMoodMsg{} }
	}
	return nil
}

func (dv *DetailView) handleBlockerInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		if err := dv.task.UpdateStatus(tasks.StatusBlocked); err == nil {
			if dv.blockerInput != "" {
				_ = dv.task.SetBlocker(dv.blockerInput)
			}
			if dv.tracker != nil {
				dv.tracker.RecordStatusChange()
			}
			dv.mode = DetailModeView
			return func() tea.Msg { return TaskUpdatedMsg{Task: dv.task} }
		}
		dv.mode = DetailModeView
	case "esc":
		dv.mode = DetailModeView
		dv.blockerInput = ""
	case "backspace":
		if len(dv.blockerInput) > 0 {
			dv.blockerInput = dv.blockerInput[:len(dv.blockerInput)-1]
		}
	default:
		if len(msg.String()) == 1 && len(dv.blockerInput) < 200 {
			dv.blockerInput += msg.String()
		}
	}
	return nil
}

// View renders the detail view.
func (dv *DetailView) View() string {
	s := strings.Builder{}

	w := dv.width - 6
	if w < 40 {
		w = 40
	}

	s.WriteString(headerStyle.Render("TASK DETAILS"))
	s.WriteString("\n\n")

	statusColor := StatusColor(string(dv.task.Status))
	statusStyle := lipgloss.NewStyle().Foreground(statusColor).Bold(true)
	fmt.Fprintf(&s, "Status: %s\n", statusStyle.Render(string(dv.task.Status)))
	fmt.Fprintf(&s, "Created: %s\n", dv.task.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(&s, "Updated: %s\n", dv.task.UpdatedAt.Format("2006-01-02 15:04"))

	if dv.task.Blocker != "" {
		blockerStyle := lipgloss.NewStyle().Foreground(colorBlocked)
		fmt.Fprintf(&s, "Blocker: %s\n", blockerStyle.Render(dv.task.Blocker))
	}

	s.WriteString("\n")
	s.WriteString(dv.task.Text)
	s.WriteString("\n")

	if dv.task.Context != "" {
		contextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true)
		fmt.Fprintf(&s, "\nWhy: %s\n", contextStyle.Render(dv.task.Context))
	}

	if len(dv.task.Notes) > 0 {
		s.WriteString("\nNotes:\n")
		for _, note := range dv.task.Notes {
			fmt.Fprintf(&s, "  [%s] %s\n", note.Timestamp.Format("15:04"), note.Text)
		}
	}

	s.WriteString("\n")
	s.WriteString(separatorStyle.Render("─────────────────────────────────"))
	s.WriteString("\n\n")

	if dv.mode == DetailModeBlockerInput {
		s.WriteString("Blocker reason (Enter to submit, Esc to cancel):\n")
		s.WriteString("> " + dv.blockerInput + "_\n")
	} else {
		s.WriteString(helpStyle.Render("[C]omplete [B]locked [I]n-progress [E]xpand [F]ork [P]rocrastinate [R]ework [M]ood [Esc]Back"))
	}

	return detailBorder.Width(w).Render(s.String())
}
