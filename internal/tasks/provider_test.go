package tasks

import (
	"fmt"
	"testing"
	"time"
)

// MockProvider implements TaskProvider for testing.
type MockProvider struct {
	Tasks        []*Task
	SavedTasks   []*Task
	DeletedIDs   []string
	CompletedIDs []string
	LoadErr      error
	SaveErr      error
	DeleteErr    error
	CompleteErr  error
	LoadDelay    time.Duration
}

func (m *MockProvider) LoadTasks() ([]*Task, error) {
	if m.LoadDelay > 0 {
		time.Sleep(m.LoadDelay)
	}
	if m.LoadErr != nil {
		return nil, m.LoadErr
	}
	return m.Tasks, nil
}

func (m *MockProvider) SaveTask(task *Task) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.SavedTasks = append(m.SavedTasks, task)
	return nil
}

func (m *MockProvider) SaveTasks(tasks []*Task) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.SavedTasks = append(m.SavedTasks, tasks...)
	return nil
}

func (m *MockProvider) DeleteTask(taskID string) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	m.DeletedIDs = append(m.DeletedIDs, taskID)
	return nil
}

func (m *MockProvider) MarkComplete(taskID string) error {
	if m.CompleteErr != nil {
		return m.CompleteErr
	}
	m.CompletedIDs = append(m.CompletedIDs, taskID)
	return nil
}

// TestMockProvider_ImplementsTaskProvider verifies interface compliance.
func TestMockProvider_ImplementsTaskProvider(t *testing.T) {
	var _ TaskProvider = (*MockProvider)(nil)
}

// TestTextFileProvider_ImplementsTaskProvider verifies interface compliance.
func TestTextFileProvider_ImplementsTaskProvider(t *testing.T) {
	var _ TaskProvider = (*TextFileProvider)(nil)
}

func TestMockProvider_LoadTasks(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	tests := []struct {
		name      string
		provider  *MockProvider
		wantCount int
		wantErr   bool
	}{
		{
			name:      "returns configured tasks",
			provider:  &MockProvider{Tasks: []*Task{taskA, taskB}},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list",
			provider:  &MockProvider{Tasks: []*Task{}},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "returns error when configured",
			provider:  &MockProvider{LoadErr: fmt.Errorf("connection failed")},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := tt.provider.LoadTasks()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTasks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(tasks) != tt.wantCount {
				t.Errorf("LoadTasks() returned %d tasks, want %d", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestMockProvider_SaveTask(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)

	t.Run("saves task to SavedTasks slice", func(t *testing.T) {
		provider := &MockProvider{}
		err := provider.SaveTask(taskA)
		if err != nil {
			t.Fatalf("SaveTask() unexpected error: %v", err)
		}
		if len(provider.SavedTasks) != 1 {
			t.Errorf("SavedTasks has %d items, want 1", len(provider.SavedTasks))
		}
		if provider.SavedTasks[0].ID != "aaa" {
			t.Errorf("SavedTasks[0].ID = %q, want %q", provider.SavedTasks[0].ID, "aaa")
		}
	})

	t.Run("returns error when configured", func(t *testing.T) {
		provider := &MockProvider{SaveErr: fmt.Errorf("disk full")}
		err := provider.SaveTask(taskA)
		if err == nil {
			t.Error("SaveTask() expected error, got nil")
		}
	})
}

func TestMockProvider_DeleteTask(t *testing.T) {
	t.Run("records deleted ID", func(t *testing.T) {
		provider := &MockProvider{}
		err := provider.DeleteTask("aaa")
		if err != nil {
			t.Fatalf("DeleteTask() unexpected error: %v", err)
		}
		if len(provider.DeletedIDs) != 1 {
			t.Errorf("DeletedIDs has %d items, want 1", len(provider.DeletedIDs))
		}
		if provider.DeletedIDs[0] != "aaa" {
			t.Errorf("DeletedIDs[0] = %q, want %q", provider.DeletedIDs[0], "aaa")
		}
	})

	t.Run("returns error when configured", func(t *testing.T) {
		provider := &MockProvider{DeleteErr: fmt.Errorf("not found")}
		err := provider.DeleteTask("aaa")
		if err == nil {
			t.Error("DeleteTask() expected error, got nil")
		}
	})
}

func TestMockProvider_MarkComplete(t *testing.T) {
	t.Run("records completed ID", func(t *testing.T) {
		provider := &MockProvider{}
		err := provider.MarkComplete("aaa")
		if err != nil {
			t.Fatalf("MarkComplete() unexpected error: %v", err)
		}
		if len(provider.CompletedIDs) != 1 {
			t.Errorf("CompletedIDs has %d items, want 1", len(provider.CompletedIDs))
		}
		if provider.CompletedIDs[0] != "aaa" {
			t.Errorf("CompletedIDs[0] = %q, want %q", provider.CompletedIDs[0], "aaa")
		}
	})

	t.Run("returns error when configured", func(t *testing.T) {
		provider := &MockProvider{CompleteErr: fmt.Errorf("disk full")}
		err := provider.MarkComplete("aaa")
		if err == nil {
			t.Error("MarkComplete() expected error, got nil")
		}
	})
}

func TestTextFileProvider_MarkComplete_Success(t *testing.T) {
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	defer SetHomeDir("")

	// Create a task and save it
	task := NewTask("Complete me")
	if err := SaveTasks([]*Task{task}); err != nil {
		t.Fatalf("SaveTasks failed: %v", err)
	}

	provider := NewTextFileProvider()
	err := provider.MarkComplete(task.ID)
	if err != nil {
		t.Fatalf("MarkComplete() unexpected error: %v", err)
	}

	// Verify task removed from active tasks
	remaining, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() after complete: %v", err)
	}
	for _, r := range remaining {
		if r.ID == task.ID {
			t.Error("completed task should not be in active tasks")
		}
	}
}

func TestTextFileProvider_MarkComplete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	defer SetHomeDir("")

	task := NewTask("Only task")
	if err := SaveTasks([]*Task{task}); err != nil {
		t.Fatalf("SaveTasks failed: %v", err)
	}

	provider := NewTextFileProvider()
	err := provider.MarkComplete("nonexistent-id")
	if err == nil {
		t.Fatal("MarkComplete() expected error for nonexistent ID")
	}
}

func TestTextFileProvider_MarkComplete_InvalidTransition(t *testing.T) {
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	defer SetHomeDir("")

	// Complete status cannot transition to complete again
	completedAt := baseTime
	task := &Task{
		ID:          "completed-task",
		Text:        "Already done",
		Status:      StatusComplete,
		Notes:       []TaskNote{},
		CreatedAt:   baseTime,
		UpdatedAt:   baseTime,
		CompletedAt: &completedAt,
	}
	if err := SaveTasks([]*Task{task}); err != nil {
		t.Fatalf("SaveTasks failed: %v", err)
	}

	provider := NewTextFileProvider()
	err := provider.MarkComplete(task.ID)
	if err == nil {
		t.Fatal("MarkComplete() expected error for invalid transition from complete")
	}
}
