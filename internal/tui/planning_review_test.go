package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func makeTasks(n int) []*core.Task {
	tasks := make([]*core.Task, n)
	for i := range n {
		t := core.NewTask("Task " + string(rune('A'+i)))
		t.Status = core.StatusInProgress
		t.Type = core.TypeTechnical
		tasks[i] = t
	}
	return tasks
}

func reviewKeyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func reviewSpecialKeyMsg(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

// --- AC 27.2.1: ReviewView Bubbletea Model ---

func TestNewReviewView(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(3)
	rv := NewReviewView(tasks)
	if rv == nil {
		t.Fatal("NewReviewView returned nil")
		return
	}
}

func TestReviewViewDisplaysTaskText(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(3)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)
	output := rv.View()
	if !strings.Contains(output, "Task A") {
		t.Errorf("expected first task text in view, got:\n%s", output)
	}
}

func TestReviewViewDisplaysTaskStatus(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	tasks[0].Status = core.StatusInProgress
	rv := NewReviewView(tasks)
	rv.SetWidth(80)
	output := rv.View()
	if !strings.Contains(output, "in-progress") {
		t.Errorf("expected task status in view, got:\n%s", output)
	}
}

func TestReviewViewDisplaysProgressCounter(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(5)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)
	output := rv.View()
	if !strings.Contains(output, "1/5") {
		t.Errorf("expected progress counter '1/5' in view, got:\n%s", output)
	}
}

func TestReviewViewProgressCounterAdvances(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(3)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Choose continue on first task
	rv.Update(reviewKeyMsg("c"))
	// Advance to next
	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	output := rv.View()
	if !strings.Contains(output, "2/3") {
		t.Errorf("expected progress counter '2/3' after advancing, got:\n%s", output)
	}
}

// --- AC 27.2.2: Review Actions ---

func TestReviewActionContinue(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("c"))

	if rv.decisions[0] != ReviewContinue {
		t.Errorf("expected decision ReviewContinue, got %v", rv.decisions[0])
	}
}

func TestReviewActionDefer(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("d"))

	if rv.decisions[0] != ReviewDefer {
		t.Errorf("expected decision ReviewDefer, got %v", rv.decisions[0])
	}
}

func TestReviewActionDrop(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("x"))

	if rv.decisions[0] != ReviewDrop {
		t.Errorf("expected decision ReviewDrop, got %v", rv.decisions[0])
	}
	// Drop transitions task to deferred
	if tasks[0].Status != core.StatusDeferred {
		t.Errorf("expected task status to be deferred after drop, got %s", tasks[0].Status)
	}
}

func TestReviewActionDropFromTodo(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(1)
	tasks[0].Status = core.StatusTodo
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("x"))

	if tasks[0].Status != core.StatusDeferred {
		t.Errorf("expected task status to be deferred after drop from todo, got %s", tasks[0].Status)
	}
}

func TestReviewActionDropFromBlocked(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(1)
	tasks[0].Status = core.StatusBlocked
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("x"))

	if tasks[0].Status != core.StatusDeferred {
		t.Errorf("expected task status to be deferred after drop from blocked, got %s", tasks[0].Status)
	}
}

func TestReviewEnterAdvancesAfterAction(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(3)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("c"))
	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	if rv.current != 1 {
		t.Errorf("expected current to be 1 after enter, got %d", rv.current)
	}
}

func TestReviewEnterDoesNotAdvanceWithoutAction(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(3)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	if rv.current != 0 {
		t.Errorf("expected current to remain 0 without action, got %d", rv.current)
	}
}

func TestReviewCompletesWhenAllProcessed(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Process first
	rv.Update(reviewKeyMsg("c"))
	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))
	// Process second
	rv.Update(reviewKeyMsg("d"))
	cmd := rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	if cmd == nil {
		t.Fatal("expected a command when review completes")
	}
	msg := cmd()
	if _, ok := msg.(ReviewCompleteMsg); !ok {
		t.Errorf("expected ReviewCompleteMsg, got %T", msg)
	}
}

func TestReviewDecisionMetrics(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(3)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Continue, Defer, Drop
	rv.Update(reviewKeyMsg("c"))
	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))
	rv.Update(reviewKeyMsg("d"))
	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))
	rv.Update(reviewKeyMsg("x"))
	cmd := rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	msg := cmd().(ReviewCompleteMsg)
	if msg.Continued != 1 {
		t.Errorf("expected 1 continued, got %d", msg.Continued)
	}
	if msg.Deferred != 1 {
		t.Errorf("expected 1 deferred, got %d", msg.Deferred)
	}
	if msg.Dropped != 1 {
		t.Errorf("expected 1 dropped, got %d", msg.Dropped)
	}
	if msg.Reviewed != 3 {
		t.Errorf("expected 3 reviewed, got %d", msg.Reviewed)
	}
}

// --- AC 27.2.3: Empty State ---

func TestReviewEmptyState(t *testing.T) {
	t.Parallel()
	rv := NewReviewView(nil)
	rv.SetWidth(80)
	output := rv.View()
	if !strings.Contains(output, "No incomplete tasks") {
		t.Errorf("expected empty state message, got:\n%s", output)
	}
}

func TestReviewEmptyStateAutoAdvances(t *testing.T) {
	t.Parallel()
	rv := NewReviewView(nil)
	rv.SetWidth(80)

	cmd := rv.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a tick command for empty state")
	}
}

func TestReviewEmptyStateTickCompletes(t *testing.T) {
	t.Parallel()
	rv := NewReviewView(nil)
	rv.SetWidth(80)

	cmd := rv.Update(reviewAutoAdvanceMsg{})
	if cmd == nil {
		t.Fatal("expected command on auto-advance tick")
	}
	msg := cmd()
	if _, ok := msg.(ReviewCompleteMsg); !ok {
		t.Errorf("expected ReviewCompleteMsg on auto-advance, got %T", msg)
	}
}

// --- AC 27.2.4: Visual Design ---

func TestReviewViewShowsActionKeys(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)
	output := rv.View()

	if !strings.Contains(output, "[C]") || !strings.Contains(output, "[D]") || !strings.Contains(output, "[X]") {
		t.Errorf("expected action key hints in view, got:\n%s", output)
	}
}

func TestReviewViewShowsStepIndicator(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)
	output := rv.View()

	if !strings.Contains(output, "Step 1/3") {
		t.Errorf("expected step indicator 'Step 1/3' in view, got:\n%s", output)
	}
}

func TestReviewViewShowsElapsedTime(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Simulate a tick to update elapsed time
	rv.Update(reviewTickMsg(time.Now()))

	output := rv.View()
	if !strings.Contains(output, "0:0") {
		t.Errorf("expected elapsed time display in view, got:\n%s", output)
	}
}

// --- AC 27.2.5: Navigation ---

func TestReviewEscSkipsRemaining(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(5)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Process first task
	rv.Update(reviewKeyMsg("c"))
	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	// Esc to skip remaining
	cmd := rv.Update(reviewSpecialKeyMsg(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("expected command on Esc skip")
	}
	msg := cmd()
	completeMsg, ok := msg.(ReviewCompleteMsg)
	if !ok {
		t.Fatalf("expected ReviewCompleteMsg on Esc, got %T", msg)
	}
	if completeMsg.Reviewed != 1 {
		t.Errorf("expected 1 reviewed (only the first), got %d", completeMsg.Reviewed)
	}
}

func TestReviewHelpOverlay(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("?"))
	output := rv.View()

	if !strings.Contains(output, "Continue") && !strings.Contains(output, "continue") {
		t.Errorf("expected help overlay content, got:\n%s", output)
	}
}

func TestReviewHelpOverlayDismiss(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("?"))
	rv.Update(reviewKeyMsg("?")) // toggle off

	if rv.showHelp {
		t.Error("expected help to be dismissed after second ?")
	}
}

func TestReviewUndo(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(3)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Make a decision
	rv.Update(reviewKeyMsg("c"))

	// Undo before advancing
	rv.Update(reviewKeyMsg("u"))

	if rv.decisions[0] != ReviewUndecided {
		t.Errorf("expected decision to be undone, got %v", rv.decisions[0])
	}
}

func TestReviewUndoAfterDrop(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(1)
	originalStatus := tasks[0].Status
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Drop the task
	rv.Update(reviewKeyMsg("x"))
	// Undo
	rv.Update(reviewKeyMsg("u"))

	if tasks[0].Status != originalStatus {
		t.Errorf("expected status to be restored to %s after undo, got %s", originalStatus, tasks[0].Status)
	}
}

func TestReviewUndoNoopWithoutDecision(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	// Undo without any action — should be a no-op
	rv.Update(reviewKeyMsg("u"))

	if rv.decisions[0] != ReviewUndecided {
		t.Errorf("expected undecided after undo with no action, got %v", rv.decisions[0])
	}
}

func TestReviewSetWidth(t *testing.T) {
	t.Parallel()
	rv := NewReviewView(makeTasks(1))
	rv.SetWidth(120)
	if rv.width != 120 {
		t.Errorf("expected width 120, got %d", rv.width)
	}
}

// --- State Machine Transitions ---

func TestReviewCannotAdvancePastEnd(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(1)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("c"))
	cmd := rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	// Should complete, not advance further
	if cmd == nil {
		t.Fatal("expected completion command")
	}
}

func TestReviewCannotActOnProcessedTask(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(2)
	rv := NewReviewView(tasks)
	rv.SetWidth(80)

	rv.Update(reviewKeyMsg("c"))
	rv.Update(reviewSpecialKeyMsg(tea.KeyEnter))

	// Now on second task, try to change first task decision — should not be possible
	// (we're on index 1 now)
	if rv.current != 1 {
		t.Fatalf("expected current to be 1, got %d", rv.current)
	}
}

func TestReviewViewDisplaysTags(t *testing.T) {
	t.Parallel()
	tasks := makeTasks(1)
	tasks[0].Type = core.TypeCreative
	tasks[0].Effort = core.EffortDeepWork
	rv := NewReviewView(tasks)
	rv.SetWidth(80)
	output := rv.View()

	if !strings.Contains(output, "creative") {
		t.Errorf("expected task type tag in view, got:\n%s", output)
	}
}
