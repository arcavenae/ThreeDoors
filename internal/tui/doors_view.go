package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/charmbracelet/lipgloss"
)

// DoorsView renders the three doors interface.
type DoorsView struct {
	pool              *tasks.TaskPool
	currentDoors      []*tasks.Task
	selectedDoorIndex int
	completedCount    int
	width             int
	tracker           *tasks.SessionTracker
}

// NewDoorsView creates a new DoorsView.
func NewDoorsView(pool *tasks.TaskPool, tracker *tasks.SessionTracker) *DoorsView {
	dv := &DoorsView{
		pool:              pool,
		selectedDoorIndex: -1,
		tracker:           tracker,
	}
	dv.RefreshDoors()
	return dv
}

// RefreshDoors selects new random doors from the pool.
func (dv *DoorsView) RefreshDoors() {
	dv.currentDoors = tasks.SelectDoors(dv.pool, 3)
	dv.selectedDoorIndex = -1
}

// GetCurrentDoorTexts returns the text of currently displayed doors.
func (dv *DoorsView) GetCurrentDoorTexts() []string {
	var texts []string
	for _, t := range dv.currentDoors {
		texts = append(texts, t.Text)
	}
	return texts
}

// IncrementCompleted increments the session completion count.
func (dv *DoorsView) IncrementCompleted() {
	dv.completedCount++
}

// SetWidth sets the terminal width for rendering.
func (dv *DoorsView) SetWidth(w int) {
	dv.width = w
}

// View renders the doors view.
func (dv *DoorsView) View() string {
	s := strings.Builder{}
	s.WriteString(headerStyle.Render("ThreeDoors - Technical Demo"))
	s.WriteString("\n\n")

	if len(dv.currentDoors) == 0 {
		s.WriteString(flashStyle.Render("All tasks done! Great work!"))
		s.WriteString("\n\nPress 'q' to quit.\n")
		return s.String()
	}

	doorWidth := 30
	if dv.width > 20 {
		doorWidth = (dv.width - 6) / 3
		if doorWidth < 15 {
			doorWidth = 15
		}
	}

	var renderedDoors []string
	for i, task := range dv.currentDoors {
		content := task.Text
		statusIndicator := lipgloss.NewStyle().
			Foreground(StatusColor(string(task.Status))).
			Render(fmt.Sprintf("[%s]", task.Status))
		content = statusIndicator + "\n\n" + content

		style := doorStyle.Width(doorWidth)
		if i == dv.selectedDoorIndex {
			style = selectedDoorStyle.Width(doorWidth)
		}
		renderedDoors = append(renderedDoors, style.Render(content))
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedDoors...))

	if dv.completedCount > 0 {
		s.WriteString(fmt.Sprintf("\n\nCompleted this session: %d", dv.completedCount))
	}

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("a/left, w/up, d/right to select | s/down to re-roll | Enter to open | M mood | q quit"))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Progress over perfection. Just pick one and start."))

	return s.String()
}
