package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestDetailView(text string) *DetailView {
	task := tasks.NewTask(text)
	return NewDetailView(task, nil)
}

func newTestDetailViewWithTracker(text string) (*DetailView, *tasks.SessionTracker) {
	task := tasks.NewTask(text)
	tracker := tasks.NewSessionTracker()
	return NewDetailView(task, tracker), tracker
}

// --- View Rendering ---

func TestDetailView_RendersFullTaskText(t *testing.T) {
	dv := newTestDetailView("This is a very long task description that should be displayed in full")
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "This is a very long task description that should be displayed in full") {
		t.Error("DetailView should render full task text (not truncated)")
	}
}

func TestDetailView_RendersStatusMenu(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	view := dv.View()

	expectedKeys := []string{"[C]omplete", "[B]locked", "[I]n-progress", "[Esc]"}
	for _, key := range expectedKeys {
		if !strings.Contains(view, key) {
			t.Errorf("DetailView should contain %q in status menu", key)
		}
	}
}

func TestDetailView_RendersTaskStatus(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "todo") {
		t.Error("DetailView should show task status")
	}
}

func TestDetailView_RendersHeader(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "TASK DETAILS") {
		t.Error("DetailView should render 'TASK DETAILS' header")
	}
}

// --- Key Handling ---

func TestDetailView_EscKey_ReturnsToDoorsMsg(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Esc should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestDetailView_CKey_CompletesTask(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cmd == nil {
		t.Fatal("'c' should return a command")
	}
	msg := cmd()
	if tcm, ok := msg.(TaskCompletedMsg); !ok {
		t.Errorf("expected TaskCompletedMsg, got %T", msg)
	} else if tcm.Task.Status != tasks.StatusComplete {
		t.Errorf("expected status %q, got %q", tasks.StatusComplete, tcm.Task.Status)
	}
}

func TestDetailView_IKey_SetsInProgress(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	if cmd == nil {
		t.Fatal("'i' should return a command")
	}
	msg := cmd()
	if tum, ok := msg.(TaskUpdatedMsg); !ok {
		t.Errorf("expected TaskUpdatedMsg, got %T", msg)
	} else if tum.Task.Status != tasks.StatusInProgress {
		t.Errorf("expected status %q, got %q", tasks.StatusInProgress, tum.Task.Status)
	}
}

func TestDetailView_BKey_TransitionsToBlockerInput(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	// 'b' should transition to blocker input mode (no command returned)
	if cmd != nil {
		t.Error("'b' should not return a command (transitions to blocker input mode)")
	}
	if dv.mode != DetailModeBlockerInput {
		t.Errorf("expected DetailModeBlockerInput, got %d", dv.mode)
	}
}

func TestDetailView_PKey_ReturnsToDoors(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if cmd == nil {
		t.Fatal("'p' should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestDetailView_RKey_ReturnsToDoors(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd == nil {
		t.Fatal("'r' should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestDetailView_MKey_ShowsMoodMsg(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	if cmd == nil {
		t.Fatal("'m' should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ShowMoodMsg); !ok {
		t.Errorf("expected ShowMoodMsg, got %T", msg)
	}
}

func TestDetailView_EKey_FlashesNotImplemented(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd == nil {
		t.Fatal("'e' should return a command")
	}
	msg := cmd()
	if fm, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if !strings.Contains(fm.Text, "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got %q", fm.Text)
	}
}

func TestDetailView_FKey_FlashesNotImplemented(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if cmd == nil {
		t.Fatal("'f' should return a command")
	}
	msg := cmd()
	if fm, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if !strings.Contains(fm.Text, "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got %q", fm.Text)
	}
}

// --- All Status Keys Table-Driven ---

func TestDetailView_AllStatusKeys(t *testing.T) {
	tests := []struct {
		key         string
		expectMsgType string
	}{
		{"c", "TaskCompletedMsg"},
		{"i", "TaskUpdatedMsg"},
		{"p", "ReturnToDoorsMsg"},
		{"r", "ReturnToDoorsMsg"},
		{"m", "ShowMoodMsg"},
		{"e", "FlashMsg"},
		{"f", "FlashMsg"},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			dv := newTestDetailView("test task")
			cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if cmd == nil {
				if tt.key == "b" {
					return // 'b' transitions to blocker mode, no cmd
				}
				t.Fatalf("key %q should return a command", tt.key)
			}
			msg := cmd()
			msgType := ""
			switch msg.(type) {
			case TaskCompletedMsg:
				msgType = "TaskCompletedMsg"
			case TaskUpdatedMsg:
				msgType = "TaskUpdatedMsg"
			case ReturnToDoorsMsg:
				msgType = "ReturnToDoorsMsg"
			case ShowMoodMsg:
				msgType = "ShowMoodMsg"
			case FlashMsg:
				msgType = "FlashMsg"
			default:
				msgType = "unknown"
			}
			if msgType != tt.expectMsgType {
				t.Errorf("key %q: expected %s, got %s", tt.key, tt.expectMsgType, msgType)
			}
		})
	}
}

// --- Blocker Input ---

func TestDetailView_BlockerInput_EnterSubmits(t *testing.T) {
	dv := newTestDetailView("test task")
	// Enter blocker mode
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if dv.mode != DetailModeBlockerInput {
		t.Fatal("should be in blocker input mode")
	}

	// Type blocker text
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})

	// Submit
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter in blocker input should return a command")
	}
	msg := cmd()
	if _, ok := msg.(TaskUpdatedMsg); !ok {
		t.Errorf("expected TaskUpdatedMsg, got %T", msg)
	}
	if dv.task.Status != tasks.StatusBlocked {
		t.Errorf("expected status blocked, got %q", dv.task.Status)
	}
}

func TestDetailView_BlockerInput_EscCancels(t *testing.T) {
	dv := newTestDetailView("test task")
	// Enter blocker mode
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	// Cancel
	dv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if dv.mode != DetailModeView {
		t.Errorf("Esc should return to DetailModeView, got %d", dv.mode)
	}
	if dv.task.Status != tasks.StatusTodo {
		t.Errorf("task status should remain todo after cancel, got %q", dv.task.Status)
	}
}

func TestDetailView_BlockerInput_BackspaceWorks(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	dv.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if dv.blockerInput != "a" {
		t.Errorf("expected 'a' after backspace, got %q", dv.blockerInput)
	}
}

// --- Blocker View Rendering ---

func TestDetailView_BlockerMode_ShowsInputPrompt(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	view := dv.View()
	if !strings.Contains(view, "Blocker reason") {
		t.Error("View should show blocker input prompt when in blocker mode")
	}
}

// --- Tracker Integration ---

func TestDetailView_RecordsDetailView(t *testing.T) {
	_, tracker := newTestDetailViewWithTracker("test task")
	metrics := tracker.Finalize()
	if metrics.DetailViews != 1 {
		t.Errorf("expected 1 detail view recorded, got %d", metrics.DetailViews)
	}
}

func TestDetailView_CKey_RecordsStatusChange(t *testing.T) {
	dv, tracker := newTestDetailViewWithTracker("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cmd != nil {
		cmd()
	}
	metrics := tracker.Finalize()
	if metrics.StatusChanges != 1 {
		t.Errorf("expected 1 status change, got %d", metrics.StatusChanges)
	}
	if metrics.TasksCompleted != 1 {
		t.Errorf("expected 1 task completed, got %d", metrics.TasksCompleted)
	}
}

// --- Invalid Transition ---

func TestDetailView_IKey_InvalidTransition_ShowsError(t *testing.T) {
	task := tasks.NewTask("test task")
	// Set to in-review (which cannot go to in-progress via 'i' directly? Actually it can.)
	// Let's complete the task first. Then try 'i' which should fail.
	_ = task.UpdateStatus(tasks.StatusComplete)
	dv := &DetailView{task: task, mode: DetailModeView}
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	if cmd == nil {
		t.Fatal("expected a flash command for invalid transition")
	}
	msg := cmd()
	if fm, ok := msg.(FlashMsg); !ok {
		t.Errorf("expected FlashMsg, got %T", msg)
	} else if !strings.Contains(fm.Text, "Cannot change status") {
		t.Errorf("expected error message, got %q", fm.Text)
	}
}
