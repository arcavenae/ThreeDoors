package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncLog_AppendAndRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	entry := SyncLogEntry{
		Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Provider:  "Local",
		Operation: "sync",
		Added:     2,
		Updated:   1,
		Summary:   "Synced: 2 new, 1 updated",
	}

	if err := sl.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}

	got := entries[0]
	if got.Provider != "Local" {
		t.Errorf("Provider = %q, want %q", got.Provider, "Local")
	}
	if got.Operation != "sync" {
		t.Errorf("Operation = %q, want %q", got.Operation, "sync")
	}
	if got.Added != 2 {
		t.Errorf("Added = %d, want 2", got.Added)
	}
	if got.Summary != "Synced: 2 new, 1 updated" {
		t.Errorf("Summary = %q, want %q", got.Summary, "Synced: 2 new, 1 updated")
	}
}

func TestSyncLog_MultipleEntries(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	for i := 0; i < 5; i++ {
		entry := SyncLogEntry{
			Timestamp: time.Now().UTC(),
			Provider:  "Test",
			Operation: "sync",
			Summary:   "test entry",
		}
		if err := sl.Append(entry); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 5 {
		t.Fatalf("got %d entries, want 5", len(entries))
	}
}

func TestSyncLog_ReadRecentEntries(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	for i := 0; i < 10; i++ {
		entry := SyncLogEntry{
			Timestamp: time.Now().UTC(),
			Provider:  "Test",
			Operation: "sync",
			Added:     i,
			Summary:   "entry",
		}
		if err := sl.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	entries, err := sl.ReadRecentEntries(3)
	if err != nil {
		t.Fatalf("ReadRecentEntries: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	// Should be the last 3 entries
	if entries[0].Added != 7 {
		t.Errorf("first recent entry Added = %d, want 7", entries[0].Added)
	}
}

func TestSyncLog_ReadEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("got %d entries, want 0", len(entries))
	}
}

func TestSyncLog_LogSyncResult(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	result := SyncResult{
		Added:   3,
		Updated: 1,
		Removed: 0,
		Summary: "Synced: 3 new, 1 updated, 0 removed",
	}

	if err := sl.LogSyncResult("WAL", result); err != nil {
		t.Fatalf("LogSyncResult: %v", err)
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Provider != "WAL" {
		t.Errorf("Provider = %q, want %q", entries[0].Provider, "WAL")
	}
	if entries[0].Added != 3 {
		t.Errorf("Added = %d, want 3", entries[0].Added)
	}
}

func TestSyncLog_LogConflictResolution(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	if err := sl.LogConflictResolution("Local", "task-1", "Buy groceries", "local"); err != nil {
		t.Fatalf("LogConflictResolution: %v", err)
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Operation != "conflict_resolved" {
		t.Errorf("Operation = %q, want %q", entries[0].Operation, "conflict_resolved")
	}
	if entries[0].Resolution != "local" {
		t.Errorf("Resolution = %q, want %q", entries[0].Resolution, "local")
	}
}

func TestSyncLog_LogError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	if err := sl.LogError("Remote", fmt.Errorf("connection refused")); err != nil {
		t.Fatalf("LogError: %v", err)
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Operation != "error" {
		t.Errorf("Operation = %q, want %q", entries[0].Operation, "error")
	}
	if entries[0].Error != "connection refused" {
		t.Errorf("Error = %q, want %q", entries[0].Error, "connection refused")
	}
}

func TestSyncLog_Rotation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	// Write a large log file that exceeds 1MB
	logPath := filepath.Join(dir, syncLogFile)
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Write ~1.1MB of data (enough to trigger rotation)
	bigEntry := `{"timestamp":"2025-01-01T00:00:00Z","provider":"Test","operation":"sync","summary":"` +
		string(make([]byte, 1000)) + `"}` + "\n"
	for written := 0; written < maxSyncLogSize+1000; {
		n, wErr := f.WriteString(bigEntry)
		if wErr != nil {
			_ = f.Close()
			t.Fatalf("write: %v", wErr)
		}
		written += n
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Appending should trigger rotation
	entry := SyncLogEntry{
		Timestamp: time.Now().UTC(),
		Provider:  "Test",
		Operation: "sync",
		Summary:   "after rotation",
	}
	if err := sl.Append(entry); err != nil {
		t.Fatalf("Append after rotation: %v", err)
	}

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() >= maxSyncLogSize {
		t.Errorf("log size after rotation = %d, want < %d", info.Size(), maxSyncLogSize)
	}
}
