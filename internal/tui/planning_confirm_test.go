package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmView_Init(t *testing.T) {
	t.Parallel()
	cv := NewConfirmView(nil, ReviewCompleteMsg{}, core.EnergyHigh, false, time.Now().UTC())
	cmd := cv.Init()
	if cmd == nil {
		t.Fatal("Init should return a command for tick and nudge timers")
		return
	}
}

func TestConfirmView_EnterConfirms(t *testing.T) {
	t.Parallel()
	tasks := []*core.Task{core.NewTask("task1 +focus")}
	cv := NewConfirmView(tasks, ReviewCompleteMsg{Reviewed: 1, Continued: 1}, core.EnergyHigh, false, time.Now().UTC())
	cv.SetWidth(80)

	cmd := cv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a command")
		return
	}

	msg := cmd()
	ccm, ok := msg.(ConfirmCompleteMsg)
	if !ok {
		t.Fatalf("expected ConfirmCompleteMsg, got %T", msg)
	}
	if len(ccm.FocusTasks) != 1 {
		t.Errorf("expected 1 focus task, got %d", len(ccm.FocusTasks))
	}
	if ccm.EnergyLevel != core.EnergyHigh {
		t.Errorf("expected EnergyHigh, got %s", ccm.EnergyLevel)
	}
}

func TestConfirmView_EscCancels(t *testing.T) {
	t.Parallel()
	cv := NewConfirmView(nil, ReviewCompleteMsg{}, core.EnergyMedium, true, time.Now().UTC())
	cv.SetWidth(80)

	cmd := cv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Esc should return a command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ConfirmCancelMsg); !ok {
		t.Fatalf("expected ConfirmCancelMsg, got %T", msg)
	}
}

func TestConfirmView_NudgeAfterTimeout(t *testing.T) {
	t.Parallel()
	cv := NewConfirmView(nil, ReviewCompleteMsg{}, core.EnergyLow, false, time.Now().UTC())
	cv.SetWidth(80)

	if cv.nudgeShown {
		t.Fatal("nudge should not be shown initially")
	}

	cv.Update(confirmNudgeMsg{})
	if !cv.nudgeShown {
		t.Fatal("nudge should be shown after nudge message")
	}
}

func TestConfirmView_TickUpdatesElapsed(t *testing.T) {
	t.Parallel()
	start := time.Now().UTC().Add(-5 * time.Second)
	cv := NewConfirmView(nil, ReviewCompleteMsg{}, core.EnergyHigh, false, start)
	cv.SetWidth(80)

	cmd := cv.Update(confirmTickMsg(time.Now().UTC()))
	if cmd == nil {
		t.Fatal("tick should return another tick command")
		return
	}
	if cv.elapsed < 4*time.Second {
		t.Errorf("elapsed should be at least 4s, got %v", cv.elapsed)
	}
}

func TestConfirmView_ViewRendersContent(t *testing.T) {
	t.Parallel()
	tasks := []*core.Task{
		core.NewTask("Write tests +focus"),
		core.NewTask("Review PR +focus"),
	}
	metrics := ReviewCompleteMsg{Reviewed: 5, Continued: 3, Deferred: 1, Dropped: 1}
	cv := NewConfirmView(tasks, metrics, core.EnergyHigh, true, time.Now().UTC())
	cv.SetWidth(80)
	cv.SetHeight(40)

	view := cv.View()

	tests := []struct {
		name     string
		expected string
	}{
		{"header", "Confirm Focus"},
		{"focus task 1", "Write tests"},
		{"focus task 2", "Review PR"},
		{"reviewed count", "Tasks reviewed: 5"},
		{"continued", "Continued: 3"},
		{"deferred", "Deferred:  1"},
		{"dropped", "Dropped:   1"},
		{"energy", "Energy:"},
		{"step indicator", "Step 3/3"},
		{"enter key", "Enter"},
		{"esc key", "Esc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(view, tt.expected) {
				t.Errorf("view should contain %q", tt.expected)
			}
		})
	}
}

func TestConfirmView_ViewNoFocusTasks(t *testing.T) {
	t.Parallel()
	cv := NewConfirmView(nil, ReviewCompleteMsg{}, core.EnergyLow, false, time.Now().UTC())
	cv.SetWidth(80)

	view := cv.View()
	if !strings.Contains(view, "No focus tasks selected") {
		t.Error("should show 'No focus tasks selected' when no tasks")
	}
}

func TestConfirmView_ViewNudgeMessage(t *testing.T) {
	t.Parallel()
	cv := NewConfirmView(nil, ReviewCompleteMsg{}, core.EnergyLow, false, time.Now().UTC())
	cv.SetWidth(80)
	cv.nudgeShown = true

	view := cv.View()
	if !strings.Contains(view, "taking a while") {
		t.Error("should show nudge message when nudgeShown is true")
	}
}
