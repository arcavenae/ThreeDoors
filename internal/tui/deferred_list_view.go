package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// ShowDeferredListMsg is sent to open the deferred list view.
type ShowDeferredListMsg struct{}

// UnsnoozeTaskMsg is sent when a task is un-snoozed from the deferred list.
type UnsnoozeTaskMsg struct {
	Task *core.Task
}

// EditDeferDateMsg is sent when the user wants to edit a task's defer date.
type EditDeferDateMsg struct {
	Task *core.Task
}

// DeferredListView displays all deferred tasks sorted by return date.
type DeferredListView struct {
	tasks  []*core.Task
	pool   *core.TaskPool
	cursor int
	width  int
}

// NewDeferredListView creates a new DeferredListView with deferred tasks from the pool.
func NewDeferredListView(pool *core.TaskPool) *DeferredListView {
	tasks := pool.GetTasksByStatus(core.StatusDeferred)
	sortDeferredTasks(tasks)
	return &DeferredListView{
		tasks: tasks,
		pool:  pool,
	}
}

// SetWidth sets the terminal width for rendering.
func (dv *DeferredListView) SetWidth(w int) {
	dv.width = w
}

// sortDeferredTasks sorts tasks by DeferUntil ascending, nil (Someday) last.
func sortDeferredTasks(tasks []*core.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i].DeferUntil, tasks[j].DeferUntil
		if a == nil && b == nil {
			return false
		}
		if a == nil {
			return false // nil goes last
		}
		if b == nil {
			return true // non-nil before nil
		}
		return a.Before(*b)
	})
}

// formatTimeRemaining returns a human-readable time remaining string.
func formatTimeRemaining(deferUntil *time.Time) string {
	if deferUntil == nil {
		return "Someday"
	}
	now := time.Now().UTC()
	if deferUntil.Before(now) {
		return "Overdue"
	}
	diff := deferUntil.Sub(now)
	hours := int(diff.Hours())
	if hours < 24 {
		return "Tomorrow"
	}
	days := hours / 24
	if days == 1 {
		return "Tomorrow"
	}
	if days < 7 {
		return fmt.Sprintf("%d days", days)
	}
	weeks := days / 7
	if weeks == 1 {
		return "1 week"
	}
	if weeks < 5 {
		return fmt.Sprintf("%d weeks", weeks)
	}
	months := days / 30
	if months == 1 {
		return "1 month"
	}
	return fmt.Sprintf("%d months", months)
}

// formatReturnDate returns the formatted return date string.
func formatReturnDate(deferUntil *time.Time) string {
	if deferUntil == nil {
		return "Someday"
	}
	return deferUntil.Local().Format("Jan 2, 2006")
}

// Refresh reloads the deferred tasks from the pool.
func (dv *DeferredListView) Refresh() {
	dv.tasks = dv.pool.GetTasksByStatus(core.StatusDeferred)
	sortDeferredTasks(dv.tasks)
	if dv.cursor >= len(dv.tasks) {
		dv.cursor = max(0, len(dv.tasks)-1)
	}
}

// Update handles messages for the deferred list view.
func (dv *DeferredListView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		case "j", "down":
			if dv.cursor < len(dv.tasks)-1 {
				dv.cursor++
			}
		case "k", "up":
			if dv.cursor > 0 {
				dv.cursor--
			}
		case "u":
			if len(dv.tasks) > 0 && dv.cursor < len(dv.tasks) {
				task := dv.tasks[dv.cursor]
				return func() tea.Msg { return UnsnoozeTaskMsg{Task: task} }
			}
		case "e":
			if len(dv.tasks) > 0 && dv.cursor < len(dv.tasks) {
				task := dv.tasks[dv.cursor]
				return func() tea.Msg { return EditDeferDateMsg{Task: task} }
			}
		}
	}
	return nil
}

// View renders the deferred list view.
func (dv *DeferredListView) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("Snoozed Tasks"))
	s.WriteString("\n\n")

	if len(dv.tasks) == 0 {
		s.WriteString(helpStyle.Render("No snoozed tasks. Use Z on a door to snooze a task."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Esc return"))
		return s.String()
	}

	maxTextWidth := dv.width - 40
	if maxTextWidth < 20 {
		maxTextWidth = 20
	}

	for i, task := range dv.tasks {
		text := task.Text
		if len(text) > maxTextWidth {
			text = text[:maxTextWidth-3] + "..."
		}

		returnDate := formatReturnDate(task.DeferUntil)
		timeRemaining := formatTimeRemaining(task.DeferUntil)

		line := fmt.Sprintf("  %s  (%s — %s)", text, returnDate, timeRemaining)
		if i == dv.cursor {
			line = searchSelectedStyle.Render(fmt.Sprintf("▸ %s  (%s — %s)", text, returnDate, timeRemaining))
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	fmt.Fprintf(&s, "\n  %d snoozed task(s)\n\n", len(dv.tasks))
	s.WriteString(helpStyle.Render("j/k navigate | u un-snooze | e edit date | Esc return"))

	return s.String()
}
