package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
)

func newTestDoorsView(texts ...string) *DoorsView {
	pool := tasks.NewTaskPool()
	for _, t := range texts {
		pool.AddTask(tasks.NewTask(t))
	}
	tracker := tasks.NewSessionTracker()
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
	pool := tasks.NewTaskPool()
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

func TestDoorsView_View_ProgressMessage(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "Progress over perfection") {
		t.Error("View should contain 'Progress over perfection' message")
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
