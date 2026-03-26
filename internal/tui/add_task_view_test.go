package tui

import (
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// --- AddTaskView Creation Tests ---

func TestAddTaskView_New(t *testing.T) {
	av := NewAddTaskView()
	if av == nil {
		t.Fatal("NewAddTaskView should not return nil")
		return
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
		return
	}

	msg := cmd()
	taskMsg, ok := msg.(TaskAddedMsg)
	if !ok {
		t.Fatalf("expected TaskAddedMsg, got %T", msg)
	}
	if taskMsg.Task.Text != "buy milk" {
		t.Errorf("expected task text 'buy milk', got %q", taskMsg.Task.Text)
	}
	if taskMsg.Task.Status != core.StatusTodo {
		t.Errorf("expected task status todo, got %s", taskMsg.Task.Status)
	}
}

// --- Esc Key: Cancel (T2) ---

func TestAddTaskView_Esc_ReturnsToDoorsMsg(t *testing.T) {
	av := NewAddTaskView()
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected a command from Esc, got nil")
		return
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
		return
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
		return
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
		return
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

// --- Multi-step Context Flow Tests ---

func TestAddTaskWithContextView_New(t *testing.T) {
	av := NewAddTaskWithContextView()
	if av == nil {
		t.Fatal("NewAddTaskWithContextView should not return nil")
		return
	}
	if !av.withContext {
		t.Error("withContext should be true")
	}
	if av.step != stepTaskText {
		t.Errorf("expected initial step to be stepTaskText, got %d", av.step)
	}
}

func TestAddTaskWithContextView_Step1_Enter_AdvancesToStep2(t *testing.T) {
	av := NewAddTaskWithContextView()
	av.SetWidth(80)

	// Type task text
	for _, r := range "Buy groceries" {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Enter → should advance to step 2, not create task
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil command when advancing to step 2")
	}
	if av.step != stepContext {
		t.Errorf("expected step to be stepContext, got %d", av.step)
	}
	if av.capturedText != "Buy groceries" {
		t.Errorf("expected captured text 'Buy groceries', got %q", av.capturedText)
	}
}

func TestAddTaskWithContextView_Step2_Enter_WithContext_CreatesTask(t *testing.T) {
	av := NewAddTaskWithContextView()
	av.SetWidth(80)

	// Step 1: type task text and press Enter
	for _, r := range "Buy groceries" {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	av.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Step 2: type context and press Enter
	for _, r := range "Need healthy food" {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from step 2 Enter")
		return
	}

	msg := cmd()
	taskMsg, ok := msg.(TaskAddedMsg)
	if !ok {
		t.Fatalf("expected TaskAddedMsg, got %T", msg)
	}
	if taskMsg.Task.Text != "Buy groceries" {
		t.Errorf("expected task text 'Buy groceries', got %q", taskMsg.Task.Text)
	}
	if taskMsg.Task.Context != "Need healthy food" {
		t.Errorf("expected context 'Need healthy food', got %q", taskMsg.Task.Context)
	}
}

func TestAddTaskWithContextView_Step2_Enter_EmptyContext_SkipsContext(t *testing.T) {
	av := NewAddTaskWithContextView()
	av.SetWidth(80)

	// Step 1: type task text and press Enter
	for _, r := range "Buy groceries" {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	av.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Step 2: press Enter with empty input (skip context)
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from step 2 Enter")
		return
	}

	msg := cmd()
	taskMsg, ok := msg.(TaskAddedMsg)
	if !ok {
		t.Fatalf("expected TaskAddedMsg, got %T", msg)
	}
	if taskMsg.Task.Text != "Buy groceries" {
		t.Errorf("expected task text 'Buy groceries', got %q", taskMsg.Task.Text)
	}
	if taskMsg.Task.Context != "" {
		t.Errorf("expected empty context, got %q", taskMsg.Task.Context)
	}
}

func TestAddTaskWithContextView_Step1_EmptyText_ShowsError(t *testing.T) {
	av := NewAddTaskWithContextView()
	av.SetWidth(80)

	// Press Enter with no text
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from Enter with empty text")
		return
	}

	msg := cmd()
	flashMsg, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg for empty text, got %T", msg)
	}
	if !strings.Contains(flashMsg.Text, "cannot be empty") {
		t.Errorf("expected error about empty text, got %q", flashMsg.Text)
	}
	if av.step != stepTaskText {
		t.Error("should remain on step 1 after empty text error")
	}
}

func TestAddTaskWithContextView_Esc_FromStep1_Cancels(t *testing.T) {
	av := NewAddTaskWithContextView()
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected a command from Esc")
		return
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Fatalf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestAddTaskWithContextView_Esc_FromStep2_Cancels(t *testing.T) {
	av := NewAddTaskWithContextView()
	av.SetWidth(80)

	// Advance to step 2
	for _, r := range "Some task" {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	av.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Esc from step 2
	cmd := av.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected a command from Esc")
		return
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Fatalf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestAddTaskWithContextView_View_ShowsStepIndicator(t *testing.T) {
	av := NewAddTaskWithContextView()
	av.SetWidth(80)
	view := av.View()
	if !strings.Contains(view, "Step 1") {
		t.Error("view should show 'Step 1' on initial state")
	}
	if !strings.Contains(view, "Context") {
		t.Error("view should mention 'Context' in header")
	}
}

func TestAddTaskWithContextView_View_Step2_ShowsStep2(t *testing.T) {
	av := NewAddTaskWithContextView()
	av.SetWidth(80)

	// Advance to step 2
	for _, r := range "Task text" {
		av.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	av.Update(tea.KeyMsg{Type: tea.KeyEnter})

	view := av.View()
	if !strings.Contains(view, "Step 2") {
		t.Error("view should show 'Step 2' after advancing")
	}
	if !strings.Contains(view, "skip") {
		t.Error("step 2 should mention 'skip' option")
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
