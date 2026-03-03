package testkit

import (
	"fmt"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// BaseTime is a fixed UTC timestamp for deterministic test data.
var BaseTime = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

// TaskOption configures a task created by NewTask.
type TaskOption func(*core.Task)

// WithStatus sets the task's status.
func WithStatus(s core.TaskStatus) TaskOption {
	return func(t *core.Task) {
		t.Status = s
	}
}

// WithContext sets the task's context field.
func WithContext(ctx string) TaskOption {
	return func(t *core.Task) {
		t.Context = ctx
	}
}

// WithType sets the task's type.
func WithType(tp core.TaskType) TaskOption {
	return func(t *core.Task) {
		t.Type = tp
	}
}

// WithEffort sets the task's effort level.
func WithEffort(e core.TaskEffort) TaskOption {
	return func(t *core.Task) {
		t.Effort = e
	}
}

// WithTimestamp sets both CreatedAt and UpdatedAt to the given time.
func WithTimestamp(ts time.Time) TaskOption {
	return func(t *core.Task) {
		t.CreatedAt = ts
		t.UpdatedAt = ts
	}
}

// WithCompleted marks the task as complete with a CompletedAt timestamp.
func WithCompleted(at time.Time) TaskOption {
	return func(t *core.Task) {
		t.Status = core.StatusComplete
		t.CompletedAt = &at
	}
}

// NewTask creates a deterministic test task with the given ID and text.
// Unlike core.NewTask, it uses a fixed timestamp and the caller-provided ID
// (no UUID generation), making tests reproducible.
func NewTask(id, text string, opts ...TaskOption) *core.Task {
	t := &core.Task{
		ID:        id,
		Text:      text,
		Status:    core.StatusTodo,
		Notes:     []core.TaskNote{},
		CreatedAt: BaseTime,
		UpdatedAt: BaseTime,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// NewTasks creates a slice of simple test tasks with sequential IDs.
// Tasks are named "Task 1", "Task 2", etc.
func NewTasks(count int) []*core.Task {
	tasks := make([]*core.Task, count)
	for i := range tasks {
		tasks[i] = NewTask(
			idFromIndex(i),
			nameFromIndex(i),
		)
	}
	return tasks
}

func idFromIndex(i int) string {
	return fmt.Sprintf("test-%03d", i+1)
}

func nameFromIndex(i int) string {
	return fmt.Sprintf("Task %d", i+1)
}
