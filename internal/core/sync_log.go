package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	syncLogFile    = "sync.log"
	maxSyncLogSize = 1 * 1024 * 1024 // 1MB
)

// SyncLogEntry represents a single sync operation log entry.
type SyncLogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Provider   string    `json:"provider"`
	Operation  string    `json:"operation"` // "sync", "conflict_resolved", "error"
	Added      int       `json:"added,omitempty"`
	Updated    int       `json:"updated,omitempty"`
	Removed    int       `json:"removed,omitempty"`
	Conflicts  int       `json:"conflicts,omitempty"`
	Resolution string    `json:"resolution,omitempty"` // "local", "remote", "both"
	TaskID     string    `json:"task_id,omitempty"`
	TaskText   string    `json:"task_text,omitempty"`
	Error      string    `json:"error,omitempty"`
	Summary    string    `json:"summary"`
}

// SyncLog manages persistent sync operation logging with rotation.
type SyncLog struct {
	logPath string
}

// NewSyncLog creates a SyncLog that writes to configDir/sync.log.
func NewSyncLog(configDir string) *SyncLog {
	return &SyncLog{
		logPath: filepath.Join(configDir, syncLogFile),
	}
}

// Append writes a new entry to the sync log, rotating if the file exceeds 1MB.
func (sl *SyncLog) Append(entry SyncLogEntry) error {
	if err := ValidatePath(sl.logPath); err != nil {
		return fmt.Errorf("sync log validation failed: %w", err)
	}

	if err := sl.rotateIfNeeded(); err != nil {
		return fmt.Errorf("sync log rotate: %w", err)
	}

	f, err := os.OpenFile(sl.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("sync log open: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("sync log marshal: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("sync log write: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync log sync: %w", err)
	}

	return nil
}

// LogSyncResult logs a completed sync operation.
func (sl *SyncLog) LogSyncResult(provider string, result SyncResult) error {
	entry := SyncLogEntry{
		Timestamp: time.Now().UTC(),
		Provider:  provider,
		Operation: "sync",
		Added:     result.Added,
		Updated:   result.Updated,
		Removed:   result.Removed,
		Conflicts: result.Conflicts,
		Summary:   result.Summary,
	}
	return sl.Append(entry)
}

// LogConflictResolution logs a user's conflict resolution choice.
func (sl *SyncLog) LogConflictResolution(provider string, taskID string, taskText string, resolution string) error {
	entry := SyncLogEntry{
		Timestamp:  time.Now().UTC(),
		Provider:   provider,
		Operation:  "conflict_resolved",
		TaskID:     taskID,
		TaskText:   taskText,
		Resolution: resolution,
		Summary:    fmt.Sprintf("Conflict on '%s' resolved: %s", taskText, resolution),
	}
	return sl.Append(entry)
}

// LogError logs a sync error.
func (sl *SyncLog) LogError(provider string, syncErr error) error {
	entry := SyncLogEntry{
		Timestamp: time.Now().UTC(),
		Provider:  provider,
		Operation: "error",
		Error:     syncErr.Error(),
		Summary:   fmt.Sprintf("Sync error: %s", syncErr.Error()),
	}
	return sl.Append(entry)
}

// ReadEntries reads all sync log entries from the log file.
// Returns entries in chronological order (oldest first).
func (sl *SyncLog) ReadEntries() ([]SyncLogEntry, error) {
	f, err := os.Open(sl.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("sync log open: %w", err)
	}
	defer func() { _ = f.Close() }()

	var entries []SyncLogEntry
	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry SyncLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // skip corrupt entries
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("sync log scan: %w", err)
	}

	return entries, nil
}

// ReadRecentEntries returns the most recent N entries.
func (sl *SyncLog) ReadRecentEntries(n int) ([]SyncLogEntry, error) {
	entries, err := sl.ReadEntries()
	if err != nil {
		return nil, err
	}
	if len(entries) <= n {
		return entries, nil
	}
	return entries[len(entries)-n:], nil
}

// EntriesSince returns all entries with timestamps at or after the given time.
func (sl *SyncLog) EntriesSince(since time.Time) ([]SyncLogEntry, error) {
	entries, err := sl.ReadEntries()
	if err != nil {
		return nil, err
	}
	var filtered []SyncLogEntry
	for _, e := range entries {
		if !e.Timestamp.Before(since) {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

// rotateIfNeeded checks if the log file exceeds maxSyncLogSize and truncates it.
func (sl *SyncLog) rotateIfNeeded() error {
	info, err := os.Stat(sl.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat sync log: %w", err)
	}

	if info.Size() < maxSyncLogSize {
		return nil
	}

	// Read all entries, keep the newest half
	entries, err := sl.ReadEntries()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		// All entries corrupt or empty — truncate the file
		return os.Truncate(sl.logPath, 0)
	}

	keepCount := len(entries) / 2
	if keepCount == 0 {
		keepCount = 1
	}
	kept := entries[len(entries)-keepCount:]

	// Atomic write
	tmpPath := sl.logPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create sync log temp: %w", err)
	}

	writer := bufio.NewWriter(f)
	encoder := json.NewEncoder(writer)
	for _, entry := range kept {
		if err := encoder.Encode(entry); err != nil {
			_ = f.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("encode sync log entry: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("flush sync log: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync log file sync: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close sync log temp: %w", err)
	}

	if err := os.Rename(tmpPath, sl.logPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename sync log temp: %w", err)
	}

	return nil
}
