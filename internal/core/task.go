package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/dispatch"
	"github.com/google/uuid"
)

// TaskNote captures a progress update added to a task.
type TaskNote struct {
	Timestamp time.Time `yaml:"timestamp" json:"timestamp"`
	Text      string    `yaml:"text" json:"text"`
}

// SourceRef links a task to its identity in a specific provider.
// A task may have multiple SourceRefs if it exists in several providers.
type SourceRef struct {
	Provider string `yaml:"provider" json:"provider"`
	NativeID string `yaml:"native_id" json:"native_id"`
}

// Task represents a single task with full lifecycle metadata.
type Task struct {
	ID             string                `yaml:"id" json:"id"`
	Text           string                `yaml:"text" json:"text"`
	Context        string                `yaml:"context,omitempty" json:"context,omitempty"`
	Status         TaskStatus            `yaml:"status" json:"status"`
	Type           TaskType              `yaml:"type,omitempty" json:"type,omitempty"`
	Effort         TaskEffort            `yaml:"effort,omitempty" json:"effort,omitempty"`
	Location       TaskLocation          `yaml:"location,omitempty" json:"location,omitempty"`
	Notes          []TaskNote            `yaml:"notes,omitempty" json:"notes,omitempty"`
	Blocker        string                `yaml:"blocker,omitempty" json:"blocker,omitempty"`
	CreatedAt      time.Time             `yaml:"created_at" json:"created_at"`
	UpdatedAt      time.Time             `yaml:"updated_at" json:"updated_at"`
	CompletedAt    *time.Time            `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
	SourceProvider string                `yaml:"source_provider,omitempty" json:"source_provider,omitempty"`
	SourceRefs     []SourceRef           `yaml:"source_refs,omitempty" json:"source_refs,omitempty"`
	DevDispatch    *dispatch.DevDispatch `yaml:"dev_dispatch,omitempty" json:"dev_dispatch,omitempty"`
}

// HasSourceRef returns true if the task has a SourceRef matching the given provider and native ID.
func (t *Task) HasSourceRef(provider, nativeID string) bool {
	for _, ref := range t.SourceRefs {
		if ref.Provider == provider && ref.NativeID == nativeID {
			return true
		}
	}
	return false
}

// AddSourceRef adds a SourceRef if one for the same provider and native ID doesn't already exist.
func (t *Task) AddSourceRef(provider, nativeID string) {
	if t.HasSourceRef(provider, nativeID) {
		return
	}
	t.SourceRefs = append(t.SourceRefs, SourceRef{
		Provider: provider,
		NativeID: nativeID,
	})
}

// EffectiveSourceProvider returns the source provider name, preferring SourceRefs[0]
// over the legacy SourceProvider field for backward compatibility.
func (t *Task) EffectiveSourceProvider() string {
	if len(t.SourceRefs) > 0 {
		return t.SourceRefs[0].Provider
	}
	return t.SourceProvider
}

// MigrateSourceProvider converts the legacy SourceProvider field to a SourceRef entry.
// This is a no-op if SourceRefs is already populated or SourceProvider is empty.
func (t *Task) MigrateSourceProvider() {
	if len(t.SourceRefs) > 0 || t.SourceProvider == "" {
		return
	}
	t.SourceRefs = []SourceRef{{
		Provider: t.SourceProvider,
		NativeID: t.ID,
	}}
}

// MigrateTasks runs MigrateSourceProvider on each task in the slice.
// This converts legacy SourceProvider fields to SourceRefs entries.
func MigrateTasks(tasks []*Task) {
	for _, t := range tasks {
		t.MigrateSourceProvider()
	}
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
	if t.CompletedAt != nil && newStatus != StatusComplete && newStatus != StatusArchived {
		t.CompletedAt = nil
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
