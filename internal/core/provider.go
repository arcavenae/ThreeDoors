package core

import "errors"

// ErrReadOnly is a sentinel error for providers that do not support write operations.
// Used by FallbackProvider to detect when to delegate to the fallback.
var ErrReadOnly = errors.New("provider is read-only")

// ChangeType identifies the kind of modification in a ChangeEvent.
type ChangeType string

const (
	// ChangeCreated indicates a new task was added.
	ChangeCreated ChangeType = "created"
	// ChangeUpdated indicates an existing task was modified.
	ChangeUpdated ChangeType = "updated"
	// ChangeDeleted indicates a task was removed.
	ChangeDeleted ChangeType = "deleted"
)

// ChangeEvent represents a task modification detected by a provider's Watch channel.
// Adapters emit these events when external changes are detected (e.g., file watcher,
// sync from remote). The Source field identifies which adapter generated the event.
type ChangeEvent struct {
	Type   ChangeType // Created, Updated, or Deleted
	TaskID string     // ID of the affected task
	Task   *Task      // The task after the change (nil for deletes)
	Source string     // Name of the adapter that detected the change
}

// TaskProvider defines the interface for task storage backends.
// Implementations must provide CRUD operations, a human-readable name,
// health reporting, and an optional watch channel for external change detection.
//
// Core CRUD methods (LoadTasks, SaveTask, SaveTasks, DeleteTask, MarkComplete)
// handle task persistence. Name identifies the provider for logging and UI.
// HealthCheck reports provider-specific health status. Watch enables reactive
// updates when the underlying storage changes outside the application.
type TaskProvider interface {
	// Name returns a human-readable identifier for this provider (e.g., "textfile", "obsidian").
	Name() string

	// LoadTasks reads all active tasks from the storage backend.
	LoadTasks() ([]*Task, error)

	// SaveTask persists a single task, creating or updating as needed.
	SaveTask(task *Task) error

	// SaveTasks persists a batch of tasks, replacing all active tasks.
	SaveTasks(tasks []*Task) error

	// DeleteTask removes the task with the given ID from storage.
	DeleteTask(taskID string) error

	// MarkComplete marks a task as complete and archives it.
	// Returns ErrReadOnly if the provider does not support writes.
	MarkComplete(taskID string) error

	// Watch returns a read-only channel that emits ChangeEvents when external
	// modifications are detected. Providers that do not support watching return nil.
	// The caller must not close the returned channel.
	Watch() <-chan ChangeEvent

	// HealthCheck reports the operational status of this provider.
	// Returns a HealthCheckResult with individual check items.
	HealthCheck() HealthCheckResult
}
