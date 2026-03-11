package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	syncLogDir       = "sync-logs"
	maxEventsPerFile = 1000
)

// SyncEventType categorizes sync events.
type SyncEventType string

const (
	EventSyncComplete   SyncEventType = "sync_complete"
	EventSyncError      SyncEventType = "sync_error"
	EventConflict       SyncEventType = "conflict"
	EventStateChange    SyncEventType = "state_change"
	EventSyncStart      SyncEventType = "sync_start"
	EventReauthRequired SyncEventType = "reauth_required"
)

// SyncEvent represents a single sync-related event for a connection.
type SyncEvent struct {
	Timestamp    time.Time     `json:"timestamp"`
	ConnectionID string        `json:"connection_id"`
	Type         SyncEventType `json:"type"`

	// Task counts — populated for sync_complete events.
	Added   int `json:"added,omitempty"`
	Updated int `json:"updated,omitempty"`
	Removed int `json:"removed,omitempty"`

	// Conflict details — populated for conflict events.
	ConflictTaskID   string `json:"conflict_task_id,omitempty"`
	ConflictTaskText string `json:"conflict_task_text,omitempty"`
	Resolution       string `json:"resolution,omitempty"` // "local", "remote", "both"

	// State change details — populated for state_change events.
	FromState string `json:"from_state,omitempty"`
	ToState   string `json:"to_state,omitempty"`

	// Error details — populated for sync_error and reauth_required events.
	Error string `json:"error,omitempty"`

	// Human-readable summary.
	Summary string `json:"summary"`
}

// SyncEventLog manages per-connection JSONL sync event logging with rolling retention.
type SyncEventLog struct {
	baseDir string // directory containing per-connection log files
}

// NewSyncEventLog creates a SyncEventLog rooted at configDir/sync-logs/.
func NewSyncEventLog(configDir string) *SyncEventLog {
	return &SyncEventLog{
		baseDir: filepath.Join(configDir, syncLogDir),
	}
}

// logPath returns the JSONL file path for a given connection.
func (l *SyncEventLog) logPath(connectionID string) string {
	return filepath.Join(l.baseDir, connectionID+".jsonl")
}

// Append writes a sync event to the connection's log file,
// enforcing rolling retention of the last 1000 events.
func (l *SyncEventLog) Append(event SyncEvent) error {
	if event.ConnectionID == "" {
		return fmt.Errorf("append sync event: connection ID must not be empty")
	}

	if err := os.MkdirAll(l.baseDir, 0o755); err != nil {
		return fmt.Errorf("create sync-logs dir: %w", err)
	}

	path := l.logPath(event.ConnectionID)

	if err := l.truncateIfNeeded(path); err != nil {
		return fmt.Errorf("sync event truncate: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("sync event log open: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("sync event marshal: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("sync event write: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync event fsync: %w", err)
	}

	return nil
}

// LogSyncComplete logs a successful sync with task counts.
func (l *SyncEventLog) LogSyncComplete(connectionID string, added, updated, removed int) error {
	return l.Append(SyncEvent{
		Timestamp:    time.Now().UTC(),
		ConnectionID: connectionID,
		Type:         EventSyncComplete,
		Added:        added,
		Updated:      updated,
		Removed:      removed,
		Summary:      fmt.Sprintf("Sync complete: %d added, %d updated, %d removed", added, updated, removed),
	})
}

// LogSyncError logs a sync failure.
func (l *SyncEventLog) LogSyncError(connectionID string, syncErr error) error {
	return l.Append(SyncEvent{
		Timestamp:    time.Now().UTC(),
		ConnectionID: connectionID,
		Type:         EventSyncError,
		Error:        syncErr.Error(),
		Summary:      fmt.Sprintf("Sync error: %s", syncErr.Error()),
	})
}

// LogConflict logs a sync conflict with resolution details.
func (l *SyncEventLog) LogConflict(connectionID, taskID, taskText, resolution string) error {
	return l.Append(SyncEvent{
		Timestamp:        time.Now().UTC(),
		ConnectionID:     connectionID,
		Type:             EventConflict,
		ConflictTaskID:   taskID,
		ConflictTaskText: taskText,
		Resolution:       resolution,
		Summary:          fmt.Sprintf("Conflict on '%s' resolved: %s", taskText, resolution),
	})
}

// LogStateChange logs a connection state transition.
func (l *SyncEventLog) LogStateChange(connectionID string, from, to ConnectionState, errMsg string) error {
	event := SyncEvent{
		Timestamp:    time.Now().UTC(),
		ConnectionID: connectionID,
		Type:         EventStateChange,
		FromState:    from.String(),
		ToState:      to.String(),
		Summary:      fmt.Sprintf("State: %s → %s", from, to),
	}
	if errMsg != "" {
		event.Error = errMsg
	}
	return l.Append(event)
}

// SyncLog returns the most recent N events for a connection in reverse chronological order.
// If limit <= 0, all events are returned.
func (l *SyncEventLog) SyncLog(connectionID string, limit int) ([]SyncEvent, error) {
	events, err := l.readAll(connectionID)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(events) > limit {
		events = events[len(events)-limit:]
	}

	// Reverse to return most recent first.
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	return events, nil
}

// EventsSince returns events for a connection at or after the given time,
// in reverse chronological order.
func (l *SyncEventLog) EventsSince(connectionID string, since time.Time) ([]SyncEvent, error) {
	events, err := l.readAll(connectionID)
	if err != nil {
		return nil, err
	}

	var filtered []SyncEvent
	for _, e := range events {
		if !e.Timestamp.Before(since) {
			filtered = append(filtered, e)
		}
	}

	// Reverse to return most recent first.
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	return filtered, nil
}

// EventsByType returns events for a connection filtered by event type,
// in reverse chronological order. If limit <= 0, all matching events are returned.
func (l *SyncEventLog) EventsByType(connectionID string, eventType SyncEventType, limit int) ([]SyncEvent, error) {
	events, err := l.readAll(connectionID)
	if err != nil {
		return nil, err
	}

	var filtered []SyncEvent
	for _, e := range events {
		if e.Type == eventType {
			filtered = append(filtered, e)
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}

	// Reverse to return most recent first.
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	return filtered, nil
}

// readAll reads all sync events from a connection's log file in chronological order.
func (l *SyncEventLog) readAll(connectionID string) ([]SyncEvent, error) {
	path := l.logPath(connectionID)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("sync event log open %s: %w", connectionID, err)
	}
	defer func() { _ = f.Close() }()

	var events []SyncEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var event SyncEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue // skip corrupt entries
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return events, fmt.Errorf("sync event log scan %s: %w", connectionID, err)
	}

	return events, nil
}

// truncateIfNeeded trims the log file to the last maxEventsPerFile events
// using an atomic write (write to .tmp, sync, rename).
func (l *SyncEventLog) truncateIfNeeded(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat sync event log: %w", err)
	}

	var events []SyncEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var event SyncEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}
		events = append(events, event)
	}
	scanErr := scanner.Err()
	_ = f.Close()

	if scanErr != nil {
		return fmt.Errorf("scan sync event log: %w", scanErr)
	}

	if len(events) < maxEventsPerFile {
		return nil
	}

	// Keep the most recent maxEventsPerFile events.
	kept := events[len(events)-maxEventsPerFile:]

	tmpPath := path + ".tmp"
	tf, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create sync event temp: %w", err)
	}

	writer := bufio.NewWriter(tf)
	encoder := json.NewEncoder(writer)
	for _, event := range kept {
		if err := encoder.Encode(event); err != nil {
			_ = tf.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("encode sync event: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		_ = tf.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("flush sync event log: %w", err)
	}

	if err := tf.Sync(); err != nil {
		_ = tf.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync event log fsync: %w", err)
	}

	if err := tf.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close sync event temp: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename sync event temp: %w", err)
	}

	return nil
}
