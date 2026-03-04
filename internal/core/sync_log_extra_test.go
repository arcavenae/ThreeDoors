package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSyncLog_Rotation_KeepsNewestHalf(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)
	logPath := filepath.Join(dir, syncLogFile)

	// Write a large file with identifiable entries
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Each entry is ~200 bytes. Write enough to exceed 1MB.
	padding := strings.Repeat("x", 150)
	entryCount := 0
	for written := 0; written < maxSyncLogSize+5000; {
		line := `{"timestamp":"2025-01-01T00:00:00Z","provider":"Test","operation":"sync","summary":"entry-` +
			padding + `","added":` + string(rune('0'+entryCount%10)) + `}` + "\n"
		n, wErr := f.WriteString(line)
		if wErr != nil {
			_ = f.Close()
			t.Fatalf("write: %v", wErr)
		}
		written += n
		entryCount++
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Trigger rotation via Append
	entry := SyncLogEntry{
		Timestamp: time.Now().UTC(),
		Provider:  "PostRotation",
		Operation: "sync",
		Summary:   "after-rotation",
	}
	if err := sl.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// File should be smaller than max
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() >= maxSyncLogSize {
		t.Errorf("expected size < %d after rotation, got %d", maxSyncLogSize, info.Size())
	}

	// Should contain the post-rotation entry
	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	lastEntry := entries[len(entries)-1]
	if lastEntry.Provider != "PostRotation" {
		t.Errorf("last entry provider = %q, want %q", lastEntry.Provider, "PostRotation")
	}
}

func TestSyncLog_Rotation_AllCorrupt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)
	logPath := filepath.Join(dir, syncLogFile)

	// Write corrupt data exceeding 1MB
	corrupt := strings.Repeat("not-json\n", maxSyncLogSize/9+100)
	if err := os.WriteFile(logPath, []byte(corrupt), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Append should trigger rotation and handle corrupt entries
	entry := SyncLogEntry{
		Timestamp: time.Now().UTC(),
		Provider:  "Test",
		Operation: "sync",
		Summary:   "after corrupt rotation",
	}
	if err := sl.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after corrupt rotation, got %d", len(entries))
	}
}

func TestSyncLog_Rotation_BelowThreshold(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	// Write a small file — should NOT trigger rotation
	entry := SyncLogEntry{
		Timestamp: time.Now().UTC(),
		Provider:  "Test",
		Operation: "sync",
		Summary:   "small entry",
	}
	for i := 0; i < 5; i++ {
		if err := sl.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries (no rotation), got %d", len(entries))
	}
}

func TestSyncLog_ReadEntries_SkipsCorrupt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)
	logPath := filepath.Join(dir, syncLogFile)

	// Write mix of valid and corrupt entries
	content := `{"timestamp":"2025-01-01T00:00:00Z","provider":"Valid1","operation":"sync","summary":"ok"}
not-valid-json
{"timestamp":"2025-01-02T00:00:00Z","provider":"Valid2","operation":"sync","summary":"also ok"}
`
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	entries, err := sl.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 valid entries, got %d", len(entries))
	}
}

func TestSyncLog_ReadRecentEntries_LessThanN(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sl := NewSyncLog(dir)

	entry := SyncLogEntry{
		Timestamp: time.Now().UTC(),
		Provider:  "Test",
		Operation: "sync",
		Summary:   "entry",
	}
	if err := sl.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	entries, err := sl.ReadRecentEntries(100)
	if err != nil {
		t.Fatalf("ReadRecentEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}
