package tasks

// Task represents a single task with text content.
// This is the simplified model for Story 1.2.
// The full model with UUID, status, notes, timestamps is for future stories.
type Task struct {
	Text string
}

// TaskLoader defines the interface for loading tasks.
// FileManager implements this. Tests can use stubs.
type TaskLoader interface {
	LoadTasks() ([]Task, error)
}
