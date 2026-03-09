package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSyncLogView_Empty(t *testing.T) {
	t.Parallel()
	sv := NewSyncLogView(nil)
	sv.SetWidth(80)

	view := sv.View()

	if !strings.Contains(view, "Sync Log") {
		t.Error("view should contain 'Sync Log' header")
	}
	if !strings.Contains(view, "No sync operations recorded") {
		t.Error("view should show empty message")
	}
}

func TestSyncLogView_WithEntries(t *testing.T) {
	t.Parallel()
	entries := []core.SyncLogEntry{
		{
			Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			Provider:  "Local",
			Operation: "sync",
			Summary:   "Synced: 2 new, 1 updated",
		},
		{
			Timestamp: time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC),
			Provider:  "WAL",
			Operation: "error",
			Error:     "connection refused",
		},
	}

	sv := NewSyncLogView(entries)
	sv.SetWidth(80)
	sv.SetHeight(40)

	view := sv.View()

	if !strings.Contains(view, "Sync Log") {
		t.Error("view should contain header")
	}
	if !strings.Contains(view, "Synced: 2 new, 1 updated") {
		t.Error("view should show sync summary")
	}
	if !strings.Contains(view, "connection refused") {
		t.Error("view should show error")
	}
	if !strings.Contains(view, "%") {
		t.Error("view should show scroll percentage")
	}
}

func TestSyncLogView_ReverseOrder(t *testing.T) {
	t.Parallel()
	entries := []core.SyncLogEntry{
		{
			Timestamp: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			Provider:  "A",
			Operation: "sync",
			Summary:   "first",
		},
		{
			Timestamp: time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC),
			Provider:  "B",
			Operation: "sync",
			Summary:   "second",
		},
	}

	sv := NewSyncLogView(entries)

	// Newest should be first after reversal
	if sv.entries[0].Provider != "B" {
		t.Errorf("first entry Provider = %q, want %q (newest first)", sv.entries[0].Provider, "B")
	}
}

func TestSyncLogView_ViewportScroll(t *testing.T) {
	t.Parallel()
	var entries []core.SyncLogEntry
	for i := 0; i < 30; i++ {
		entries = append(entries, core.SyncLogEntry{
			Timestamp: time.Now().UTC(),
			Provider:  "Test",
			Operation: "sync",
			Summary:   "entry",
		})
	}

	sv := NewSyncLogView(entries)
	sv.SetHeight(10) // Small height so content overflows

	initialOffset := sv.viewport.YOffset

	// Scroll down
	sv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sv.viewport.YOffset <= initialOffset {
		t.Error("viewport should scroll down after down key")
	}

	// Scroll up
	offset := sv.viewport.YOffset
	sv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sv.viewport.YOffset >= offset {
		t.Error("viewport should scroll up after up key")
	}
}

func TestSyncLogView_MouseWheelEnabled(t *testing.T) {
	t.Parallel()
	sv := NewSyncLogView(nil)

	if !sv.viewport.MouseWheelEnabled {
		t.Error("viewport should have mouse wheel enabled")
	}
}

func TestSyncLogView_EscReturns(t *testing.T) {
	t.Parallel()
	sv := NewSyncLogView(nil)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command on Esc")
	}

	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestSyncLogView_ConflictResolved(t *testing.T) {
	t.Parallel()
	entries := []core.SyncLogEntry{
		{
			Timestamp:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			Provider:   "Local",
			Operation:  "conflict_resolved",
			TaskText:   "Buy groceries",
			Resolution: "local",
		},
	}

	sv := NewSyncLogView(entries)
	sv.SetWidth(80)
	sv.SetHeight(40)

	view := sv.View()

	if !strings.Contains(view, "Conflict") {
		t.Error("view should show conflict resolution entry")
	}
	if !strings.Contains(view, "Buy groceries") {
		t.Error("view should show task text")
	}
}

func TestSyncLogView_SetWidth(t *testing.T) {
	t.Parallel()
	sv := NewSyncLogView(nil)

	sv.SetWidth(120)
	if sv.width != 120 {
		t.Errorf("width = %d, want 120", sv.width)
	}
	if sv.viewport.Width != 120 {
		t.Errorf("viewport width = %d, want 120", sv.viewport.Width)
	}
}

func TestSyncLogView_SetHeight(t *testing.T) {
	t.Parallel()
	sv := NewSyncLogView(nil)

	sv.SetHeight(40)
	if sv.height != 40 {
		t.Errorf("height = %d, want 40", sv.height)
	}
	wantVPHeight := 40 - syncLogHeaderHeight - syncLogFooterHeight
	if sv.viewport.Height != wantVPHeight {
		t.Errorf("viewport height = %d, want %d", sv.viewport.Height, wantVPHeight)
	}
}

func TestSyncLogView_SetHeight_MinimumClamp(t *testing.T) {
	t.Parallel()
	sv := NewSyncLogView(nil)

	sv.SetHeight(1)
	if sv.viewport.Height < 1 {
		t.Errorf("viewport height should be at least 1, got %d", sv.viewport.Height)
	}
}

func TestSyncLogView_ScrollPercent(t *testing.T) {
	t.Parallel()
	var entries []core.SyncLogEntry
	for i := 0; i < 30; i++ {
		entries = append(entries, core.SyncLogEntry{
			Timestamp: time.Now().UTC(),
			Provider:  "Test",
			Operation: "sync",
			Summary:   "entry",
		})
	}

	sv := NewSyncLogView(entries)
	sv.SetHeight(10)

	view := sv.View()
	if !strings.Contains(view, "%") {
		t.Error("view should contain scroll percentage indicator")
	}
}
