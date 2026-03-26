// Package testkit provides reusable test helpers for adapter and provider testing.
// It centralizes mock implementations and factory functions that were previously
// duplicated across test files, providing a single import for Epic 9 test expansion.
package testkit

import (
	"fmt"
	"sync"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// MockProvider is a configurable in-memory TaskProvider for testing.
// All operations are goroutine-safe. Error injection is supported via
// the Err* fields — set them to non-nil to simulate provider failures.
type MockProvider struct {
	mu    sync.Mutex
	tasks []*core.Task

	// Error injection: set these to force methods to return errors.
	ErrLoad         error
	ErrSave         error
	ErrSaveBatch    error
	ErrDelete       error
	ErrMarkComplete error

	// Call counters for verifying provider interactions.
	LoadCount         int
	SaveCount         int
	SaveBatchCount    int
	DeleteCount       int
	MarkCompleteCount int
}

// NewMockProvider returns a MockProvider with no tasks and no injected errors.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// NewMockProviderWithTasks returns a MockProvider seeded with the given tasks.
func NewMockProviderWithTasks(tasks []*core.Task) *MockProvider {
	cp := make([]*core.Task, len(tasks))
	copy(cp, tasks)
	return &MockProvider{tasks: cp}
}

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) LoadTasks() ([]*core.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LoadCount++
	if m.ErrLoad != nil {
		return nil, m.ErrLoad
	}
	cp := make([]*core.Task, len(m.tasks))
	copy(cp, m.tasks)
	return cp, nil
}

func (m *MockProvider) SaveTask(task *core.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SaveCount++
	if m.ErrSave != nil {
		return m.ErrSave
	}
	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks[i] = task
			return nil
		}
	}
	m.tasks = append(m.tasks, task)
	return nil
}

func (m *MockProvider) SaveTasks(tasks []*core.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SaveBatchCount++
	if m.ErrSaveBatch != nil {
		return m.ErrSaveBatch
	}
	cp := make([]*core.Task, len(tasks))
	copy(cp, tasks)
	m.tasks = cp
	return nil
}

func (m *MockProvider) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCount++
	if m.ErrDelete != nil {
		return m.ErrDelete
	}
	for i, t := range m.tasks {
		if t.ID == taskID {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task %q not found", taskID)
}

func (m *MockProvider) MarkComplete(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MarkCompleteCount++
	if m.ErrMarkComplete != nil {
		return m.ErrMarkComplete
	}
	for _, t := range m.tasks {
		if t.ID == taskID {
			t.Status = core.StatusComplete
			now := time.Now().UTC()
			t.CompletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("task %q not found", taskID)
}

func (m *MockProvider) Watch() <-chan core.ChangeEvent { return nil }

func (m *MockProvider) HealthCheck() core.HealthCheckResult {
	return core.HealthCheckResult{}
}

// Tasks returns a copy of the current task list (for test assertions).
func (m *MockProvider) Tasks() []*core.Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]*core.Task, len(m.tasks))
	copy(cp, m.tasks)
	return cp
}

// Reset clears all tasks and resets call counters.
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = nil
	m.LoadCount = 0
	m.SaveCount = 0
	m.SaveBatchCount = 0
	m.DeleteCount = 0
	m.MarkCompleteCount = 0
	m.ErrLoad = nil
	m.ErrSave = nil
	m.ErrSaveBatch = nil
	m.ErrDelete = nil
	m.ErrMarkComplete = nil
}
