package dispatch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestAuditLogger(t *testing.T) *AuditLogger {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dev-dispatch.log")
	logger, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
		return nil
	}
	return logger
}

func TestAuditLoggerLog(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	entry := AuditEntry{
		Timestamp:   time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC),
		EventType:   AuditDispatch,
		TaskID:      "task-1",
		QueueItemID: "dq-abc12345",
		WorkerName:  "calm-tiger",
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log: %v", err)
	}

	data, err := os.ReadFile(logger.path)
	if err != nil {
		t.Fatalf("read log: %v", err)
		return
	}

	var got AuditEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.EventType != AuditDispatch {
		t.Errorf("EventType = %q, want %q", got.EventType, AuditDispatch)
	}
	if got.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want %q", got.TaskID, "task-1")
	}
	if got.WorkerName != "calm-tiger" {
		t.Errorf("WorkerName = %q, want %q", got.WorkerName, "calm-tiger")
	}
}

func TestAuditLoggerLogSetsTimestamp(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	before := time.Now().UTC()
	if err := logger.Log(AuditEntry{EventType: AuditDispatch, TaskID: "task-1"}); err != nil {
		t.Fatalf("Log: %v", err)
	}

	entries, err := logger.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
		return
	}
	if len(entries) != 1 {
		t.Fatalf("len = %d, want 1", len(entries))
	}
	if entries[0].Timestamp.Before(before) {
		t.Error("auto-set timestamp should be >= before")
	}
}

func TestAuditLoggerMultipleEntries(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	for _, et := range []AuditEventType{AuditDispatch, AuditComplete, AuditFail} {
		if err := logger.Log(AuditEntry{EventType: et, TaskID: "task-1"}); err != nil {
			t.Fatalf("Log %s: %v", et, err)
		}
	}

	entries, err := logger.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
		return
	}
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}
	if entries[0].EventType != AuditDispatch {
		t.Errorf("entries[0].EventType = %q, want %q", entries[0].EventType, AuditDispatch)
	}
	if entries[2].EventType != AuditFail {
		t.Errorf("entries[2].EventType = %q, want %q", entries[2].EventType, AuditFail)
	}
}

func TestAuditLoggerCountDispatchesSince(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	now := time.Now().UTC()
	old := now.Add(-2 * time.Hour)
	recent := now.Add(-30 * time.Minute)

	entries := []AuditEntry{
		{Timestamp: old, EventType: AuditDispatch, TaskID: "task-1"},
		{Timestamp: recent, EventType: AuditDispatch, TaskID: "task-2"},
		{Timestamp: recent, EventType: AuditComplete, TaskID: "task-1"},
		{Timestamp: now, EventType: AuditDispatch, TaskID: "task-3"},
	}
	for _, e := range entries {
		if err := logger.Log(e); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	count, err := logger.CountDispatchesSince(now.Add(-1 * time.Hour))
	if err != nil {
		t.Fatalf("CountDispatchesSince: %v", err)
		return
	}
	if count != 2 {
		t.Errorf("count = %d, want 2 (recent dispatch + now dispatch)", count)
	}
}

func TestAuditLoggerLastDispatchForTask(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	t1 := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)

	entries := []AuditEntry{
		{Timestamp: t1, EventType: AuditDispatch, TaskID: "task-1"},
		{Timestamp: t2, EventType: AuditDispatch, TaskID: "task-1"},
		{Timestamp: t2, EventType: AuditDispatch, TaskID: "task-2"},
	}
	for _, e := range entries {
		if err := logger.Log(e); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	last, err := logger.LastDispatchForTask("task-1")
	if err != nil {
		t.Fatalf("LastDispatchForTask: %v", err)
		return
	}
	if !last.Equal(t2) {
		t.Errorf("last = %v, want %v", last, t2)
	}
}

func TestAuditLoggerLastDispatchForTaskNotFound(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	last, err := logger.LastDispatchForTask("task-nonexistent")
	if err != nil {
		t.Fatalf("LastDispatchForTask: %v", err)
		return
	}
	if !last.IsZero() {
		t.Errorf("last = %v, want zero time", last)
	}
}

func TestAuditLoggerReadAllEmptyFile(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	entries, err := logger.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
		return
	}
	if len(entries) != 0 {
		t.Errorf("len = %d, want 0", len(entries))
	}
}

func TestAuditLoggerJSONLFormat(t *testing.T) {
	t.Parallel()
	logger := newTestAuditLogger(t)

	entry := AuditEntry{
		Timestamp:   time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC),
		EventType:   AuditFail,
		TaskID:      "task-1",
		QueueItemID: "dq-abc12345",
		Error:       "worker crashed",
	}
	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log: %v", err)
	}

	data, err := os.ReadFile(logger.path)
	if err != nil {
		t.Fatalf("read: %v", err)
		return
	}

	// Verify it's a single line ending with newline
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 1 {
		t.Errorf("expected 1 newline-terminated line, got %d newlines", lines)
	}

	var parsed AuditEntry
	if err := json.Unmarshal(data[:len(data)-1], &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Error != "worker crashed" {
		t.Errorf("Error = %q, want %q", parsed.Error, "worker crashed")
	}
}

func TestAuditEventTypeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		et   AuditEventType
		want string
	}{
		{"dispatch", AuditDispatch, "dispatch"},
		{"complete", AuditComplete, "complete"},
		{"fail", AuditFail, "fail"},
		{"kill", AuditKill, "kill"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.et) != tt.want {
				t.Errorf("AuditEventType = %q, want %q", tt.et, tt.want)
			}
		})
	}
}
