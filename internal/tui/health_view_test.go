package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Test Helpers ---

func newTestHealthView(items ...tasks.HealthCheckItem) *HealthView {
	overall := tasks.HealthOK
	for _, item := range items {
		if item.Status == tasks.HealthFail {
			overall = tasks.HealthFail
			break
		}
		if item.Status == tasks.HealthWarn && overall == tasks.HealthOK {
			overall = tasks.HealthWarn
		}
	}
	result := tasks.HealthCheckResult{
		Items:    items,
		Overall:  overall,
		Duration: 42 * time.Millisecond,
	}
	hv := NewHealthView(result)
	hv.SetWidth(80)
	return hv
}

// --- View Rendering Tests ---

func TestHealthView_View_RendersOKItem(t *testing.T) {
	hv := newTestHealthView(tasks.HealthCheckItem{
		Name:    "Task File",
		Status:  tasks.HealthOK,
		Message: "12 tasks loaded successfully",
	})
	view := hv.View()
	if !strings.Contains(view, "[OK]") {
		t.Error("view should contain [OK] indicator")
	}
	if !strings.Contains(view, "Task File") {
		t.Error("view should contain item name")
	}
}

func TestHealthView_View_RendersFAILItem(t *testing.T) {
	hv := newTestHealthView(tasks.HealthCheckItem{
		Name:       "Apple Notes",
		Status:     tasks.HealthFail,
		Message:    "Cannot access Apple Notes database",
		Suggestion: "Grant Full Disk Access in System Settings",
	})
	view := hv.View()
	if !strings.Contains(view, "[FAIL]") {
		t.Error("view should contain [FAIL] indicator")
	}
	if !strings.Contains(view, "→") {
		t.Error("view should contain → arrow for suggestion")
	}
	if !strings.Contains(view, "Grant Full Disk Access") {
		t.Error("view should contain suggestion text")
	}
}

func TestHealthView_View_RendersWARNItem(t *testing.T) {
	hv := newTestHealthView(tasks.HealthCheckItem{
		Name:       "Sync Status",
		Status:     tasks.HealthWarn,
		Message:    "Last sync was 48 hours ago",
		Suggestion: "Press S in doors view to trigger a sync",
	})
	view := hv.View()
	if !strings.Contains(view, "[WARN]") {
		t.Error("view should contain [WARN] indicator")
	}
	if !strings.Contains(view, "→") {
		t.Error("view should contain → arrow for suggestion")
	}
}

func TestHealthView_View_RendersOverallAndDuration(t *testing.T) {
	hv := newTestHealthView(tasks.HealthCheckItem{
		Name:   "Test",
		Status: tasks.HealthOK,
	})
	view := hv.View()
	if !strings.Contains(view, "Overall:") {
		t.Error("view should contain 'Overall:' footer")
	}
	if !strings.Contains(view, "Completed in") {
		t.Error("view should contain 'Completed in' duration")
	}
}

func TestHealthView_View_EmptyResult(t *testing.T) {
	hv := newTestHealthView() // no items
	view := hv.View()
	if !strings.Contains(view, "Overall:") {
		t.Error("empty result should still contain 'Overall:' footer")
	}
}

// --- Update Tests ---

func TestHealthView_Update_EscReturnsToDoorsMsg(t *testing.T) {
	hv := newTestHealthView(tasks.HealthCheckItem{
		Name:   "Test",
		Status: tasks.HealthOK,
	})
	cmd := hv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Esc press")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestHealthView_Update_QKeyReturnsNil(t *testing.T) {
	hv := newTestHealthView(tasks.HealthCheckItem{
		Name:   "Test",
		Status: tasks.HealthOK,
	})
	cmd := hv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Error("expected nil cmd on 'q' press from health view — q should NOT quit from this view")
	}
}

// --- SetWidth Test ---

func TestHealthView_SetWidth(t *testing.T) {
	hv := newTestHealthView()
	hv.SetWidth(120)
	if hv.width != 120 {
		t.Errorf("SetWidth(120) → width = %d, want 120", hv.width)
	}
}
