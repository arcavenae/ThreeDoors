package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
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
	entries := []tasks.SyncLogEntry{
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
	if !strings.Contains(view, "1-2 of 2") {
		t.Error("view should show entry count")
	}
}

func TestSyncLogView_ReverseOrder(t *testing.T) {
	t.Parallel()
	entries := []tasks.SyncLogEntry{
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

func TestSyncLogView_Scroll(t *testing.T) {
	t.Parallel()
	var entries []tasks.SyncLogEntry
	for i := 0; i < 30; i++ {
		entries = append(entries, tasks.SyncLogEntry{
			Timestamp: time.Now().UTC(),
			Provider:  "Test",
			Operation: "sync",
			Summary:   "entry",
		})
	}

	sv := NewSyncLogView(entries)

	// Scroll down
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if sv.offset != 1 {
		t.Errorf("offset after j = %d, want 1", sv.offset)
	}

	// Scroll up
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if sv.offset != 0 {
		t.Errorf("offset after k = %d, want 0", sv.offset)
	}

	// Don't scroll above 0
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if sv.offset != 0 {
		t.Errorf("offset should not go below 0, got %d", sv.offset)
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
	entries := []tasks.SyncLogEntry{
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

	view := sv.View()

	if !strings.Contains(view, "Conflict") {
		t.Error("view should show conflict resolution entry")
	}
	if !strings.Contains(view, "Buy groceries") {
		t.Error("view should show task text")
	}
}
