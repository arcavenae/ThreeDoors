package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConflictLog_AppendAndRead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cl, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	entry := ConflictLogEntry{
		ConflictID:        "conflict-abc12345",
		Timestamp:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TaskID:            "task-1",
		DeviceIDs:         []string{"deviceA", "deviceB"},
		ResolutionOutcome: "auto-resolved",
		RejectedValues:    map[string]string{"text": "rejected value"},
	}

	if err := cl.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	entries, err := cl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ConflictID != "conflict-abc12345" {
		t.Errorf("expected conflict-abc12345, got %s", entries[0].ConflictID)
	}
	if entries[0].TaskID != "task-1" {
		t.Errorf("expected task-1, got %s", entries[0].TaskID)
	}
}

func TestConflictLog_CreatesSyncSubdir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	// Verify sync/ subdirectory was created
	syncDir := filepath.Join(dir, "sync")
	if !dirExists(syncDir) {
		t.Error("expected sync/ subdirectory to be created")
	}
}

func TestConflictLog_ReadRecentEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cl, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	// Write 5 entries
	for i := 0; i < 5; i++ {
		entry := ConflictLogEntry{
			ConflictID:        NewConflictID(),
			Timestamp:         time.Date(2026, 1, i+1, 0, 0, 0, 0, time.UTC),
			TaskID:            "task-1",
			ResolutionOutcome: "auto-resolved",
		}
		if err := cl.Append(entry); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}

	recent, err := cl.ReadRecentEntries(3)
	if err != nil {
		t.Fatalf("ReadRecentEntries: %v", err)
	}
	if len(recent) != 3 {
		t.Fatalf("expected 3 recent entries, got %d", len(recent))
	}
}

func TestConflictLog_FindByID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cl, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	targetID := "conflict-target99"
	for i := 0; i < 3; i++ {
		id := NewConflictID()
		if i == 1 {
			id = targetID
		}
		entry := ConflictLogEntry{
			ConflictID:        id,
			Timestamp:         time.Now().UTC(),
			TaskID:            "task-1",
			ResolutionOutcome: "auto-resolved",
		}
		if err := cl.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	found, err := cl.FindByID(targetID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.ConflictID != targetID {
		t.Errorf("expected %s, got %s", targetID, found.ConflictID)
	}

	_, err = cl.FindByID("nonexistent")
	if !errors.Is(err, ErrConflictNotFound) {
		t.Errorf("expected ErrConflictNotFound, got %v", err)
	}
}

func TestConflictLog_EntriesForTask(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cl, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	tasks := []string{"task-1", "task-2", "task-1", "task-3", "task-1"}
	for _, taskID := range tasks {
		entry := ConflictLogEntry{
			ConflictID:        NewConflictID(),
			Timestamp:         time.Now().UTC(),
			TaskID:            taskID,
			ResolutionOutcome: "auto-resolved",
		}
		if err := cl.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	filtered, err := cl.EntriesForTask("task-1")
	if err != nil {
		t.Fatalf("EntriesForTask: %v", err)
	}
	if len(filtered) != 3 {
		t.Errorf("expected 3 entries for task-1, got %d", len(filtered))
	}
}

func TestConflictLog_Rotation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cl, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	// Write entries until we exceed 1MB
	// Each entry is ~200 bytes of JSON, so ~5200 entries should exceed 1MB
	largeValue := strings.Repeat("x", 500)
	for i := 0; i < 5500; i++ {
		entry := ConflictLogEntry{
			ConflictID:        NewConflictID(),
			Timestamp:         time.Now().UTC(),
			TaskID:            "task-rotation",
			ResolutionOutcome: "auto-resolved",
			RejectedValues:    map[string]string{"text": largeValue},
		}
		if err := cl.Append(entry); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}

	entries, err := cl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries after rotation: %v", err)
	}

	// After rotation, should have roughly half the entries
	if len(entries) >= 5500 {
		t.Errorf("expected rotation to reduce entries, still have %d", len(entries))
	}
	if len(entries) == 0 {
		t.Error("expected some entries after rotation")
	}
}

func TestConflictLog_LogConflicts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cl, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	records := []ConflictRecord{
		{
			ConflictID: "conflict-12345678",
			Timestamp:  time.Now().UTC(),
			TaskID:     "task-1",
			DeviceIDs:  []string{"A", "B"},
			Fields: []FieldConflictDetail{
				{
					Field:       "text",
					LocalValue:  "local text",
					RemoteValue: "remote text",
					Winner:      "local",
					Reason:      "causal",
				},
			},
			ResolutionOutcome: "auto-resolved",
		},
	}

	if err := cl.LogConflicts(records); err != nil {
		t.Fatalf("LogConflicts: %v", err)
	}

	entries, err := cl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].RejectedValues["text"] != "remote text" {
		t.Errorf("expected rejected value 'remote text', got %q", entries[0].RejectedValues["text"])
	}
}

func TestConflictLog_EmptyLog(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cl, err := NewConflictLog(dir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	entries, err := cl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries on empty: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries on empty log, got %d", len(entries))
	}
}

// dirExists returns true if the path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
