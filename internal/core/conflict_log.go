package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const (
	conflictLogFile    = "conflicts.jsonl"
	conflictLogSubdir  = "sync"
	maxConflictLogSize = 1 * 1024 * 1024 // 1MB
)

// ErrConflictNotFound is returned when a conflict ID is not found in the log.
var ErrConflictNotFound = errors.New("conflict not found")

// ErrConflictAlreadyResolved is returned when attempting to manually resolve
// a conflict that has already been manually overridden.
var ErrConflictAlreadyResolved = errors.New("conflict already resolved")

// ConflictLogEntry represents a single conflict event in the log.
type ConflictLogEntry struct {
	ConflictID        string                `json:"conflict_id"`
	Timestamp         time.Time             `json:"timestamp"`
	TaskID            string                `json:"task_id"`
	DeviceIDs         []string              `json:"device_ids"`
	Fields            []FieldConflictDetail `json:"fields"`
	ResolutionOutcome string                `json:"resolution_outcome"` // "auto-resolved" or "manual-override"
	RejectedValues    map[string]string     `json:"rejected_values,omitempty"`
	OverrideOf        string                `json:"override_of,omitempty"` // conflict ID this overrides
}

// ConflictLog manages persistent conflict logging with rotation.
type ConflictLog struct {
	logPath string
}

// NewConflictLog creates a ConflictLog that writes to configDir/sync/conflicts.jsonl.
// Creates the sync/ subdirectory if it doesn't exist.
func NewConflictLog(configDir string) (*ConflictLog, error) {
	syncDir := filepath.Join(configDir, conflictLogSubdir)
	if err := os.MkdirAll(syncDir, 0o700); err != nil {
		return nil, fmt.Errorf("create sync dir: %w", err)
	}
	return &ConflictLog{
		logPath: filepath.Join(syncDir, conflictLogFile),
	}, nil
}

// NewConflictID generates a unique conflict identifier.
func NewConflictID() string {
	return fmt.Sprintf("conflict-%s", uuid.New().String()[:8])
}

// Append writes a new entry to the conflict log, rotating if needed.
func (cl *ConflictLog) Append(entry ConflictLogEntry) error {
	if err := cl.rotateIfNeeded(); err != nil {
		return fmt.Errorf("conflict log rotate: %w", err)
	}

	f, err := os.OpenFile(cl.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("conflict log open: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("conflict log marshal: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("conflict log write: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("conflict log sync: %w", err)
	}

	return nil
}

// LogConflicts writes conflict records from a merge operation to the log.
func (cl *ConflictLog) LogConflicts(records []ConflictRecord) error {
	for _, rec := range records {
		rejected := make(map[string]string)
		for _, f := range rec.Fields {
			if f.Winner == "local" {
				rejected[f.Field] = f.RemoteValue
			} else {
				rejected[f.Field] = f.LocalValue
			}
		}

		entry := ConflictLogEntry{
			ConflictID:        rec.ConflictID,
			Timestamp:         rec.Timestamp,
			TaskID:            rec.TaskID,
			DeviceIDs:         rec.DeviceIDs,
			Fields:            rec.Fields,
			ResolutionOutcome: rec.ResolutionOutcome,
			RejectedValues:    rejected,
		}
		if err := cl.Append(entry); err != nil {
			return err
		}
	}
	return nil
}

// ReadEntries reads all conflict log entries.
func (cl *ConflictLog) ReadEntries() ([]ConflictLogEntry, error) {
	f, err := os.Open(cl.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("conflict log open: %w", err)
	}
	defer func() { _ = f.Close() }()

	var entries []ConflictLogEntry
	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry ConflictLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // skip corrupt entries
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("conflict log scan: %w", err)
	}

	return entries, nil
}

// ReadRecentEntries returns the most recent N entries.
func (cl *ConflictLog) ReadRecentEntries(n int) ([]ConflictLogEntry, error) {
	entries, err := cl.ReadEntries()
	if err != nil {
		return nil, err
	}
	if len(entries) <= n {
		return entries, nil
	}
	return entries[len(entries)-n:], nil
}

// FindByID looks up a specific conflict entry by its conflict ID.
func (cl *ConflictLog) FindByID(conflictID string) (*ConflictLogEntry, error) {
	entries, err := cl.ReadEntries()
	if err != nil {
		return nil, err
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].ConflictID == conflictID {
			return &entries[i], nil
		}
	}
	return nil, ErrConflictNotFound
}

// EntriesSince returns all entries with timestamps at or after the given time.
func (cl *ConflictLog) EntriesSince(since time.Time) ([]ConflictLogEntry, error) {
	entries, err := cl.ReadEntries()
	if err != nil {
		return nil, err
	}
	var filtered []ConflictLogEntry
	for _, e := range entries {
		if !e.Timestamp.Before(since) {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

// EntriesForTask returns all entries for a specific task ID.
func (cl *ConflictLog) EntriesForTask(taskID string) ([]ConflictLogEntry, error) {
	entries, err := cl.ReadEntries()
	if err != nil {
		return nil, err
	}
	var filtered []ConflictLogEntry
	for _, e := range entries {
		if e.TaskID == taskID {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

// rotateIfNeeded checks if the log exceeds maxConflictLogSize and truncates.
func (cl *ConflictLog) rotateIfNeeded() error {
	info, err := os.Stat(cl.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat conflict log: %w", err)
	}

	if info.Size() < maxConflictLogSize {
		return nil
	}

	entries, err := cl.ReadEntries()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return os.Truncate(cl.logPath, 0)
	}

	keepCount := len(entries) / 2
	if keepCount == 0 {
		keepCount = 1
	}
	kept := entries[len(entries)-keepCount:]

	tmpPath := cl.logPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create conflict log temp: %w", err)
	}

	for _, entry := range kept {
		data, err := json.Marshal(entry)
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("encode conflict log entry: %w", err)
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			_ = f.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("write conflict log entry: %w", err)
		}
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync conflict log: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close conflict log temp: %w", err)
	}

	if err := os.Rename(tmpPath, cl.logPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename conflict log temp: %w", err)
	}

	return nil
}
