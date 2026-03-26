package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
)

// newTestLogManager creates a ConnectionManager and SyncEventLog with test data.
func newTestLogManager(t *testing.T) (*connection.ConnectionManager, *connection.SyncEventLog) {
	t.Helper()

	manager := connection.NewConnectionManager(nil)
	conn, err := manager.Add("todoist", "Personal Todoist", map[string]string{})
	if err != nil {
		t.Fatalf("add connection: %v", err)
	}

	dir := t.TempDir()
	eventLog := connection.NewSyncEventLog(dir)

	// Add test events in chronological order.
	events := []connection.SyncEvent{
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC),
			ConnectionID: conn.ID,
			Type:         connection.EventSyncComplete,
			Added:        2,
			Updated:      1,
			Summary:      "Sync complete: 2 added, 1 updated, 0 removed",
		},
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 31, 0, 0, time.UTC),
			ConnectionID: conn.ID,
			Type:         connection.EventSyncError,
			Error:        "connection timeout",
			Summary:      "Sync error: connection timeout",
		},
		{
			Timestamp:        time.Date(2026, 3, 10, 14, 32, 0, 0, time.UTC),
			ConnectionID:     conn.ID,
			Type:             connection.EventConflict,
			ConflictTaskText: "Buy groceries",
			Resolution:       "remote",
			Summary:          "Conflict on 'Buy groceries' resolved: remote",
		},
		{
			Timestamp:    time.Date(2026, 3, 10, 14, 33, 0, 0, time.UTC),
			ConnectionID: conn.ID,
			Type:         connection.EventSyncComplete,
			Added:        0,
			Updated:      0,
			Summary:      "Sync complete: 0 added, 0 updated, 0 removed",
		},
	}
	for _, e := range events {
		if err := eventLog.Append(e); err != nil {
			t.Fatalf("append event: %v", err)
		}
	}

	return manager, eventLog
}

func TestRunSourcesLog(t *testing.T) {
	t.Parallel()

	t.Run("table output for named connection", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "Personal Todoist", 20, false, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "TIMESTAMP") {
			t.Error("missing table header TIMESTAMP")
		}
		if !strings.Contains(out, "STATUS") {
			t.Error("missing table header STATUS")
		}
		if !strings.Contains(out, "DESCRIPTION") {
			t.Error("missing table header DESCRIPTION")
		}
		if !strings.Contains(out, "Sync complete") {
			t.Error("missing sync complete event")
		}
		if !strings.Contains(out, "Sync error") {
			t.Error("missing sync error event")
		}
		if !strings.Contains(out, "Conflict") {
			t.Error("missing conflict event")
		}
	})

	t.Run("default limit of 20", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		// 4 events exist, limit=20 should show all
		err := runSourcesLogTo(cmd, manager, eventLog, "Personal Todoist", 20, false, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		out := buf.String()
		// Count data lines (excluding header)
		lines := strings.Split(strings.TrimSpace(out), "\n")
		// Header + 4 events = 5 lines
		if len(lines) != 5 {
			t.Errorf("expected 5 lines (1 header + 4 events), got %d:\n%s", len(lines), out)
		}
	})

	t.Run("last flag limits events", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "Personal Todoist", 2, false, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		out := buf.String()
		lines := strings.Split(strings.TrimSpace(out), "\n")
		// Header + 2 events = 3 lines
		if len(lines) != 3 {
			t.Errorf("expected 3 lines (1 header + 2 events), got %d:\n%s", len(lines), out)
		}
	})

	t.Run("errors flag filters to errors and conflicts", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "Personal Todoist", 20, true, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Sync error") {
			t.Error("missing error event in filtered output")
		}
		if !strings.Contains(out, "Conflict") {
			t.Error("missing conflict event in filtered output")
		}
		if strings.Contains(out, "Sync complete") {
			t.Error("sync_complete event should be filtered out with --errors")
		}
	})

	t.Run("json output", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "Personal Todoist", 20, false, &buf, true)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "sources log" {
			t.Errorf("command = %q, want %q", env.Command, "sources log")
		}

		items, ok := env.Data.([]interface{})
		if !ok {
			t.Fatalf("data type = %T, want []interface{}", env.Data)
		}
		if len(items) != 4 {
			t.Errorf("len(data) = %d, want 4", len(items))
		}

		// Verify first item (most recent) has expected fields.
		first, ok := items[0].(map[string]interface{})
		if !ok {
			t.Fatalf("item type = %T, want map", items[0])
		}
		if first["connection"] != "Personal Todoist" {
			t.Errorf("connection = %v, want Personal Todoist", first["connection"])
		}
		if first["summary"] == nil || first["summary"] == "" {
			t.Error("missing summary in JSON event")
		}
	})

	t.Run("json output with errors filter", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "Personal Todoist", 20, true, &buf, true)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}

		items, ok := env.Data.([]interface{})
		if !ok {
			t.Fatalf("data type = %T, want []interface{}", env.Data)
		}
		if len(items) != 2 {
			t.Errorf("len(data) = %d, want 2 (1 error + 1 conflict)", len(items))
		}
	})

	t.Run("connection not found", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "Nonexistent", 20, false, &buf, false)
		if err == nil {
			t.Fatal("expected error for nonexistent connection")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error = %v, want contains 'not found'", err)
		}
	})

	t.Run("connection not found json", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "Nonexistent", 20, false, &buf, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Error == nil {
			t.Fatal("expected JSON error envelope")
		}
		if env.Error.Code != ExitNotFound {
			t.Errorf("error code = %d, want %d", env.Error.Code, ExitNotFound)
		}
	})

	t.Run("no events for connection", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		_, err := manager.Add("jira", "Empty Jira", map[string]string{})
		if err != nil {
			t.Fatalf("add connection: %v", err)
		}
		dir := t.TempDir()
		eventLog := connection.NewSyncEventLog(dir)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err = runSourcesLogTo(cmd, manager, eventLog, "Empty Jira", 20, false, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "No events found") {
			t.Errorf("expected 'No events found', got %q", out)
		}
	})

	t.Run("nil event log", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, nil, "", 20, false, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "No sync log available") {
			t.Errorf("expected 'No sync log available', got %q", out)
		}
	})

	t.Run("nil event log json", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, nil, "", 20, false, &buf, true)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		items, ok := env.Data.([]interface{})
		if !ok {
			t.Fatalf("data type = %T, want []interface{}", env.Data)
		}
		if len(items) != 0 {
			t.Errorf("len(data) = %d, want 0", len(items))
		}
	})

	t.Run("all connections with errors flag", func(t *testing.T) {
		t.Parallel()
		manager, eventLog := newTestLogManager(t)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		// No name = all connections
		err := runSourcesLogTo(cmd, manager, eventLog, "", 20, true, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "CONNECTION") {
			t.Error("missing table header CONNECTION for all-connections view")
		}
		if !strings.Contains(out, "Sync error") {
			t.Error("missing error event")
		}
		if !strings.Contains(out, "Conflict") {
			t.Error("missing conflict event")
		}
		if strings.Contains(out, "Sync complete") {
			t.Error("sync_complete should be filtered out with --errors")
		}
	})

	t.Run("no connections configured", func(t *testing.T) {
		t.Parallel()
		manager := connection.NewConnectionManager(nil)
		dir := t.TempDir()
		eventLog := connection.NewSyncEventLog(dir)

		var buf bytes.Buffer
		cmd := newSourcesLogCmd()
		err := runSourcesLogTo(cmd, manager, eventLog, "", 20, false, &buf, false)
		if err != nil {
			t.Fatalf("runSourcesLogTo: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "No connections configured") {
			t.Errorf("expected 'No connections configured', got %q", out)
		}
	})
}

func TestEventStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    connection.SyncEvent
		wantIcon string
	}{
		{"sync complete", connection.SyncEvent{Type: connection.EventSyncComplete}, "✓"},
		{"sync error", connection.SyncEvent{Type: connection.EventSyncError}, "✗"},
		{"conflict", connection.SyncEvent{Type: connection.EventConflict}, "⚠"},
		{"state change", connection.SyncEvent{Type: connection.EventStateChange}, "→"},
		{"sync start", connection.SyncEvent{Type: connection.EventSyncStart}, "▶"},
		{"reauth required", connection.SyncEvent{Type: connection.EventReauthRequired}, "✗"},
		{"unknown", connection.SyncEvent{Type: "unknown"}, "·"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := eventStatus(tt.event)
			if got != tt.wantIcon {
				t.Errorf("eventStatus() = %q, want %q", got, tt.wantIcon)
			}
		})
	}
}

func TestSortLabeledEvents(t *testing.T) {
	t.Parallel()

	events := []labeledEvent{
		{label: "a", event: connection.SyncEvent{Timestamp: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)}},
		{label: "b", event: connection.SyncEvent{Timestamp: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}},
		{label: "c", event: connection.SyncEvent{Timestamp: time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)}},
	}

	sortLabeledEvents(events)

	if events[0].label != "b" {
		t.Errorf("first event label = %q, want %q (most recent)", events[0].label, "b")
	}
	if events[1].label != "c" {
		t.Errorf("second event label = %q, want %q", events[1].label, "c")
	}
	if events[2].label != "a" {
		t.Errorf("third event label = %q, want %q (oldest)", events[2].label, "a")
	}
}
