package textfile

import (
	"fmt"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// TextFileProvider wraps the existing file_manager.go functions
// to implement the TaskProvider interface.
type TextFileProvider struct{}

// NewTextFileProvider creates a new TextFileProvider.
func NewTextFileProvider() *TextFileProvider {
	return &TextFileProvider{}
}

// IsTextFileBackend implements the marker interface for health checker identification.
func (p *TextFileProvider) IsTextFileBackend() bool {
	return true
}

// LoadTasks reads tasks from the YAML file via file_manager.go.
func (p *TextFileProvider) LoadTasks() ([]*core.Task, error) {
	return LoadTasks()
}

// SaveTask saves a single task by loading all tasks, updating, and saving.
func (p *TextFileProvider) SaveTask(task *core.Task) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}
	found := false
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			found = true
			break
		}
	}
	if !found {
		tasks = append(tasks, task)
	}
	return SaveTasks(tasks)
}

// SaveTasks persists all tasks via file_manager.go.
func (p *TextFileProvider) SaveTasks(tasks []*core.Task) error {
	return SaveTasks(tasks)
}

// DeleteTask removes a task by ID.
func (p *TextFileProvider) DeleteTask(taskID string) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}
	filtered := make([]*core.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.ID != taskID {
			filtered = append(filtered, t)
		}
	}
	return SaveTasks(filtered)
}

// MarkComplete marks a task as complete, removes it from active tasks, and logs to completed.txt.
func (p *TextFileProvider) MarkComplete(taskID string) error {
	allTasks, err := LoadTasks()
	if err != nil {
		return fmt.Errorf("mark complete: load tasks: %w", err)
	}

	var target *core.Task
	for _, t := range allTasks {
		if t.ID == taskID {
			target = t
			break
		}
	}
	if target == nil {
		return fmt.Errorf("mark complete: task %q not found", taskID)
	}

	if !core.IsValidTransition(target.Status, core.StatusComplete) {
		return fmt.Errorf("mark complete: invalid transition from %s to complete", target.Status)
	}

	target.Status = core.StatusComplete
	now := time.Now().UTC()
	target.CompletedAt = &now

	filtered := make([]*core.Task, 0, len(allTasks))
	for _, t := range allTasks {
		if t.ID != taskID {
			filtered = append(filtered, t)
		}
	}
	if err := SaveTasks(filtered); err != nil {
		return fmt.Errorf("mark complete: save tasks: %w", err)
	}

	if err := AppendCompleted(target); err != nil {
		return fmt.Errorf("mark complete: append completed: %w", err)
	}

	return nil
}
