package dispatch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuditEventType represents the type of dispatch audit event.
type AuditEventType string

const (
	AuditDispatch AuditEventType = "dispatch"
	AuditComplete AuditEventType = "complete"
	AuditFail     AuditEventType = "fail"
	AuditKill     AuditEventType = "kill"
)

// AuditEntry represents a single line in the dispatch audit log.
type AuditEntry struct {
	Timestamp   time.Time      `json:"timestamp"`
	EventType   AuditEventType `json:"event_type"`
	TaskID      string         `json:"task_id"`
	QueueItemID string         `json:"queue_item_id"`
	WorkerName  string         `json:"worker_name,omitempty"`
	PRNumber    int            `json:"pr_number,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// AuditLogger appends dispatch events to a JSONL file.
type AuditLogger struct {
	path string
}

// NewAuditLogger creates an AuditLogger writing to the given file path.
// The parent directory is created if it doesn't exist.
func NewAuditLogger(path string) (*AuditLogger, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create audit log directory: %w", err)
	}
	return &AuditLogger{path: path}, nil
}

// Log appends an audit entry as a single JSON line.
func (a *AuditLogger) Log(entry AuditEntry) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}

	f, err := os.OpenFile(a.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on append file

	_, err = f.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("write audit entry: %w", err)
	}
	return nil
}

// CountDispatchesToday returns the number of "dispatch" events logged
// on the current UTC day.
func (a *AuditLogger) CountDispatchesToday() (int, error) {
	return a.CountDispatchesSince(todayStartUTC())
}

// CountDispatchesSince returns the number of "dispatch" events logged
// at or after the given timestamp.
func (a *AuditLogger) CountDispatchesSince(since time.Time) (int, error) {
	entries, err := a.ReadAll()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, e := range entries {
		if e.EventType == AuditDispatch && !e.Timestamp.Before(since) {
			count++
		}
	}
	return count, nil
}

// LastDispatchForTask returns the timestamp of the most recent dispatch
// event for the given task ID. Returns zero time if none found.
func (a *AuditLogger) LastDispatchForTask(taskID string) (time.Time, error) {
	entries, err := a.ReadAll()
	if err != nil {
		return time.Time{}, err
	}

	var latest time.Time
	for _, e := range entries {
		if e.EventType == AuditDispatch && e.TaskID == taskID && e.Timestamp.After(latest) {
			latest = e.Timestamp
		}
	}
	return latest, nil
}

// ReadAll reads and parses all entries from the audit log.
// Returns an empty slice if the file does not exist.
func (a *AuditLogger) ReadAll() ([]AuditEntry, error) {
	data, err := os.ReadFile(a.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read audit log: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var entries []AuditEntry
	dec := json.NewDecoder(
		// Wrap bytes in a reader that json.NewDecoder can consume line-by-line.
		// We use bytes.NewReader via the stdlib approach.
		newByteReader(data),
	)

	for dec.More() {
		var entry AuditEntry
		if err := dec.Decode(&entry); err != nil {
			return entries, fmt.Errorf("decode audit entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// todayStartUTC returns the start of the current UTC day.
func todayStartUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// newByteReader returns a bytes.Reader for the given data.
// Extracted to avoid importing bytes in multiple places.
func newByteReader(data []byte) *byteReader {
	return &byteReader{data: data, pos: 0}
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
