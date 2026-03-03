package core

import "errors"

// ErrReadOnly is a sentinel error for providers that do not support write operations.
// Used by FallbackProvider to detect when to delegate to the fallback.
var ErrReadOnly = errors.New("provider is read-only")

// TaskProvider defines the interface for task storage backends.
// Implementations include TextFileProvider and AppleNotesProvider.
type TaskProvider interface {
	LoadTasks() ([]*Task, error)
	SaveTask(task *Task) error
	SaveTasks(tasks []*Task) error
	DeleteTask(taskID string) error
	MarkComplete(taskID string) error
}
