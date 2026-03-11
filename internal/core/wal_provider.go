package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	walFile            = "sync-queue.jsonl"
	defaultMaxWALSize  = 10000
	maxReplayRetries   = 10
	baseBackoffSeconds = 1
)

// WALOperation represents the type of write operation stored in the WAL.
type WALOperation string

const (
	WALOpSave         WALOperation = "save"
	WALOpSaveBatch    WALOperation = "save_batch"
	WALOpDelete       WALOperation = "delete"
	WALOpMarkComplete WALOperation = "mark_complete"
)

// WALEntry represents a single write-ahead log entry.
type WALEntry struct {
	Sequence  int64        `json:"seq"`
	Operation WALOperation `json:"op"`
	TaskID    string       `json:"task_id"`
	Task      *Task        `json:"task,omitempty"`
	Tasks     []*Task      `json:"tasks,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Retries   int          `json:"retries"`
	LastError string       `json:"last_error,omitempty"`
}

// WALProvider wraps a TaskProvider with a write-ahead log for offline-first operation.
// When the underlying provider is unavailable, writes are queued in a JSONL WAL file
// and replayed in order when connectivity is restored.
type WALProvider struct {
	inner    TaskProvider
	walPath  string
	maxSize  int
	mu       sync.Mutex
	pending  []WALEntry
	nextSeq  int64
	replying bool
}

// NewWALProvider creates a WALProvider wrapping the given provider.
// The WAL file is stored at configDir/sync-queue.jsonl.
func NewWALProvider(inner TaskProvider, configDir string) *WALProvider {
	wp := &WALProvider{
		inner:   inner,
		walPath: filepath.Join(configDir, walFile),
		maxSize: defaultMaxWALSize,
		nextSeq: 1,
	}
	wp.loadPending()
	return wp
}

// SetMaxSize sets the maximum number of WAL entries before oldest-first eviction.
func (wp *WALProvider) SetMaxSize(max int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.maxSize = max
}

// LoadTasks delegates directly to the inner provider.
// On successful load, attempts to replay any pending WAL entries.
func (wp *WALProvider) LoadTasks() ([]*Task, error) {
	tasks, err := wp.inner.LoadTasks()
	if err != nil {
		return nil, err
	}

	// Provider is available — try replaying pending entries
	wp.ReplayPending()

	return tasks, nil
}

// SaveTask attempts to save via the inner provider. If the provider is unavailable,
// the operation is queued in the WAL for later replay.
func (wp *WALProvider) SaveTask(task *Task) error {
	err := wp.inner.SaveTask(task)
	if err == nil {
		return nil
	}

	// Provider unavailable — queue in WAL
	entry := WALEntry{
		Operation: WALOpSave,
		TaskID:    task.ID,
		Task:      task,
		Timestamp: time.Now().UTC(),
	}
	return wp.enqueue(entry)
}

// SaveTasks attempts to save via the inner provider. If the provider is unavailable,
// the operation is queued in the WAL for later replay.
func (wp *WALProvider) SaveTasks(tasks []*Task) error {
	err := wp.inner.SaveTasks(tasks)
	if err == nil {
		return nil
	}

	// Provider unavailable — queue in WAL
	entry := WALEntry{
		Operation: WALOpSaveBatch,
		Timestamp: time.Now().UTC(),
		Tasks:     tasks,
	}
	return wp.enqueue(entry)
}

// DeleteTask attempts to delete via the inner provider. If the provider is unavailable,
// the operation is queued in the WAL for later replay.
func (wp *WALProvider) DeleteTask(taskID string) error {
	err := wp.inner.DeleteTask(taskID)
	if err == nil {
		return nil
	}

	entry := WALEntry{
		Operation: WALOpDelete,
		TaskID:    taskID,
		Timestamp: time.Now().UTC(),
	}
	return wp.enqueue(entry)
}

// MarkComplete attempts to mark complete via the inner provider. If the provider is
// unavailable, the operation is queued in the WAL for later replay.
func (wp *WALProvider) MarkComplete(taskID string) error {
	err := wp.inner.MarkComplete(taskID)
	if err == nil {
		return nil
	}

	entry := WALEntry{
		Operation: WALOpMarkComplete,
		TaskID:    taskID,
		Timestamp: time.Now().UTC(),
	}
	return wp.enqueue(entry)
}

// Name returns the name of the inner provider with a WAL suffix.
func (wp *WALProvider) Name() string {
	return wp.inner.Name() + " (WAL)"
}

// Watch delegates to the inner provider's Watch channel.
func (wp *WALProvider) Watch() <-chan ChangeEvent {
	return wp.inner.Watch()
}

// HealthCheck delegates to the inner provider's HealthCheck.
func (wp *WALProvider) HealthCheck() HealthCheckResult {
	return wp.inner.HealthCheck()
}

// PendingCount returns the number of pending WAL entries.
func (wp *WALProvider) PendingCount() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return len(wp.pending)
}

// OldestPending returns the timestamp of the oldest pending WAL entry.
// Returns the zero time if there are no pending entries.
func (wp *WALProvider) OldestPending() time.Time {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if len(wp.pending) == 0 {
		return time.Time{}
	}
	return wp.pending[0].Timestamp
}

// ReplayPending attempts to replay all pending WAL entries against the inner provider.
// Entries that succeed are removed. Entries that fail are kept with incremented retry
// counts and exponential backoff is applied. Entries exceeding maxReplayRetries are dropped.
func (wp *WALProvider) ReplayPending() []error {
	wp.mu.Lock()
	if wp.replying || len(wp.pending) == 0 {
		wp.mu.Unlock()
		return nil
	}
	wp.replying = true
	entries := make([]WALEntry, len(wp.pending))
	copy(entries, wp.pending)
	wp.mu.Unlock()

	var replayErrors []error
	var remaining []WALEntry

	for _, entry := range entries {
		// Check if backoff period has elapsed
		if entry.Retries > 0 {
			backoff := time.Duration(math.Pow(2, float64(entry.Retries-1))) * time.Second * baseBackoffSeconds
			if time.Since(entry.Timestamp.Add(backoff)) < 0 {
				// Not yet time to retry — keep in queue
				remaining = append(remaining, entry)
				continue
			}
		}

		err := wp.replayEntry(entry)
		if err != nil {
			entry.Retries++
			entry.LastError = err.Error()

			if entry.Retries >= maxReplayRetries {
				fmt.Fprintf(os.Stderr, "Warning: dropping WAL entry seq=%d op=%s task=%s after %d retries: %v\n",
					entry.Sequence, entry.Operation, entry.TaskID, entry.Retries, err)
			} else {
				remaining = append(remaining, entry)
			}
			replayErrors = append(replayErrors, fmt.Errorf("replay WAL entry seq=%d: %w", entry.Sequence, err))
		}
	}

	wp.mu.Lock()
	wp.pending = remaining
	wp.replying = false
	wp.mu.Unlock()

	if err := wp.persistWAL(); err != nil {
		replayErrors = append(replayErrors, fmt.Errorf("persist WAL after replay: %w", err))
	}

	return replayErrors
}

func (wp *WALProvider) replayEntry(entry WALEntry) error {
	switch entry.Operation {
	case WALOpSave:
		if entry.Task == nil {
			return fmt.Errorf("save entry missing task data")
		}
		return wp.inner.SaveTask(entry.Task)
	case WALOpSaveBatch:
		if len(entry.Tasks) == 0 {
			return fmt.Errorf("save_batch entry missing tasks data")
		}
		return wp.inner.SaveTasks(entry.Tasks)
	case WALOpDelete:
		return wp.inner.DeleteTask(entry.TaskID)
	case WALOpMarkComplete:
		return wp.inner.MarkComplete(entry.TaskID)
	default:
		return fmt.Errorf("unknown WAL operation: %s", entry.Operation)
	}
}

func (wp *WALProvider) enqueue(entry WALEntry) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	entry.Sequence = wp.nextSeq
	wp.nextSeq++

	wp.pending = append(wp.pending, entry)

	// Enforce size limit with oldest-first eviction
	if len(wp.pending) > wp.maxSize {
		evictCount := len(wp.pending) - wp.maxSize
		fmt.Fprintf(os.Stderr, "Warning: WAL queue exceeded max size %d, evicting %d oldest entries\n",
			wp.maxSize, evictCount)
		wp.pending = wp.pending[evictCount:]
	}

	return wp.persistWALLocked()
}

// persistWAL writes the current pending entries to the WAL file using atomic write.
func (wp *WALProvider) persistWAL() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.persistWALLocked()
}

// persistWALLocked writes the WAL file. Caller must hold wp.mu.
func (wp *WALProvider) persistWALLocked() error {
	tmpPath := wp.walPath + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create WAL temp file: %w", err)
	}

	writer := bufio.NewWriter(f)
	encoder := json.NewEncoder(writer)
	for _, entry := range wp.pending {
		if err := encoder.Encode(entry); err != nil {
			_ = f.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("encode WAL entry seq=%d: %w", entry.Sequence, err)
		}
	}

	if err := writer.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("flush WAL temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync WAL temp file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close WAL temp file: %w", err)
	}

	if err := os.Rename(tmpPath, wp.walPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename WAL temp file: %w", err)
	}

	return nil
}

// loadPending reads existing WAL entries from the JSONL file.
func (wp *WALProvider) loadPending() {
	f, err := os.Open(wp.walPath)
	if err != nil {
		return
	}
	defer f.Close() //nolint:errcheck // best-effort close on read

	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	var maxSeq int64
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry WALEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping corrupt WAL entry: %v\n", err)
			continue
		}
		wp.pending = append(wp.pending, entry)
		if entry.Sequence > maxSeq {
			maxSeq = entry.Sequence
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error reading WAL file: %v\n", err)
	}

	wp.nextSeq = maxSeq + 1
}
