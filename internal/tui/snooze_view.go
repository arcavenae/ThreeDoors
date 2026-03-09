package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// SnoozeOption represents one of the snooze duration choices.
type SnoozeOption int

const (
	SnoozeTomorrow SnoozeOption = iota
	SnoozeNextWeek
	SnoozePickDate
	SnoozeSomeday
)

// snoozeOptionLabels maps each option to its display label.
var snoozeOptionLabels = [4]string{
	"Tomorrow",
	"Next Week",
	"Pick Date",
	"Someday",
}

// SnoozeView displays snooze duration options for a task.
type SnoozeView struct {
	task      *core.Task
	cursor    int
	inputMode bool
	dateInput string
	errMsg    string
	width     int
}

// NewSnoozeView creates a snooze view for the given task.
func NewSnoozeView(task *core.Task) *SnoozeView {
	return &SnoozeView{
		task: task,
	}
}

// SetWidth sets the terminal width for rendering.
func (v *SnoozeView) SetWidth(w int) {
	v.width = w
}

// Update handles key input for the snooze view.
func (v *SnoozeView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if v.inputMode {
			return v.handleDateInput(msg)
		}
		return v.handleMenuKeys(msg)
	}
	return nil
}

func (v *SnoozeView) handleMenuKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < 3 {
			v.cursor++
		}
	case "esc":
		return func() tea.Msg { return SnoozeCancelledMsg{} }
	case "enter":
		return v.selectOption()
	}
	return nil
}

func (v *SnoozeView) selectOption() tea.Cmd {
	switch SnoozeOption(v.cursor) {
	case SnoozeTomorrow:
		t := tomorrow9am()
		return v.snoozeCmd(&t, "tomorrow")
	case SnoozeNextWeek:
		t := nextMonday9am()
		return v.snoozeCmd(&t, "next_week")
	case SnoozePickDate:
		v.inputMode = true
		v.dateInput = ""
		v.errMsg = ""
		return nil
	case SnoozeSomeday:
		return v.snoozeCmd(nil, "someday")
	}
	return nil
}

func (v *SnoozeView) snoozeCmd(deferDate *time.Time, option string) tea.Cmd {
	task := v.task
	return func() tea.Msg {
		return TaskSnoozedMsg{
			Task:      task,
			DeferDate: deferDate,
			Option:    option,
		}
	}
}

func (v *SnoozeView) handleDateInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		v.inputMode = false
		v.dateInput = ""
		v.errMsg = ""
	case "enter":
		t, err := parsePickDate(v.dateInput)
		if err != nil {
			v.errMsg = err.Error()
			return nil
		}
		return v.snoozeCmd(&t, "pick_date")
	case "backspace":
		if len(v.dateInput) > 0 {
			v.dateInput = v.dateInput[:len(v.dateInput)-1]
		}
		v.errMsg = ""
	default:
		if len(msg.String()) == 1 && len(v.dateInput) < 10 {
			v.dateInput += msg.String()
			v.errMsg = ""
		}
	}
	return nil
}

// View renders the snooze view.
func (v *SnoozeView) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("SNOOZE TASK"))
	s.WriteString("\n\n")

	taskText := v.task.Text
	if len(taskText) > 60 {
		taskText = taskText[:57] + "..."
	}
	fmt.Fprintf(&s, "  %s\n\n", taskText)

	if v.inputMode {
		s.WriteString("  Enter date (YYYY-MM-DD):\n")
		fmt.Fprintf(&s, "  > %s_\n", v.dateInput)
		if v.errMsg != "" {
			fmt.Fprintf(&s, "  %s\n", v.errMsg)
		}
		s.WriteString("\n  [Enter] Confirm [Esc] Back\n")
	} else {
		for i, label := range snoozeOptionLabels {
			prefix := "  "
			if i == v.cursor {
				prefix = "> "
			}
			fmt.Fprintf(&s, "  %s%s\n", prefix, label)
		}
		s.WriteString("\n  [Enter] Select [Esc] Cancel\n")
	}

	return s.String()
}

// tomorrow9am returns tomorrow at 09:00 local time, stored as UTC.
func tomorrow9am() time.Time {
	now := time.Now().Local()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())
	return tomorrow.UTC()
}

// nextMonday9am returns next Monday at 09:00 local time, stored as UTC.
func nextMonday9am() time.Time {
	now := time.Now().Local()
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 9, 0, 0, 0, now.Location())
	return monday.UTC()
}

// parsePickDate parses a YYYY-MM-DD string and returns that date at 09:00 local time in UTC.
func parsePickDate(input string) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", input, time.Now().Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	at9 := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 9, 0, 0, 0, parsed.Location())
	return at9.UTC(), nil
}
