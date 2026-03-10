package connection

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncEventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typ      SyncEventType
		expected string
	}{
		{"sync complete", EventSyncComplete, "sync_complete"},
		{"sync error", EventSyncError, "sync_error"},
		{"conflict", EventConflict, "conflict"},
		{"state change", EventStateChange, "state_change"},
		{"sync start", EventSyncStart, "sync_start"},
		{"reauth required", EventReauthRequired, "reauth_required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.typ) != tt.expected {
				t.Errorf("got %q, want %q", tt.typ, tt.expected)
			}
		})
	}
}

func TestSyncEventJSONRoundTrip(t *testing.T) {
	t.Parallel()

	event := SyncEvent{
		Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		ConnectionID: "conn-1",
		Type:         EventSyncComplete,
		Added:        5,
		Updated:      3,
		Removed:      1,
		Summary:      "Sync complete: 5 added, 3 updated, 1 removed",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SyncEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ConnectionID != event.ConnectionID {
		t.Errorf("connection_id: got %q, want %q", decoded.ConnectionID, event.ConnectionID)
	}
	if decoded.Type != event.Type {
		t.Errorf("type: got %q, want %q", decoded.Type, event.Type)
	}
	if decoded.Added != event.Added {
		t.Errorf("added: got %d, want %d", decoded.Added, event.Added)
	}
}

func newTestLog(t *testing.T) *SyncEventLog {
	t.Helper()
	dir := t.TempDir()
	return NewSyncEventLog(dir)
}

func TestAppendAndSyncLog(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	// Append three events.
	for i := range 3 {
		err := log.Append(SyncEvent{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, i, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         EventSyncComplete,
			Added:        i + 1,
			Summary:      "test",
		})
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	// SyncLog with no limit returns all, most recent first.
	events, err := log.SyncLog("conn-1", 0)
	if err != nil {
		t.Fatalf("SyncLog: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// Most recent first.
	if events[0].Added != 3 {
		t.Errorf("first event Added: got %d, want 3", events[0].Added)
	}
	if events[2].Added != 1 {
		t.Errorf("last event Added: got %d, want 1", events[2].Added)
	}
}

func TestSyncLogWithLimit(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	for i := range 10 {
		err := log.Append(SyncEvent{
			Timestamp:    time.Date(2026, 3, 10, 12, 0, i, 0, time.UTC),
			ConnectionID: "conn-1",
			Type:         EventSyncComplete,
			Added:        i + 1,
			Summary:      "test",
		})
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	events, err := log.SyncLog("conn-1", 3)
	if err != nil {
		t.Fatalf("SyncLog: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// Most recent 3, in reverse chronological order.
	if events[0].Added != 10 {
		t.Errorf("first event Added: got %d, want 10", events[0].Added)
	}
	if events[2].Added != 8 {
		t.Errorf("last event Added: got %d, want 8", events[2].Added)
	}
}

func TestSyncLogEmptyConnection(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	events, err := log.SyncLog("nonexistent", 10)
	if err != nil {
		t.Fatalf("SyncLog: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestAppendEmptyConnectionID(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	err := log.Append(SyncEvent{
		Timestamp: time.Now().UTC(),
		Type:      EventSyncComplete,
		Summary:   "test",
	})
	if err == nil {
		t.Fatal("expected error for empty connection ID")
	}
}

func TestEventsSince(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	for i := range 5 {
		err := log.Append(SyncEvent{
			Timestamp:    base.Add(time.Duration(i) * time.Minute),
			ConnectionID: "conn-1",
			Type:         EventSyncComplete,
			Added:        i + 1,
			Summary:      "test",
		})
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	// Events since minute 2 (inclusive) — should get minutes 2, 3, 4.
	since := base.Add(2 * time.Minute)
	events, err := log.EventsSince("conn-1", since)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// Reverse chronological.
	if events[0].Added != 5 {
		t.Errorf("first event Added: got %d, want 5", events[0].Added)
	}
}

func TestEventsByType(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	connID := "conn-1"
	err := log.LogSyncComplete(connID, 5, 2, 1)
	if err != nil {
		t.Fatalf("LogSyncComplete: %v", err)
	}
	err = log.LogSyncError(connID, errors.New("timeout"))
	if err != nil {
		t.Fatalf("LogSyncError: %v", err)
	}
	err = log.LogConflict(connID, "task-1", "Buy groceries", "local")
	if err != nil {
		t.Fatalf("LogConflict: %v", err)
	}
	err = log.LogSyncComplete(connID, 3, 0, 0)
	if err != nil {
		t.Fatalf("LogSyncComplete: %v", err)
	}

	// Filter by sync_complete.
	events, err := log.EventsByType(connID, EventSyncComplete, 0)
	if err != nil {
		t.Fatalf("EventsByType: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 sync_complete events, got %d", len(events))
	}

	// Filter by sync_error.
	events, err = log.EventsByType(connID, EventSyncError, 0)
	if err != nil {
		t.Fatalf("EventsByType: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 sync_error event, got %d", len(events))
	}
	if events[0].Error != "timeout" {
		t.Errorf("error: got %q, want %q", events[0].Error, "timeout")
	}
}

func TestEventsByTypeWithLimit(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	connID := "conn-1"
	for i := range 5 {
		err := log.LogSyncComplete(connID, i+1, 0, 0)
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	events, err := log.EventsByType(connID, EventSyncComplete, 2)
	if err != nil {
		t.Fatalf("EventsByType: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	// Most recent first.
	if events[0].Added != 5 {
		t.Errorf("first event Added: got %d, want 5", events[0].Added)
	}
}

func TestLogHelpers(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	connID := "conn-1"

	t.Run("LogSyncComplete", func(t *testing.T) {
		err := log.LogSyncComplete(connID, 5, 3, 1)
		if err != nil {
			t.Fatalf("LogSyncComplete: %v", err)
		}
		events, err := log.SyncLog(connID, 1)
		if err != nil {
			t.Fatalf("SyncLog: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		e := events[0]
		if e.Type != EventSyncComplete {
			t.Errorf("type: got %q, want %q", e.Type, EventSyncComplete)
		}
		if e.Added != 5 || e.Updated != 3 || e.Removed != 1 {
			t.Errorf("counts: got %d/%d/%d, want 5/3/1", e.Added, e.Updated, e.Removed)
		}
	})

	t.Run("LogSyncError", func(t *testing.T) {
		err := log.LogSyncError(connID, errors.New("connection reset"))
		if err != nil {
			t.Fatalf("LogSyncError: %v", err)
		}
		events, err := log.EventsByType(connID, EventSyncError, 1)
		if err != nil {
			t.Fatalf("EventsByType: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Error != "connection reset" {
			t.Errorf("error: got %q, want %q", events[0].Error, "connection reset")
		}
	})

	t.Run("LogConflict", func(t *testing.T) {
		err := log.LogConflict(connID, "task-42", "Review PR", "remote")
		if err != nil {
			t.Fatalf("LogConflict: %v", err)
		}
		events, err := log.EventsByType(connID, EventConflict, 1)
		if err != nil {
			t.Fatalf("EventsByType: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		e := events[0]
		if e.ConflictTaskID != "task-42" {
			t.Errorf("task ID: got %q, want %q", e.ConflictTaskID, "task-42")
		}
		if e.Resolution != "remote" {
			t.Errorf("resolution: got %q, want %q", e.Resolution, "remote")
		}
	})

	t.Run("LogStateChange", func(t *testing.T) {
		err := log.LogStateChange(connID, StateDisconnected, StateConnecting, "")
		if err != nil {
			t.Fatalf("LogStateChange: %v", err)
		}
		events, err := log.EventsByType(connID, EventStateChange, 1)
		if err != nil {
			t.Fatalf("EventsByType: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		e := events[0]
		if e.FromState != "disconnected" {
			t.Errorf("from_state: got %q, want %q", e.FromState, "disconnected")
		}
		if e.ToState != "connecting" {
			t.Errorf("to_state: got %q, want %q", e.ToState, "connecting")
		}
	})

	t.Run("LogStateChangeWithError", func(t *testing.T) {
		err := log.LogStateChange(connID, StateConnecting, StateError, "auth failed")
		if err != nil {
			t.Fatalf("LogStateChange: %v", err)
		}
		events, err := log.EventsByType(connID, EventStateChange, 1)
		if err != nil {
			t.Fatalf("EventsByType: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Error != "auth failed" {
			t.Errorf("error: got %q, want %q", events[0].Error, "auth failed")
		}
	})
}

func TestPerConnectionIsolation(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)

	err := log.LogSyncComplete("conn-a", 10, 0, 0)
	if err != nil {
		t.Fatalf("append conn-a: %v", err)
	}
	err = log.LogSyncComplete("conn-b", 20, 0, 0)
	if err != nil {
		t.Fatalf("append conn-b: %v", err)
	}

	eventsA, err := log.SyncLog("conn-a", 0)
	if err != nil {
		t.Fatalf("SyncLog conn-a: %v", err)
	}
	eventsB, err := log.SyncLog("conn-b", 0)
	if err != nil {
		t.Fatalf("SyncLog conn-b: %v", err)
	}

	if len(eventsA) != 1 {
		t.Errorf("conn-a: expected 1 event, got %d", len(eventsA))
	}
	if len(eventsB) != 1 {
		t.Errorf("conn-b: expected 1 event, got %d", len(eventsB))
	}
	if eventsA[0].Added != 10 {
		t.Errorf("conn-a Added: got %d, want 10", eventsA[0].Added)
	}
	if eventsB[0].Added != 20 {
		t.Errorf("conn-b Added: got %d, want 20", eventsB[0].Added)
	}
}

func TestRollingRetention(t *testing.T) {
	t.Parallel()
	log := newTestLog(t)
	connID := "conn-retention"

	// Write maxEventsPerFile + 50 events.
	for i := range maxEventsPerFile + 50 {
		err := log.Append(SyncEvent{
			Timestamp:    time.Date(2026, 3, 10, 0, 0, i, 0, time.UTC),
			ConnectionID: connID,
			Type:         EventSyncComplete,
			Added:        i + 1,
			Summary:      "test",
		})
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	// After the next append, truncation should have occurred.
	events, err := log.SyncLog(connID, 0)
	if err != nil {
		t.Fatalf("SyncLog: %v", err)
	}

	// Should have at most maxEventsPerFile + 1 events
	// (truncation happens before append, so 1000 kept + up to 50 appended after last truncation).
	// The exact count depends on when truncation triggered. Let's verify it's bounded.
	if len(events) > maxEventsPerFile+50 {
		t.Errorf("expected at most %d events after retention, got %d", maxEventsPerFile+50, len(events))
	}

	// The most recent event should have Added = maxEventsPerFile + 50.
	if len(events) > 0 && events[0].Added != maxEventsPerFile+50 {
		t.Errorf("most recent Added: got %d, want %d", events[0].Added, maxEventsPerFile+50)
	}

	// Verify oldest events were trimmed by checking that the oldest remaining
	// event's Added value is > 1.
	if len(events) > 0 {
		oldest := events[len(events)-1]
		if oldest.Added <= 1 {
			t.Errorf("oldest event should have been trimmed, but Added=%d", oldest.Added)
		}
	}
}

func TestCorruptEntriesSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	log := NewSyncEventLog(dir)
	connID := "conn-corrupt"

	// Write a valid event.
	err := log.LogSyncComplete(connID, 1, 0, 0)
	if err != nil {
		t.Fatalf("LogSyncComplete: %v", err)
	}

	// Manually append a corrupt line.
	path := filepath.Join(dir, syncLogDir, connID+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	_, _ = f.WriteString("not valid json{{{}\n")
	_ = f.Close()

	// Write another valid event.
	err = log.LogSyncComplete(connID, 2, 0, 0)
	if err != nil {
		t.Fatalf("LogSyncComplete: %v", err)
	}

	events, err := log.SyncLog(connID, 0)
	if err != nil {
		t.Fatalf("SyncLog: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 valid events (corrupt skipped), got %d", len(events))
	}
}

func TestSyncLogDirCreation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	nestedDir := filepath.Join(dir, "deep", "nested")
	log := NewSyncEventLog(nestedDir)

	err := log.LogSyncComplete("conn-1", 1, 0, 0)
	if err != nil {
		t.Fatalf("LogSyncComplete in nested dir: %v", err)
	}

	// Verify the directory was created.
	logDir := filepath.Join(nestedDir, syncLogDir)
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("stat sync-logs dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("sync-logs should be a directory")
	}
}
