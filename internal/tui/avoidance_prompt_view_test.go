package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestAvoidanceView(t *testing.T) *AvoidancePromptView {
	t.Helper()
	task := core.NewTask("Avoided task")
	return NewAvoidancePromptView(task, 15)
}

func TestNewAvoidancePromptView(t *testing.T) {
	t.Parallel()
	task := core.NewTask("Test")
	v := NewAvoidancePromptView(task, 10)
	if v.task != task {
		t.Error("expected task to be set")
	}
	if v.count != 10 {
		t.Errorf("expected count 10, got %d", v.count)
	}
}

func TestAvoidancePromptView_SetWidth(t *testing.T) {
	t.Parallel()
	v := newTestAvoidanceView(t)
	v.SetWidth(100)
	if v.width != 100 {
		t.Errorf("expected width 100, got %d", v.width)
	}
}

func TestAvoidancePromptView_Update(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		key        string
		wantAction string
	}{
		{"reconsider lowercase", "r", "reconsider"},
		{"reconsider uppercase", "R", "reconsider"},
		{"breakdown lowercase", "b", "breakdown"},
		{"breakdown uppercase", "B", "breakdown"},
		{"defer lowercase", "d", "defer"},
		{"defer uppercase", "D", "defer"},
		{"archive lowercase", "a", "archive"},
		{"archive uppercase", "A", "archive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := newTestAvoidanceView(t)

			cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if cmd == nil {
				t.Fatal("expected non-nil cmd")
				return
			}

			msg := cmd()
			actionMsg, ok := msg.(AvoidanceActionMsg)
			if !ok {
				t.Fatalf("expected AvoidanceActionMsg, got %T", msg)
			}
			if actionMsg.Action != tt.wantAction {
				t.Errorf("expected action %q, got %q", tt.wantAction, actionMsg.Action)
			}
			if actionMsg.Task == nil {
				t.Error("expected non-nil task")
			}
		})
	}
}

func TestAvoidancePromptView_Update_Escape(t *testing.T) {
	t.Parallel()
	v := newTestAvoidanceView(t)

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected non-nil cmd from Escape")
		return
	}

	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestAvoidancePromptView_Update_UnhandledKey(t *testing.T) {
	t.Parallel()
	v := newTestAvoidanceView(t)

	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd != nil {
		t.Error("expected nil cmd for unhandled key")
	}
}

func TestAvoidancePromptView_View(t *testing.T) {
	t.Parallel()
	v := newTestAvoidanceView(t)

	output := v.View()
	if !strings.Contains(output, "Avoided task") {
		t.Error("expected task text in view")
	}
	if !strings.Contains(output, "15 times") {
		t.Error("expected bypass count in view")
	}
	if !strings.Contains(output, "[R]") {
		t.Error("expected [R] option")
	}
	if !strings.Contains(output, "[B]") {
		t.Error("expected [B] option")
	}
	if !strings.Contains(output, "[D]") {
		t.Error("expected [D] option")
	}
	if !strings.Contains(output, "[A]") {
		t.Error("expected [A] option")
	}
}

func TestAvoidancePromptView_View_LongTaskText(t *testing.T) {
	t.Parallel()
	task := core.NewTask(strings.Repeat("x", 100))
	v := NewAvoidancePromptView(task, 5)

	output := v.View()
	// Long text should be truncated to 60 chars with "..."
	if !strings.Contains(output, "...") {
		t.Error("expected truncated text with '...' for long task")
	}
}
