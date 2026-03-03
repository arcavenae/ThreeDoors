package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Shared test time constants for deterministic testing.
var (
	baseTime   = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	laterTime  = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	latestTime = time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)
)

// newTestTask creates a Task with explicit fields for deterministic testing.
// No UUID generation - caller controls the ID.
func newTestTask(id, text string, status TaskStatus, updatedAt time.Time) *Task {
	return &Task{
		ID:        id,
		Text:      text,
		Status:    status,
		Notes:     []TaskNote{},
		CreatedAt: updatedAt,
		UpdatedAt: updatedAt,
	}
}

// newCompletedTestTask creates a completed Task with CompletedAt set.
func newCompletedTestTask(id, text string, completedAt time.Time) *Task {
	return &Task{
		ID:          id,
		Text:        text,
		Status:      StatusComplete,
		Notes:       []TaskNote{},
		CreatedAt:   completedAt.Add(-1 * time.Hour),
		UpdatedAt:   completedAt,
		CompletedAt: &completedAt,
	}
}

// newTestSyncState creates a SyncState from a list of tasks.
// All task snapshots have Dirty=false.
func newTestSyncState(tasks ...*Task) SyncState {
	state := SyncState{
		LastSyncTime:  time.Now().UTC(),
		TaskSnapshots: make(map[string]TaskSnapshot),
	}
	for _, t := range tasks {
		state.TaskSnapshots[t.ID] = TaskSnapshot{
			ID:        t.ID,
			Text:      t.Text,
			Status:    t.Status,
			UpdatedAt: t.UpdatedAt,
			Dirty:     false,
		}
	}
	return state
}

// newDirtySyncState creates a SyncState where specified task IDs are marked Dirty.
func newDirtySyncState(dirtyIDs []string, tasks ...*Task) SyncState {
	state := newTestSyncState(tasks...)
	dirtySet := make(map[string]bool)
	for _, id := range dirtyIDs {
		dirtySet[id] = true
	}
	for id, snap := range state.TaskSnapshots {
		if dirtySet[id] {
			snap.Dirty = true
			state.TaskSnapshots[id] = snap
		}
	}
	return state
}

// poolFromTasks creates a TaskPool populated with the given tasks.
func poolFromTasks(tasks ...*Task) *TaskPool {
	pool := NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}
	return pool
}

// newCategorizedTestTask creates a Task with categorization fields for testing.
func newCategorizedTestTask(id, text string, status TaskStatus, taskType TaskType, effort TaskEffort, loc TaskLocation) *Task {
	t := newTestTask(id, text, status, baseTime)
	t.Type = taskType
	t.Effort = effort
	t.Location = loc
	return t
}

// tasksFile represents the YAML structure for test task persistence.
type tasksFile struct {
	Tasks []*Task `yaml:"tasks"`
}

// saveTestTasks writes tasks to tasks.yaml in the test config directory.
// This replaces textfile.SaveTasks in core tests to avoid import cycles.
func saveTestTasks(tasks []*Task) error {
	configPath, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	yamlPath := filepath.Join(configPath, "tasks.yaml")
	tf := tasksFile{Tasks: tasks}
	data, err := yaml.Marshal(&tf)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}
	return os.WriteFile(yamlPath, data, 0o644)
}

// inMemoryProvider is a simple in-memory TaskProvider for core tests.
type inMemoryProvider struct {
	tasks []*Task
}

func newInMemoryProvider() *inMemoryProvider {
	return &inMemoryProvider{}
}

func (p *inMemoryProvider) LoadTasks() ([]*Task, error) {
	return p.tasks, nil
}

func (p *inMemoryProvider) SaveTask(task *Task) error {
	for i, t := range p.tasks {
		if t.ID == task.ID {
			p.tasks[i] = task
			return nil
		}
	}
	p.tasks = append(p.tasks, task)
	return nil
}

func (p *inMemoryProvider) SaveTasks(tasks []*Task) error {
	p.tasks = tasks
	return nil
}

func (p *inMemoryProvider) DeleteTask(taskID string) error {
	for i, t := range p.tasks {
		if t.ID == taskID {
			p.tasks = append(p.tasks[:i], p.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task %q not found", taskID)
}

func (p *inMemoryProvider) MarkComplete(taskID string) error {
	for _, t := range p.tasks {
		if t.ID == taskID {
			t.Status = StatusComplete
			now := time.Now().UTC()
			t.CompletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("task %q not found", taskID)
}

// IsTextFileBackend identifies this as a textfile-like backend for health checker tests.
func (p *inMemoryProvider) IsTextFileBackend() bool {
	return true
}
