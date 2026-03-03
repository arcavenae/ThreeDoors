package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func newTestDoorsView(texts ...string) *DoorsView {
	pool := core.NewTaskPool()
	for _, t := range texts {
		pool.AddTask(core.NewTask(t))
	}
	tracker := core.NewSessionTracker()
	return NewDoorsView(pool, tracker)
}

// --- Creation and Defaults ---

func TestNewDoorsView_DefaultsNoSelection(t *testing.T) {
	dv := newTestDoorsView("task1", "task2", "task3")
	if dv.selectedDoorIndex != -1 {
		t.Errorf("expected -1, got %d", dv.selectedDoorIndex)
	}
}

func TestNewDoorsView_LoadsDoors(t *testing.T) {
	dv := newTestDoorsView("task1", "task2", "task3")
	if len(dv.currentDoors) == 0 {
		t.Error("currentDoors should be populated")
	}
}

func TestNewDoorsView_MaxThreeDoors(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3", "t4", "t5", "t6")
	if len(dv.currentDoors) > 3 {
		t.Errorf("expected at most 3 doors, got %d", len(dv.currentDoors))
	}
}

// --- RefreshDoors ---

func TestRefreshDoors_ResetsSelection(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3", "t4", "t5")
	dv.selectedDoorIndex = 1
	dv.RefreshDoors()
	if dv.selectedDoorIndex != -1 {
		t.Errorf("expected -1 after refresh, got %d", dv.selectedDoorIndex)
	}
}

// --- GetCurrentDoorTexts ---

func TestGetCurrentDoorTexts_ReturnsTexts(t *testing.T) {
	dv := newTestDoorsView("Alpha", "Beta", "Gamma")
	texts := dv.GetCurrentDoorTexts()
	if len(texts) == 0 {
		t.Fatal("expected door texts")
	}
	for _, text := range texts {
		if text == "" {
			t.Error("door text should not be empty")
		}
	}
}

// --- IncrementCompleted ---

func TestIncrementCompleted_IncrementsCounter(t *testing.T) {
	dv := newTestDoorsView("task1")
	dv.IncrementCompleted()
	dv.IncrementCompleted()
	if dv.completedCount != 2 {
		t.Errorf("expected 2, got %d", dv.completedCount)
	}
}

// --- View Rendering ---

func TestDoorsView_View_RendersHeader(t *testing.T) {
	dv := newTestDoorsView("task1", "task2", "task3")
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("View should contain 'ThreeDoors' header")
	}
}

func TestDoorsView_View_RendersHelp(t *testing.T) {
	dv := newTestDoorsView("task1", "task2", "task3")
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "quit") {
		t.Error("View should contain quit instruction")
	}
}

func TestDoorsView_View_EmptyPool_ShowsAllDone(t *testing.T) {
	pool := core.NewTaskPool()
	dv := NewDoorsView(pool, nil)
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "All tasks done") {
		t.Errorf("Empty pool should show 'All tasks done', got: %s", view)
	}
}

func TestDoorsView_View_CompletionCounter_Visible(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.IncrementCompleted()
	view := dv.View()
	if !strings.Contains(view, "Completed this session: 1") {
		t.Errorf("View should contain completion counter, got: %s", view)
	}
}

func TestDoorsView_View_CompletionCounter_ZeroNotShown(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	view := dv.View()
	if strings.Contains(view, "Completed this session") {
		t.Error("Completion counter should not be shown when count is 0")
	}
}

func TestDoorsView_View_FooterMessage(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	view := dv.View()
	// Footer should contain one of the greeting messages (rotating pool)
	found := false
	for _, msg := range greetingMessages {
		if strings.Contains(view, msg) {
			found = true
			break
		}
	}
	if !found {
		t.Error("View should contain a footer message from the greeting pool")
	}
}

func TestDoorsView_View_ShowsStatusIndicator(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	view := dv.View()
	// Each door shows status indicator
	if !strings.Contains(view, "[todo]") {
		t.Error("View should show [todo] status indicator for doors")
	}
}

// --- Story 1.6: Essential Polish - Greeting Tests ---

func TestDoorsView_GreetingDisplayed(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	view := dv.View()

	// View must contain at least one greeting from the pool
	found := false
	for _, greeting := range greetingMessages {
		if strings.Contains(view, greeting) {
			found = true
			break
		}
	}
	if !found {
		t.Error("View should contain a greeting message from the greetingMessages pool")
	}
}

func TestDoorsView_GreetingIsFromPool(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")

	greeting := dv.Greeting()
	if greeting == "" {
		t.Fatal("DoorsView.Greeting() should return a non-empty string")
	}

	found := false
	for _, msg := range greetingMessages {
		if greeting == msg {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Greeting %q is not in the greetingMessages pool", greeting)
	}
}

func TestDoorsView_GreetingSetOnce(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	first := dv.Greeting()
	// Greeting should not change on re-render
	dv.SetWidth(80)
	_ = dv.View()
	second := dv.Greeting()
	if first != second {
		t.Errorf("Greeting should not change between renders: first=%q, second=%q", first, second)
	}
}

func TestDoorsView_GreetingNotChangedByRefresh(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3", "t4", "t5")
	original := dv.Greeting()
	dv.RefreshDoors()
	after := dv.Greeting()
	if original != after {
		t.Errorf("Greeting should persist across door refreshes: before=%q, after=%q", original, after)
	}
}

// --- Story 1.6: Essential Polish - Narrow Terminal Fallback ---

func TestDoorsView_NarrowTerminal_StillRenders(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(40)
	view := dv.View()
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("Narrow terminal should still render the header")
	}
	// Should still show task content
	found := false
	for _, door := range dv.currentDoors {
		if strings.Contains(view, door.Text) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Narrow terminal should still render door content")
	}
}
