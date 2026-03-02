package tasks

import (
	"fmt"
	"testing"
	"time"
)

// MockProvider implements TaskProvider for testing.
type MockProvider struct {
	Tasks      []*Task
	SavedTasks []*Task
	DeletedIDs []string
	LoadErr    error
	SaveErr    error
	DeleteErr  error
	LoadDelay  time.Duration
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
