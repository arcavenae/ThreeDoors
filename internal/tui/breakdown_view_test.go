package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/intelligence/services"
	tea "github.com/charmbracelet/bubbletea"
)

func makeBreakdownResult() *services.BreakdownResult {
	return &services.BreakdownResult{
		ParentTaskID: "task-1",
		ParentText:   "Build the feature",
		Subtasks: []services.Subtask{
			{Text: "Set up project structure", EffortEstimate: "small"},
			{Text: "Implement core logic", EffortEstimate: "medium"},
			{Text: "Write tests", EffortEstimate: "small"},
		},
		Backend:     "test-llm",
		GeneratedAt: time.Now().UTC(),
	}
}

func makeParentTask() *core.Task {
	t := core.NewTask("Build the feature")
	return t
}

func TestNewBreakdownView(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	if bv.loading {
		t.Error("should not be loading")
	}
	if len(bv.subtasks) != 3 {
		t.Errorf("got %d subtasks, want 3", len(bv.subtasks))
	}
	if bv.SelectedCount() != 3 {
		t.Errorf("all subtasks should be selected by default, got %d", bv.SelectedCount())
	}
}

func TestNewBreakdownViewLoading(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	bv := NewBreakdownViewLoading(parent)

	if !bv.loading {
		t.Error("should be loading")
	}
}

func TestBreakdownViewToggle(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	// Toggle first item off
	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if bv.SelectedCount() != 2 {
		t.Errorf("after toggle off, got %d selected, want 2", bv.SelectedCount())
	}
	if bv.selected[0] {
		t.Error("first item should be deselected")
	}

	// Toggle first item back on
	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if bv.SelectedCount() != 3 {
		t.Errorf("after toggle on, got %d selected, want 3", bv.SelectedCount())
	}
}

func TestBreakdownViewNavigation(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	if bv.cursorIndex != 0 {
		t.Errorf("initial cursor = %d, want 0", bv.cursorIndex)
	}

	// Move down
	bv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if bv.cursorIndex != 1 {
		t.Errorf("after down, cursor = %d, want 1", bv.cursorIndex)
	}

	// Move down again
	bv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if bv.cursorIndex != 2 {
		t.Errorf("after second down, cursor = %d, want 2", bv.cursorIndex)
	}

	// Can't go past end
	bv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if bv.cursorIndex != 2 {
		t.Errorf("should clamp at end, cursor = %d, want 2", bv.cursorIndex)
	}

	// Move up
	bv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if bv.cursorIndex != 1 {
		t.Errorf("after up, cursor = %d, want 1", bv.cursorIndex)
	}
}

func TestBreakdownViewSelectAll(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	// All selected — pressing 'a' should deselect all
	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if bv.SelectedCount() != 0 {
		t.Errorf("after toggle all off, got %d selected, want 0", bv.SelectedCount())
	}

	// None selected — pressing 'a' should select all
	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if bv.SelectedCount() != 3 {
		t.Errorf("after toggle all on, got %d selected, want 3", bv.SelectedCount())
	}
}

func TestBreakdownViewCancel(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	cmd := bv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should return a command")
	}
	msg := cmd()
	if _, ok := msg.(BreakdownCancelMsg); !ok {
		t.Errorf("expected BreakdownCancelMsg, got %T", msg)
	}
}

func TestBreakdownViewImport(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	// Deselect first item
	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	cmd := bv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should return a command")
	}
	msg := cmd()
	importMsg, ok := msg.(BreakdownImportMsg)
	if !ok {
		t.Fatalf("expected BreakdownImportMsg, got %T", msg)
	}
	if len(importMsg.Subtasks) != 2 {
		t.Errorf("imported %d subtasks, want 2", len(importMsg.Subtasks))
	}
	if importMsg.ParentTask != parent {
		t.Error("parent task should be the original task")
	}
}

func TestBreakdownViewImportNoneSelected(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	// Deselect all
	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	cmd := bv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter with nothing selected should return nil")
	}
}

func TestBreakdownViewLoadingCancel(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	bv := NewBreakdownViewLoading(parent)

	cmd := bv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc during loading should return a command")
	}
	msg := cmd()
	if _, ok := msg.(BreakdownCancelMsg); !ok {
		t.Errorf("expected BreakdownCancelMsg, got %T", msg)
	}
}

func TestBreakdownViewErrorDismiss(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	bv := NewBreakdownViewLoading(parent)
	bv.SetError("something went wrong")

	cmd := bv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("any key during error should return a command")
	}
	msg := cmd()
	if _, ok := msg.(BreakdownCancelMsg); !ok {
		t.Errorf("expected BreakdownCancelMsg, got %T", msg)
	}
}

func TestBreakdownViewSetResult(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	bv := NewBreakdownViewLoading(parent)
	result := makeBreakdownResult()
	bv.SetResult(result)

	if bv.loading {
		t.Error("should not be loading after SetResult")
	}
	if len(bv.subtasks) != 3 {
		t.Errorf("got %d subtasks, want 3", len(bv.subtasks))
	}
}

func TestBreakdownViewRender(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)
	bv.SetWidth(80)

	view := bv.View()
	if !strings.Contains(view, "TASK BREAKDOWN") {
		t.Error("view should contain header")
	}
	if !strings.Contains(view, "Set up project structure") {
		t.Error("view should contain subtask text")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("view should contain selected checkbox")
	}
	if !strings.Contains(view, "3/3 selected") {
		t.Error("view should show selection count")
	}
}

func TestBreakdownViewRenderLoading(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	bv := NewBreakdownViewLoading(parent)
	bv.SetWidth(80)

	view := bv.View()
	if !strings.Contains(view, "Breaking down task") {
		t.Error("loading view should show loading message")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("loading view should show cancel hint")
	}
}

func TestBreakdownViewRenderError(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	bv := NewBreakdownViewLoading(parent)
	bv.SetError("network timeout")
	bv.SetWidth(80)

	view := bv.View()
	if !strings.Contains(view, "network timeout") {
		t.Error("error view should show error message")
	}
}

func TestBreakdownViewSelectedSubtasks(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	// Deselect middle item
	bv.Update(tea.KeyMsg{Type: tea.KeyDown})
	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	selected := bv.SelectedSubtasks()
	if len(selected) != 2 {
		t.Fatalf("got %d selected, want 2", len(selected))
	}
	if selected[0].Text != "Set up project structure" {
		t.Errorf("first selected = %q", selected[0].Text)
	}
	if selected[1].Text != "Write tests" {
		t.Errorf("second selected = %q", selected[1].Text)
	}
}

func TestBreakdownViewViNavigation(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if bv.cursorIndex != 1 {
		t.Errorf("j should move down, got cursor=%d", bv.cursorIndex)
	}

	bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if bv.cursorIndex != 0 {
		t.Errorf("k should move up, got cursor=%d", bv.cursorIndex)
	}
}

func TestBreakdownViewQuit(t *testing.T) {
	t.Parallel()
	parent := makeParentTask()
	result := makeBreakdownResult()
	bv := NewBreakdownView(parent, result)

	cmd := bv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return a command")
	}
	msg := cmd()
	if _, ok := msg.(BreakdownCancelMsg); !ok {
		t.Errorf("expected BreakdownCancelMsg, got %T", msg)
	}
}
