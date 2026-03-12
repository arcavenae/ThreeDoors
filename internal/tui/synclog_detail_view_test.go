package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSyncLogDetailView_EmptyEvents(t *testing.T) {
	t.Parallel()
	v := NewSyncLogDetailView("conn-1", nil)
	out := v.View()
	if !strings.Contains(out, "No sync events yet") {
		t.Error("expected empty state message")
	}
	if !strings.Contains(out, "Esc to return") {
		t.Error("expected navigation hint in empty state")
	}
}

func TestSyncLogDetailView_SyncCompleteIndicator(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncComplete,
			Added:        3,
			Updated:      1,
			Removed:      0,
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "✓") {
		t.Error("expected ✓ indicator for sync complete")
	}
	if !strings.Contains(out, "+3") {
		t.Error("expected task count in sync complete event")
	}
}

func TestSyncLogDetailView_ErrorIndicator(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncError,
			Error:        "connection refused",
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "✗") {
		t.Error("expected ✗ indicator for sync error")
	}
	if !strings.Contains(out, "connection refused") {
		t.Error("expected error message in output")
	}
}

func TestSyncLogDetailView_ConflictIndicator(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:        time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID:     "conn-1",
			Type:             connection.EventConflict,
			ConflictTaskText: "Fix login bug",
			Resolution:       "local",
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "⚠") {
		t.Error("expected ⚠ indicator for conflict")
	}
	if !strings.Contains(out, "Fix login bug") {
		t.Error("expected conflict task text")
	}
	if !strings.Contains(out, "local") {
		t.Error("expected resolution detail")
	}
}

func TestSyncLogDetailView_StateChangeIndicator(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventStateChange,
			FromState:    "connected",
			ToState:      "syncing",
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "→") {
		t.Error("expected → indicator for state change")
	}
	if !strings.Contains(out, "connected") {
		t.Error("expected from state")
	}
	if !strings.Contains(out, "syncing") {
		t.Error("expected to state")
	}
}

func TestSyncLogDetailView_SyncStartIndicator(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncStart,
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "⟳") {
		t.Error("expected ⟳ indicator for sync start")
	}
}

func TestSyncLogDetailView_ReauthRequiredIndicator(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventReauthRequired,
			Error:        "token expired",
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "⚠") {
		t.Error("expected ⚠ indicator for reauth required")
	}
	if !strings.Contains(out, "token expired") {
		t.Error("expected error detail for reauth")
	}
}

func TestSyncLogDetailView_Timestamp(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 45, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncStart,
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "2026-03-10 14:30:45") {
		t.Error("expected formatted timestamp in output")
	}
}

func TestSyncLogDetailView_EscReturnsToSourceDetail(t *testing.T) {
	t.Parallel()
	v := NewSyncLogDetailView("conn-42", nil)
	cmd := v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc")
	}
	msg := cmd()
	sdm, ok := msg.(ShowSourceDetailMsg)
	if !ok {
		t.Fatalf("expected ShowSourceDetailMsg, got %T", msg)
	}
	if sdm.ConnectionID != "conn-42" {
		t.Errorf("expected connection ID conn-42, got %s", sdm.ConnectionID)
	}
}

func TestSyncLogDetailView_QReturnsToSourceDetail(t *testing.T) {
	t.Parallel()
	v := NewSyncLogDetailView("conn-99", nil)
	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'q'")
	}
	msg := cmd()
	sdm, ok := msg.(ShowSourceDetailMsg)
	if !ok {
		t.Fatalf("expected ShowSourceDetailMsg, got %T", msg)
	}
	if sdm.ConnectionID != "conn-99" {
		t.Errorf("expected connection ID conn-99, got %s", sdm.ConnectionID)
	}
}

func TestSyncLogDetailView_SetWidth(t *testing.T) {
	t.Parallel()
	v := NewSyncLogDetailView("conn-1", nil)
	v.SetWidth(120)
	if v.width != 120 {
		t.Errorf("expected width 120, got %d", v.width)
	}
}

func TestSyncLogDetailView_SetHeight(t *testing.T) {
	t.Parallel()
	v := NewSyncLogDetailView("conn-1", nil)
	v.SetHeight(40)
	if v.height != 40 {
		t.Errorf("expected height 40, got %d", v.height)
	}
	expectedVP := 40 - syncLogDetailHeaderHeight - syncLogDetailFooterHeight
	if v.viewport.Height != expectedVP {
		t.Errorf("expected viewport height %d, got %d", expectedVP, v.viewport.Height)
	}
}

func TestSyncLogDetailView_SetHeightMinimum(t *testing.T) {
	t.Parallel()
	v := NewSyncLogDetailView("conn-1", nil)
	v.SetHeight(1)
	if v.viewport.Height != 1 {
		t.Errorf("expected viewport height clamped to 1, got %d", v.viewport.Height)
	}
}

func TestSyncLogDetailView_ScrollPercentShown(t *testing.T) {
	t.Parallel()
	var events []connection.SyncEvent
	for i := 0; i < 100; i++ {
		events = append(events, connection.SyncEvent{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncStart,
		})
	}
	v := NewSyncLogDetailView("conn-1", events)
	v.SetHeight(10)
	out := v.View()
	if !strings.Contains(out, "%") {
		t.Error("expected scroll percentage in output")
	}
}

func TestSyncLogDetailView_MultipleEventTypes(t *testing.T) {
	t.Parallel()
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncComplete,
			Added:        2,
			Updated:      0,
			Removed:      1,
		},
		{
			Timestamp:    time.Date(2026, 3, 10, 11, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncError,
			Error:        "timeout",
		},
		{
			Timestamp:        time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC),
			ConnectionID:     "conn-1",
			Type:             connection.EventConflict,
			ConflictTaskText: "Deploy fix",
			Resolution:       "remote",
		},
	}
	v := NewSyncLogDetailView("conn-1", events)
	out := v.View()
	if !strings.Contains(out, "✓") {
		t.Error("expected ✓ for sync complete")
	}
	if !strings.Contains(out, "✗") {
		t.Error("expected ✗ for sync error")
	}
	if !strings.Contains(out, "⚠") {
		t.Error("expected ⚠ for conflict")
	}
}

func TestSyncLogDetailView_HeaderShown(t *testing.T) {
	t.Parallel()
	v := NewSyncLogDetailView("conn-1", []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncStart,
		},
	})
	out := v.View()
	if !strings.Contains(out, "Sync Log") {
		t.Error("expected Sync Log header")
	}
}

func TestSyncLogDetailView_ViewportScrolling(t *testing.T) {
	t.Parallel()
	var events []connection.SyncEvent
	for i := 0; i < 50; i++ {
		events = append(events, connection.SyncEvent{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncStart,
		})
	}
	v := NewSyncLogDetailView("conn-1", events)
	v.SetHeight(10)

	// Scroll down with 'j'
	cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Should not return a navigation command
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(ShowSourceDetailMsg); ok {
			t.Error("'j' should scroll, not navigate")
		}
	}
}
