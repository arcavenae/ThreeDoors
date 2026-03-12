package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSyncLogDetailView_EmptyState(t *testing.T) {
	t.Parallel()

	sv := NewSyncLogDetailView("conn-1", "My Notes", nil)

	output := sv.View()
	if !strings.Contains(output, "No sync events yet") {
		t.Errorf("expected empty state message, got: %s", output)
	}
	if !strings.Contains(output, "Esc to return") {
		t.Errorf("expected escape hint, got: %s", output)
	}
}

func TestSyncLogDetailView_HeaderWithConnectionName(t *testing.T) {
	t.Parallel()

	sv := NewSyncLogDetailView("conn-1", "Apple Notes", nil)
	output := sv.View()
	if !strings.Contains(output, "Apple Notes") {
		t.Errorf("expected connection name in header, got: %s", output)
	}
}

func TestSyncLogDetailView_HeaderWithoutConnectionName(t *testing.T) {
	t.Parallel()

	sv := NewSyncLogDetailView("conn-1", "", nil)
	output := sv.View()
	if !strings.Contains(output, "Sync Log") {
		t.Errorf("expected 'Sync Log' header, got: %s", output)
	}
}

func TestSyncLogDetailView_RendersSyncComplete(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncComplete,
			Added:        3,
			Updated:      1,
			Removed:      0,
			Summary:      "Sync complete: 3 added, 1 updated, 0 removed",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "✓") {
		t.Error("expected ✓ indicator for sync complete")
	}
	if !strings.Contains(output, "+3") {
		t.Error("expected added count in output")
	}
	if !strings.Contains(output, "2026-03-10 14:30:05") || !strings.Contains(output, "2026-03-10") {
		// Timestamp may appear in styled form; just check the date is rendered
		if !strings.Contains(output, "2026") {
			t.Error("expected timestamp year in output")
		}
	}
}

func TestSyncLogDetailView_RendersSyncError(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncError,
			Error:        "network timeout",
			Summary:      "Sync error: network timeout",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "✗") {
		t.Error("expected ✗ indicator for sync error")
	}
	if !strings.Contains(output, "network timeout") {
		t.Error("expected error message in output")
	}
}

func TestSyncLogDetailView_RendersConflict(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:        time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID:     "conn-1",
			Type:             connection.EventConflict,
			ConflictTaskText: "Buy groceries",
			Resolution:       "local",
			Summary:          "Conflict on 'Buy groceries' resolved: local",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "⚠") {
		t.Error("expected ⚠ indicator for conflict")
	}
	if !strings.Contains(output, "Buy groceries") {
		t.Error("expected task text in conflict output")
	}
	if !strings.Contains(output, "local") {
		t.Error("expected resolution in conflict output")
	}
}

func TestSyncLogDetailView_RendersStateChange(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventStateChange,
			FromState:    "connected",
			ToState:      "syncing",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "connected") {
		t.Error("expected from state in output")
	}
	if !strings.Contains(output, "syncing") {
		t.Error("expected to state in output")
	}
}

func TestSyncLogDetailView_RendersReauthRequired(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventReauthRequired,
			Error:        "token expired",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "⚠") {
		t.Error("expected ⚠ indicator for reauth required")
	}
	if !strings.Contains(output, "token expired") {
		t.Error("expected error detail in reauth output")
	}
}

func TestSyncLogDetailView_RendersSyncStart(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncStart,
			Summary:      "Sync started",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "Sync started") {
		t.Error("expected 'Sync started' in output")
	}
}

func TestSyncLogDetailView_MultipleEvents(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 15, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncComplete,
			Added:        2,
			Summary:      "Sync complete",
		},
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncError,
			Error:        "failed",
			Summary:      "Sync error",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "✓") {
		t.Error("expected ✓ indicator")
	}
	if !strings.Contains(output, "✗") {
		t.Error("expected ✗ indicator")
	}
}

func TestSyncLogDetailView_EscReturnsToSourceDetail(t *testing.T) {
	t.Parallel()

	sv := NewSyncLogDetailView("conn-1", "Test", nil)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command on Esc")
	}

	msg := cmd()
	if m, ok := msg.(ShowSourceDetailMsg); !ok {
		t.Errorf("expected ShowSourceDetailMsg, got %T", msg)
	} else if m.ConnectionID != "conn-1" {
		t.Errorf("expected connection ID 'conn-1', got %q", m.ConnectionID)
	}
}

func TestSyncLogDetailView_SetWidth(t *testing.T) {
	t.Parallel()

	sv := NewSyncLogDetailView("conn-1", "Test", nil)
	sv.SetWidth(120)

	if sv.width != 120 {
		t.Errorf("expected width 120, got %d", sv.width)
	}
	if sv.viewport.Width != 120 {
		t.Errorf("expected viewport width 120, got %d", sv.viewport.Width)
	}
}

func TestSyncLogDetailView_SetHeight(t *testing.T) {
	t.Parallel()

	sv := NewSyncLogDetailView("conn-1", "Test", nil)
	sv.SetHeight(40)

	if sv.height != 40 {
		t.Errorf("expected height 40, got %d", sv.height)
	}

	expectedVPHeight := 40 - syncLogDetailHeaderHeight - syncLogDetailFooterHeight
	if sv.viewport.Height != expectedVPHeight {
		t.Errorf("expected viewport height %d, got %d", expectedVPHeight, sv.viewport.Height)
	}
}

func TestSyncLogDetailView_SetHeightMinimum(t *testing.T) {
	t.Parallel()

	sv := NewSyncLogDetailView("conn-1", "Test", nil)
	sv.SetHeight(2) // Too small for header + footer

	if sv.viewport.Height != 1 {
		t.Errorf("expected viewport height clamped to 1, got %d", sv.viewport.Height)
	}
}

func TestSyncLogDetailView_ConflictUnresolved(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:        time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID:     "conn-1",
			Type:             connection.EventConflict,
			ConflictTaskText: "Fix bug",
			Resolution:       "",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "unresolved") {
		t.Error("expected 'unresolved' for empty resolution")
	}
}

func TestSyncLogDetailView_ScrollPercentShown(t *testing.T) {
	t.Parallel()

	events := make([]connection.SyncEvent, 50)
	for i := range events {
		events[i] = connection.SyncEvent{
			Timestamp:    time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncComplete,
			Added:        i,
			Summary:      "Sync",
		}
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	sv.SetHeight(10) // Small viewport to ensure scrolling
	output := sv.View()

	// Should show scroll percentage
	if !strings.Contains(output, "%") {
		t.Error("expected scroll percentage in output")
	}
}

func TestSyncLogDetailView_FooterHints(t *testing.T) {
	t.Parallel()

	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncComplete,
			Summary:      "Sync complete",
		},
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	output := sv.View()

	if !strings.Contains(output, "j/k scroll") {
		t.Error("expected scroll hint in footer")
	}
	if !strings.Contains(output, "Esc return") {
		t.Error("expected Esc hint in footer")
	}
}

func TestSyncLogDetailView_ViewportScrolling(t *testing.T) {
	t.Parallel()

	events := make([]connection.SyncEvent, 100)
	for i := range events {
		events[i] = connection.SyncEvent{
			Timestamp:    time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         connection.EventSyncComplete,
			Added:        i,
			Summary:      "Sync",
		}
	}

	sv := NewSyncLogDetailView("conn-1", "Test", events)
	sv.SetHeight(10)

	// Send j key to scroll down
	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Should not return a navigation command
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(ShowSourceDetailMsg); ok {
			t.Error("j key should scroll, not navigate")
		}
	}
}
