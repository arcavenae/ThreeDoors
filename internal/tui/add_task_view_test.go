package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// --- AddTaskView Creation Tests ---

func TestAddTaskView_New(t *testing.T) {
	av := NewAddTaskView()
	if av == nil {
		t.Fatal("NewAddTaskView should not return nil")
	}
}

func TestAddTaskView_InitialState(t *testing.T) {
	av := NewAddTaskView()
	view := av.View()
	if !strings.Contains(view, "Add Task") {
		t.Error("view should contain 'Add Task' header")
	}
}

func TestAddTaskView_SetWidth(t *testing.T) {
	av := NewAddTaskView()
	av.SetWidth(120)
	// Should not panic; width is stored for rendering
}

// --- Enter Key: Create Task (T1) ---

func TestAddTaskView_Enter_WithText_EmitsTaskAddedMsg(t *testing.T) {
	av := NewAddTaskView()
	av.SetWidth(80)

	// Simulate typing "buy milk"
	for _, r := range "buy milk" {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Enter
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter, got nil")
	}

	msg := cmd()
	taskMsg, ok := msg.(TaskAddedMsg)
	if !ok {
		t.Fatalf("expected TaskAddedMsg, got %T", msg)
	}
	if taskMsg.Task.Text != "buy milk" {
		t.Errorf("expected task text 'buy milk', got %q", taskMsg.Task.Text)
	}
	if taskMsg.Task.Status != tasks.StatusTodo {
		t.Errorf("expected task status todo, got %s", taskMsg.Task.Status)
	}
}

// --- Esc Key: Cancel (T2) ---

func TestAddTaskView_Esc_ReturnsToDoorsMsg(t *testing.T) {
	av := NewAddTaskView()
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected a command from Esc, got nil")
	}

	msg := cmd()
	_, ok := msg.(ReturnToDoorsMsg)
	if !ok {
		t.Fatalf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

// --- Empty Input Validation (T3, T11) ---

func TestAddTaskView_Enter_EmptyText_ShowsError(t *testing.T) {
	av := NewAddTaskView()
	// Press Enter with no text
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter with empty text, got nil")
	}

	msg := cmd()
	flashMsg, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg for empty text, got %T", msg)
	}
	if !strings.Contains(flashMsg.Text, "cannot be empty") {
		t.Errorf("expected error about empty text, got %q", flashMsg.Text)
	}
}

func TestAddTaskView_Enter_WhitespaceOnly_ShowsError(t *testing.T) {
	av := NewAddTaskView()
	// Type only spaces
	for _, r := range "   " {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter with whitespace, got nil")
	}

	msg := cmd()
	flashMsg, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg for whitespace-only text, got %T", msg)
	}
	if !strings.Contains(flashMsg.Text, "cannot be empty") {
		t.Errorf("expected error about empty text, got %q", flashMsg.Text)
	}
}

// --- Character Limit (T4) ---

func TestAddTaskView_CharLimit_500(t *testing.T) {
	av := NewAddTaskView()
	// The textinput CharLimit should be set to 500
	// We verify by checking that the view accepts text up to the limit
	longText := strings.Repeat("a", 500)
	for _, r := range longText {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command for 500-char text")
	}
	msg := cmd()
	taskMsg, ok := msg.(TaskAddedMsg)
	if !ok {
		t.Fatalf("expected TaskAddedMsg, got %T", msg)
	}
	if len(taskMsg.Task.Text) > 500 {
		t.Errorf("task text should not exceed 500 chars, got %d", len(taskMsg.Task.Text))
	}
}

// --- View Rendering ---

func TestAddTaskView_View_ContainsHelpText(t *testing.T) {
	av := NewAddTaskView()
	av.SetWidth(80)
	view := av.View()
	if !strings.Contains(view, "Enter") || !strings.Contains(view, "Esc") {
		t.Error("view should contain help text about Enter and Esc keys")
	}
}
