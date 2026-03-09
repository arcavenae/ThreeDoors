package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestDetailView(text string) *DetailView {
	task := core.NewTask(text)
	return NewDetailView(task, nil, nil, nil)
}

func newTestDetailViewWithTracker(text string) (*DetailView, *core.SessionTracker) {
	task := core.NewTask(text)
	tracker := core.NewSessionTracker()
	return NewDetailView(task, tracker, nil, nil), tracker
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

func TestDetailView_RendersContext(t *testing.T) {
	task := core.NewTaskWithContext("Buy groceries", "Need healthy food for the week")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "Why:") {
		t.Error("DetailView should show 'Why:' label when context is present")
	}
	if !strings.Contains(view, "Need healthy food for the week") {
		t.Error("DetailView should show the context text")
	}
}

func TestDetailView_NoContext_DoesNotShowWhy(t *testing.T) {
	task := core.NewTask("Simple task")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)
	view := dv.View()
	if strings.Contains(view, "Why:") {
		t.Error("DetailView should NOT show 'Why:' when context is empty")
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
	} else if tcm.Task.Status != core.StatusComplete {
		t.Errorf("expected status %q, got %q", core.StatusComplete, tcm.Task.Status)
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
	} else if tum.Task.Status != core.StatusInProgress {
		t.Errorf("expected status %q, got %q", core.StatusInProgress, tum.Task.Status)
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

// --- Expand ('E' key) ---

func TestDetailView_EKey_EntersExpandInputMode(t *testing.T) {
	dv := newTestDetailView("test task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd != nil {
		t.Error("'e' should not return a command (transitions to expand input mode)")
	}
	if dv.mode != DetailModeExpandInput {
		t.Errorf("expected DetailModeExpandInput, got %d", dv.mode)
	}
}

func TestDetailView_ExpandInput_EnterWithTextSendsExpandMsg(t *testing.T) {
	dv := newTestDetailView("parent task")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})

	// Type subtask text
	for _, r := range "subtask" {
		dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter with text should return a command")
	}
	msg := cmd()
	em, ok := msg.(ExpandTaskMsg)
	if !ok {
		t.Fatalf("expected ExpandTaskMsg, got %T", msg)
	}
	if em.NewTaskText != "subtask" {
		t.Errorf("expected new task text 'subtask', got %q", em.NewTaskText)
	}
	if em.ParentTask.Text != "parent task" {
		t.Errorf("expected parent task 'parent task', got %q", em.ParentTask.Text)
	}
}

func TestDetailView_ExpandInput_EnterEmptyShowsFlash(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter with empty text should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "cannot be empty") {
		t.Errorf("expected empty warning, got %q", fm.Text)
	}
	if dv.mode != DetailModeExpandInput {
		t.Error("should stay in expand input mode after empty submit")
	}
}

func TestDetailView_ExpandInput_EscCancels(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})

	dv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if dv.mode != DetailModeView {
		t.Errorf("Esc should return to DetailModeView, got %d", dv.mode)
	}
	if dv.expandInput != "" {
		t.Errorf("expand input should be cleared after cancel, got %q", dv.expandInput)
	}
}

func TestDetailView_ExpandInput_BackspaceWorks(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	dv.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if dv.expandInput != "a" {
		t.Errorf("expected 'a' after backspace, got %q", dv.expandInput)
	}
}

func TestDetailView_ExpandMode_ShowsInputPrompt(t *testing.T) {
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	view := dv.View()
	if !strings.Contains(view, "subtask") || !strings.Contains(view, "Enter") {
		t.Error("View should show expand input prompt when in expand mode")
	}
}

// --- Fork ('F' key) ---

func TestDetailView_FKey_SendsTaskAddedMsg(t *testing.T) {
	dv := newTestDetailView("original task")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if cmd == nil {
		t.Fatal("'f' should return a command")
	}
	msg := cmd()
	tam, ok := msg.(TaskAddedMsg)
	if !ok {
		t.Fatalf("expected TaskAddedMsg, got %T", msg)
	}
	if tam.Task.Text != "original task" {
		t.Errorf("forked task should have same text, got %q", tam.Task.Text)
	}
	if tam.Task.ID == dv.task.ID {
		t.Error("forked task should have a different ID")
	}
	if tam.Task.Status != core.StatusTodo {
		t.Errorf("forked task should have todo status, got %q", tam.Task.Status)
	}
}

// --- All Status Keys Table-Driven ---

func TestDetailView_AllStatusKeys(t *testing.T) {
	tests := []struct {
		key           string
		expectMsgType string
		expectNoCmd   bool // for keys that transition to input modes
	}{
		{"c", "TaskCompletedMsg", false},
		{"i", "TaskUpdatedMsg", false},
		{"p", "ReturnToDoorsMsg", false},
		{"r", "ReturnToDoorsMsg", false},
		{"m", "ShowMoodMsg", false},
		{"e", "", true}, // transitions to expand input mode
		{"f", "TaskAddedMsg", false},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			dv := newTestDetailView("test task")
			cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if tt.expectNoCmd {
				if cmd != nil {
					t.Fatalf("key %q should not return a command", tt.key)
				}
				return
			}
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
			case TaskAddedMsg:
				msgType = "TaskAddedMsg"
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
	if dv.task.Status != core.StatusBlocked {
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
	if dv.task.Status != core.StatusTodo {
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

// --- Undo Complete ('U' key) ---

func TestDetailView_UKey_UndoComplete_SendsTaskUndoneMsg(t *testing.T) {
	task := core.NewTask("test task")
	_ = task.UpdateStatus(core.StatusComplete)
	tracker := core.NewSessionTracker()
	dv := NewDetailView(task, tracker, nil, nil)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if cmd == nil {
		t.Fatal("'u' on completed task should return a command")
	}
	msg := cmd()
	tum, ok := msg.(TaskUndoneMsg)
	if !ok {
		t.Fatalf("expected TaskUndoneMsg, got %T", msg)
	}
	if tum.Task.Status != core.StatusTodo {
		t.Errorf("expected status %q, got %q", core.StatusTodo, tum.Task.Status)
	}
}

func TestDetailView_UKey_UndoComplete_RecordsUndoEvent(t *testing.T) {
	task := core.NewTask("test task")
	_ = task.UpdateStatus(core.StatusComplete)
	tracker := core.NewSessionTracker()
	dv := NewDetailView(task, tracker, nil, nil)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if cmd != nil {
		cmd()
	}

	metrics := tracker.Finalize()
	if metrics.StatusChanges != 1 {
		t.Errorf("expected 1 status change, got %d", metrics.StatusChanges)
	}
	if metrics.UndoCompleteCount != 1 {
		t.Errorf("expected 1 undo complete, got %d", metrics.UndoCompleteCount)
	}
	if len(metrics.UndoCompletes) != 1 {
		t.Fatalf("expected 1 undo entry, got %d", len(metrics.UndoCompletes))
	}
	if metrics.UndoCompletes[0].TaskID != task.ID {
		t.Errorf("expected task ID %q, got %q", task.ID, metrics.UndoCompletes[0].TaskID)
	}
}

func TestDetailView_UKey_NonComplete_NoOp(t *testing.T) {
	task := core.NewTask("test task") // status = todo
	dv := NewDetailView(task, nil, nil, nil)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if cmd != nil {
		t.Error("'u' on non-completed task should be no-op")
	}
	if task.Status != core.StatusTodo {
		t.Errorf("status should remain todo, got %q", task.Status)
	}
}

func TestDetailView_UKey_DoesNotSendTaskCompletedMsg(t *testing.T) {
	task := core.NewTask("test task")
	_ = task.UpdateStatus(core.StatusComplete)
	dv := NewDetailView(task, nil, nil, nil)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if cmd == nil {
		t.Fatal("expected command from undo")
	}
	msg := cmd()
	if _, ok := msg.(TaskCompletedMsg); ok {
		t.Error("undo should NOT send TaskCompletedMsg (which would trigger MarkComplete)")
	}
	if _, ok := msg.(TaskUpdatedMsg); ok {
		t.Error("undo should send TaskUndoneMsg, not TaskUpdatedMsg")
	}
}

func TestDetailView_CompletedTask_ShowsUndoHint(t *testing.T) {
	task := core.NewTask("test task")
	_ = task.UpdateStatus(core.StatusComplete)
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)

	view := dv.View()
	if !strings.Contains(view, "[U]ndo") {
		t.Error("completed task should show [U]ndo hint")
	}
}

func TestDetailView_TodoTask_NoUndoHint(t *testing.T) {
	task := core.NewTask("test task")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)

	view := dv.View()
	if strings.Contains(view, "[U]ndo") {
		t.Error("todo task should NOT show [U]ndo hint")
	}
}

// --- Space/Enter Toggle (Story 36.4) ---

func TestDetailView_SpaceEnterToggle(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"space closes door", tea.KeyMsg{Type: tea.KeySpace}},
		{"enter closes door", tea.KeyMsg{Type: tea.KeyEnter}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dv := newTestDetailView("test task")
			cmd := dv.Update(tt.key)
			if cmd == nil {
				t.Fatal("expected a command to return to doors")
			}
			msg := cmd()
			if _, ok := msg.(ReturnToDoorsMsg); !ok {
				t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
			}
		})
	}
}

func TestDetailView_SpaceEnterToggle_TextInputGuard(t *testing.T) {
	tests := []struct {
		name    string
		mode    DetailViewMode
		setupFn func(dv *DetailView)
		key     tea.KeyMsg
	}{
		{
			"space in blocker input not intercepted",
			DetailModeBlockerInput,
			func(dv *DetailView) {
				dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
			},
			tea.KeyMsg{Type: tea.KeySpace},
		},
		{
			"enter in expand input not intercepted as toggle",
			DetailModeExpandInput,
			func(dv *DetailView) {
				dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
			},
			tea.KeyMsg{Type: tea.KeyEnter},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dv := newTestDetailView("test task")
			tt.setupFn(dv)
			if dv.mode != tt.mode {
				t.Fatalf("expected mode %d, got %d", tt.mode, dv.mode)
			}
			cmd := dv.Update(tt.key)
			// In text input modes, space/enter should NOT return ReturnToDoorsMsg
			if cmd != nil {
				msg := cmd()
				if _, ok := msg.(ReturnToDoorsMsg); ok {
					t.Error("space/enter in text input mode should NOT return to doors")
				}
			}
		})
	}
}

func TestDetailView_RapidToggle(t *testing.T) {
	// Simulate opening a door, then immediately pressing space to close
	dv := newTestDetailView("test task")

	// First space should close (return to doors)
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeySpace})
	if cmd == nil {
		t.Fatal("first space should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}

	// Create a fresh detail view (simulating re-opening)
	dv2 := newTestDetailView("test task")
	cmd = dv2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should return a command")
	}
	msg = cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestDetailView_EscapeStillWorks(t *testing.T) {
	// AC-3: Escape still works after adding space/enter toggle
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

// --- Invalid Transition ---

func TestDetailView_IKey_InvalidTransition_ShowsError(t *testing.T) {
	task := core.NewTask("test task")
	// Set to in-review (which cannot go to in-progress via 'i' directly? Actually it can.)
	// Let's complete the task first. Then try 'i' which should fail.
	_ = task.UpdateStatus(core.StatusComplete)
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
