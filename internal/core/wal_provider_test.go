package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// mockProvider is a test double for TaskProvider that can simulate failures.
type mockProvider struct {
	tasks          []*Task
	saveErr        error
	deleteErr      error
	completeErr    error
	loadErr        error
	saveCalls      int
	deleteCalls    int
	completeCalls  int
	saveBatchCalls int
}

func (m *mockProvider) LoadTasks() ([]*Task, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.tasks, nil
}

func (m *mockProvider) SaveTask(task *Task) error {
	m.saveCalls++
	if m.saveErr != nil {
		return m.saveErr
	}
	m.tasks = append(m.tasks, task)
	return nil
}

func (m *mockProvider) SaveTasks(tasks []*Task) error {
	m.saveBatchCalls++
	if m.saveErr != nil {
		return m.saveErr
	}
	m.tasks = tasks
	return nil
}

func (m *mockProvider) DeleteTask(taskID string) error {
	m.deleteCalls++
	if m.deleteErr != nil {
		return m.deleteErr
	}
	filtered := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		if t.ID != taskID {
			filtered = append(filtered, t)
		}
	}
	m.tasks = filtered
	return nil
}

func (m *mockProvider) MarkComplete(taskID string) error {
	m.completeCalls++
	return m.completeErr
}

func (m *mockProvider) Name() string              { return "mock" }
func (m *mockProvider) Watch() <-chan ChangeEvent { return nil }
func (m *mockProvider) HealthCheck() HealthCheckResult {
	return HealthCheckResult{}
}

func newTestWALProvider(t *testing.T) (*WALProvider, *mockProvider, string) {
	t.Helper()
	dir := t.TempDir()
	mock := &mockProvider{}
	wp := NewWALProvider(mock, dir)
	return wp, mock, dir
}

func TestWALProvider_SaveTask_PassesThrough(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)

	task := NewTask("test task")
	if err := wp.SaveTask(task); err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	if mock.saveCalls != 1 {
		t.Errorf("expected 1 save call, got %d", mock.saveCalls)
	}
	if wp.PendingCount() != 0 {
		t.Errorf("expected 0 pending entries, got %d", wp.PendingCount())
	}
}

func TestWALProvider_SaveTask_QueuesOnFailure(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	mock.saveErr = errors.New("provider unavailable")

	task := NewTask("test task")
	if err := wp.SaveTask(task); err != nil {
		t.Fatalf("SaveTask should not return error when queuing: %v", err)
	}

	if wp.PendingCount() != 1 {
		t.Errorf("expected 1 pending entry, got %d", wp.PendingCount())
	}
}

func TestWALProvider_DeleteTask_QueuesOnFailure(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	mock.deleteErr = errors.New("provider unavailable")

	if err := wp.DeleteTask("task-123"); err != nil {
		t.Fatalf("DeleteTask should not return error when queuing: %v", err)
	}

	if wp.PendingCount() != 1 {
		t.Errorf("expected 1 pending entry, got %d", wp.PendingCount())
	}
}

func TestWALProvider_MarkComplete_QueuesOnFailure(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	mock.completeErr = errors.New("provider unavailable")

	if err := wp.MarkComplete("task-123"); err != nil {
		t.Fatalf("MarkComplete should not return error when queuing: %v", err)
	}

	if wp.PendingCount() != 1 {
		t.Errorf("expected 1 pending entry, got %d", wp.PendingCount())
	}
}

func TestWALProvider_ReplayPending_Success(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)

	// Queue entries while provider is down
	mock.saveErr = errors.New("provider unavailable")
	task := NewTask("replay me")
	_ = wp.SaveTask(task)

	if wp.PendingCount() != 1 {
		t.Fatalf("expected 1 pending entry, got %d", wp.PendingCount())
	}

	// Restore provider
	mock.saveErr = nil
	errs := wp.ReplayPending()
	if len(errs) != 0 {
		t.Errorf("expected no errors on replay, got %v", errs)
	}

	if wp.PendingCount() != 0 {
		t.Errorf("expected 0 pending entries after replay, got %d", wp.PendingCount())
	}

	// save was called once for the initial attempt + once for replay
	if mock.saveCalls != 2 {
		t.Errorf("expected 2 save calls (initial + replay), got %d", mock.saveCalls)
	}
}

func TestWALProvider_ReplayPending_OrderPreserved(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)

	mock.saveErr = errors.New("provider unavailable")
	_ = wp.SaveTask(NewTask("first"))
	_ = wp.SaveTask(NewTask("second"))
	_ = wp.SaveTask(NewTask("third"))

	if wp.PendingCount() != 3 {
		t.Fatalf("expected 3 pending entries, got %d", wp.PendingCount())
	}

	// Replay and verify order by checking mock.tasks
	mock.saveErr = nil
	errs := wp.ReplayPending()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	if len(mock.tasks) != 3 {
		t.Fatalf("expected 3 tasks after replay, got %d", len(mock.tasks))
	}

	// Verify order matches queue order
	if mock.tasks[0].Text != "first" {
		t.Errorf("expected first task text 'first', got %q", mock.tasks[0].Text)
	}
	if mock.tasks[1].Text != "second" {
		t.Errorf("expected second task text 'second', got %q", mock.tasks[1].Text)
	}
	if mock.tasks[2].Text != "third" {
		t.Errorf("expected third task text 'third', got %q", mock.tasks[2].Text)
	}

	if wp.PendingCount() != 0 {
		t.Errorf("expected 0 pending, got %d", wp.PendingCount())
	}
}

func TestWALProvider_QueueSizeLimit(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	wp.SetMaxSize(3)
	mock.saveErr = errors.New("provider unavailable")

	for i := 0; i < 5; i++ {
		_ = wp.SaveTask(NewTask("task"))
	}

	if wp.PendingCount() != 3 {
		t.Errorf("expected 3 pending entries (max size), got %d", wp.PendingCount())
	}
}

func TestWALProvider_QueueEvictsOldest(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	wp.SetMaxSize(2)
	mock.saveErr = errors.New("provider unavailable")

	_ = wp.SaveTask(NewTask("oldest"))
	_ = wp.SaveTask(NewTask("middle"))
	_ = wp.SaveTask(NewTask("newest"))

	if wp.PendingCount() != 2 {
		t.Fatalf("expected 2 pending entries, got %d", wp.PendingCount())
	}

	// Replay to verify oldest was evicted
	mock.saveErr = nil
	wp.ReplayPending()

	if len(mock.tasks) != 2 {
		t.Fatalf("expected 2 replayed tasks, got %d", len(mock.tasks))
	}
	if mock.tasks[0].Text != "middle" {
		t.Errorf("expected 'middle', got %q", mock.tasks[0].Text)
	}
	if mock.tasks[1].Text != "newest" {
		t.Errorf("expected 'newest', got %q", mock.tasks[1].Text)
	}
}

func TestWALProvider_Persistence(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mock := &mockProvider{saveErr: errors.New("provider unavailable")}
	wp := NewWALProvider(mock, dir)

	task := NewTask("persisted task")
	_ = wp.SaveTask(task)

	// Verify WAL file exists
	walPath := filepath.Join(dir, walFile)
	if _, err := os.Stat(walPath); err != nil {
		t.Fatalf("WAL file should exist: %v", err)
	}

	// Create new WALProvider from same dir — should load pending entries
	mock2 := &mockProvider{}
	wp2 := NewWALProvider(mock2, dir)

	if wp2.PendingCount() != 1 {
		t.Errorf("expected 1 pending entry loaded from WAL file, got %d", wp2.PendingCount())
	}

	// Replay with working provider
	wp2.ReplayPending()

	if len(mock2.tasks) != 1 {
		t.Fatalf("expected 1 task replayed, got %d", len(mock2.tasks))
	}
	if mock2.tasks[0].Text != "persisted task" {
		t.Errorf("expected 'persisted task', got %q", mock2.tasks[0].Text)
	}
}

func TestWALProvider_LoadTasks_TriggersReplay(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)

	// Queue a save while provider is down
	mock.saveErr = errors.New("down")
	_ = wp.SaveTask(NewTask("queued"))

	// Restore provider and call LoadTasks
	mock.saveErr = nil
	tasks, err := wp.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks failed: %v", err)
		return
	}
	// LoadTasks returns what the inner provider has (initially empty from mock)
	_ = tasks

	// Replay should have been triggered by LoadTasks
	if wp.PendingCount() != 0 {
		t.Errorf("expected 0 pending after LoadTasks triggered replay, got %d", wp.PendingCount())
	}
}

func TestWALProvider_ReplayPending_Empty(t *testing.T) {
	t.Parallel()
	wp, _, _ := newTestWALProvider(t)

	errs := wp.ReplayPending()
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty replay, got %v", errs)
	}
}

func TestWALProvider_ReplayPending_RetryIncrement(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)

	// Queue entry
	mock.saveErr = errors.New("unavailable")
	_ = wp.SaveTask(NewTask("retry me"))

	// Replay with provider still down — should increment retry count
	errs := wp.ReplayPending()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if wp.PendingCount() != 1 {
		t.Errorf("expected 1 pending entry still queued, got %d", wp.PendingCount())
	}
}

func TestWALProvider_SaveTasks_QueuesOnFailure(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	mock.saveErr = errors.New("provider unavailable")

	tasks := []*Task{NewTask("a"), NewTask("b")}
	if err := wp.SaveTasks(tasks); err != nil {
		t.Fatalf("SaveTasks should not return error when queuing: %v", err)
	}

	if wp.PendingCount() != 1 {
		t.Errorf("expected 1 pending entry (batch), got %d", wp.PendingCount())
	}

	// Replay
	mock.saveErr = nil
	wp.ReplayPending()

	if mock.saveBatchCalls != 2 { // 1 initial + 1 replay
		t.Errorf("expected 2 save batch calls, got %d", mock.saveBatchCalls)
	}
}

func TestWALProvider_DeleteTask_PassesThrough(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	mock.tasks = []*Task{NewTask("to delete")}
	taskID := mock.tasks[0].ID

	if err := wp.DeleteTask(taskID); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	if mock.deleteCalls != 1 {
		t.Errorf("expected 1 delete call, got %d", mock.deleteCalls)
	}
	if wp.PendingCount() != 0 {
		t.Errorf("expected 0 pending, got %d", wp.PendingCount())
	}
}

func TestWALProvider_MixedOperations(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)

	// All operations fail
	mock.saveErr = errors.New("down")
	mock.deleteErr = errors.New("down")
	mock.completeErr = errors.New("down")

	_ = wp.SaveTask(NewTask("save"))
	_ = wp.DeleteTask("del-123")
	_ = wp.MarkComplete("comp-456")

	if wp.PendingCount() != 3 {
		t.Fatalf("expected 3 pending entries, got %d", wp.PendingCount())
	}

	// Restore and replay
	mock.saveErr = nil
	mock.deleteErr = nil
	mock.completeErr = nil

	errs := wp.ReplayPending()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if wp.PendingCount() != 0 {
		t.Errorf("expected 0 pending after replay, got %d", wp.PendingCount())
	}
}
