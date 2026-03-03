package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaskNote captures a progress update added to a task.
type TaskNote struct {
	Timestamp time.Time `yaml:"timestamp" json:"timestamp"`
	Text      string    `yaml:"text" json:"text"`
}

// Task represents a single task with full lifecycle metadata.
type Task struct {
	ID             string       `yaml:"id" json:"id"`
	Text           string       `yaml:"text" json:"text"`
	Context        string       `yaml:"context,omitempty" json:"context,omitempty"`
	Status         TaskStatus   `yaml:"status" json:"status"`
	Type           TaskType     `yaml:"type,omitempty" json:"type,omitempty"`
	Effort         TaskEffort   `yaml:"effort,omitempty" json:"effort,omitempty"`
	Location       TaskLocation `yaml:"location,omitempty" json:"location,omitempty"`
	Notes          []TaskNote   `yaml:"notes,omitempty" json:"notes,omitempty"`
	Blocker        string       `yaml:"blocker,omitempty" json:"blocker,omitempty"`
	CreatedAt      time.Time    `yaml:"created_at" json:"created_at"`
	UpdatedAt      time.Time    `yaml:"updated_at" json:"updated_at"`
	CompletedAt    *time.Time   `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
	SourceProvider string       `yaml:"source_provider,omitempty" json:"source_provider,omitempty"`
}

// NewTask creates a new task with a UUID and default "todo" status.
func NewTask(text string) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:        uuid.New().String(),
		Text:      strings.TrimSpace(text),
		Status:    StatusTodo,
		Notes:     []TaskNote{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewTaskWithContext creates a new task with a UUID, default "todo" status, and context.
func NewTaskWithContext(text, context string) *Task {
	task := NewTask(text)
	task.Context = strings.TrimSpace(context)
	return task
}

// UpdateStatus changes the task's status after validating the transition.
func (t *Task) UpdateStatus(newStatus TaskStatus) error {
	if t.Status == newStatus {
		return nil // no-op for same status
	}
	if !IsValidTransition(t.Status, newStatus) {
		return fmt.Errorf("invalid transition from %q to %q", t.Status, newStatus)
	}
	now := time.Now().UTC()
	t.Status = newStatus
	t.UpdatedAt = now
	if newStatus == StatusComplete || newStatus == StatusArchived {
		t.CompletedAt = &now
	}
	if newStatus != StatusBlocked {
		t.Blocker = ""
	}
	return nil
}

// AddNote appends a progress note to the task.
func (t *Task) AddNote(text string) {
	t.Notes = append(t.Notes, TaskNote{
		Timestamp: time.Now().UTC(),
		Text:      strings.TrimSpace(text),
	})
	t.UpdatedAt = time.Now().UTC()
}

// SetBlocker sets the blocker description. Only valid when status is blocked.
func (t *Task) SetBlocker(reason string) error {
	if t.Status != StatusBlocked {
		return fmt.Errorf("cannot set blocker on task with status %q", t.Status)
	}
	t.Blocker = strings.TrimSpace(reason)
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Validate checks all fields for consistency.
func (t *Task) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID is required")
	}
	text := strings.TrimSpace(t.Text)
	if text == "" || len(text) > 500 {
		return fmt.Errorf("task text must be 1-500 characters")
	}
	if err := ValidateStatus(string(t.Status)); err != nil {
		return err
	}
	if t.CreatedAt.IsZero() {
		return fmt.Errorf("createdAt is required")
	}
	if !t.UpdatedAt.IsZero() && t.UpdatedAt.Before(t.CreatedAt) {
		return fmt.Errorf("updatedAt must be >= createdAt")
	}
	if t.CompletedAt != nil && t.Status != StatusComplete && t.Status != StatusArchived {
		return fmt.Errorf("completedAt should only be set when status is complete or archived")
	}
	if err := ValidateTaskType(t.Type); err != nil {
		return err
	}
	if err := ValidateTaskEffort(t.Effort); err != nil {
		return err
	}
	if err := ValidateTaskLocation(t.Location); err != nil {
		return err
	}
	return nil
}
