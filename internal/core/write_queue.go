package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	pendingWritesFile = "pending_writes.yaml"
	maxRetries        = 5
)

// PendingWrite represents a failed write operation queued for retry.
type PendingWrite struct {
	TaskID    string    `yaml:"task_id"`
	Task      *Task     `yaml:"task,omitempty"`
	Operation string    `yaml:"operation"` // "save" or "delete"
	FailedAt  time.Time `yaml:"failed_at"`
	Retries   int       `yaml:"retries"`
	LastError string    `yaml:"last_error"`
}

// WriteQueue manages pending write operations with file persistence.
type WriteQueue struct {
	path    string
	pending []PendingWrite
}

// NewWriteQueue creates a WriteQueue, loading any persisted pending writes.
func NewWriteQueue(configDir string) *WriteQueue {
	q := &WriteQueue{
		path: filepath.Join(configDir, pendingWritesFile),
	}
	q.load()
	return q
}

// Enqueue adds a pending write and persists to disk.
func (q *WriteQueue) Enqueue(pw PendingWrite) error {
	q.pending = append(q.pending, pw)
	return q.save()
}

// Dequeue removes a pending write by task ID and persists.
func (q *WriteQueue) Dequeue(taskID string) error {
	filtered := make([]PendingWrite, 0, len(q.pending))
	for _, pw := range q.pending {
		if pw.TaskID != taskID {
			filtered = append(filtered, pw)
		}
	}
	q.pending = filtered
	return q.save()
}

// Pending returns all pending writes.
func (q *WriteQueue) Pending() []PendingWrite {
	result := make([]PendingWrite, len(q.pending))
	copy(result, q.pending)
	return result
}

// RetryAll attempts to replay all pending writes using the given provider.
// Returns errors for any writes that fail. Successful writes are dequeued.
// Writes that exceed maxRetries are removed with a log.
func (q *WriteQueue) RetryAll(provider TaskProvider) []error {
	if len(q.pending) == 0 {
		return nil
	}

	var retryErrors []error
	var remaining []PendingWrite

	for _, pw := range q.pending {
		var err error
		switch pw.Operation {
		case "save":
			if pw.Task != nil {
				err = provider.SaveTask(pw.Task)
			}
		case "delete":
			err = provider.DeleteTask(pw.TaskID)
		default:
			err = fmt.Errorf("unknown operation: %s", pw.Operation)
		}

		if err != nil {
			pw.Retries++
			pw.LastError = err.Error()

			if pw.Retries >= maxRetries {
				// Max retries exceeded — drop from queue
				fmt.Fprintf(os.Stderr, "Warning: dropping pending write for task %s after %d retries: %v\n", pw.TaskID, pw.Retries, err)
			} else {
				remaining = append(remaining, pw)
			}
			retryErrors = append(retryErrors, fmt.Errorf("retry failed for task %s: %w", pw.TaskID, err))
		}
		// Success: don't add to remaining (effectively dequeued)
	}

	q.pending = remaining
	_ = q.save() //nolint:errcheck // best-effort save after retry

	return retryErrors
}

// load reads pending writes from the persisted YAML file.
func (q *WriteQueue) load() {
	data, err := os.ReadFile(q.path)
	if err != nil {
		// File doesn't exist or unreadable — start empty
		q.pending = nil
		return
	}

	var loaded []PendingWrite
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		// Corrupt file — start empty (graceful recovery)
		fmt.Fprintf(os.Stderr, "Warning: corrupt pending_writes.yaml, starting empty: %v\n", err)
		q.pending = nil
		return
	}
	q.pending = loaded
}

// save persists pending writes to YAML using atomic write pattern.
func (q *WriteQueue) save() error {
	data, err := yaml.Marshal(q.pending)
	if err != nil {
		return fmt.Errorf("failed to marshal pending writes: %w", err)
	}

	tmpPath := q.path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, q.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}
