package core

import (
	"errors"
	"fmt"
	"testing"
)

func TestFallbackProvider_ImplementsTaskProvider(t *testing.T) {
	var _ TaskProvider = (*FallbackProvider)(nil)
}

func TestFallbackProvider_PrimarySucceeds(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	primary := &MockProvider{Tasks: []*Task{taskA}}
	fallback := &MockProvider{Tasks: []*Task{}}

	fp := NewFallbackProvider(primary, fallback)

	tasks, err := fp.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "aaa" {
		t.Errorf("tasks[0].ID = %q, want %q", tasks[0].ID, "aaa")
	}
	if fp.IsFallback() {
		t.Error("IsFallback() = true, want false when primary succeeds")
	}
}

func TestFallbackProvider_PrimaryFails_UsesFallback(t *testing.T) {
	taskB := newTestTask("bbb", "Fallback Task", StatusTodo, baseTime)
	primary := &MockProvider{LoadErr: fmt.Errorf("apple notes: note not found")}
	fallback := &MockProvider{Tasks: []*Task{taskB}}

	fp := NewFallbackProvider(primary, fallback)

	tasks, err := fp.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("LoadTasks() returned %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "bbb" {
		t.Errorf("tasks[0].ID = %q, want %q", tasks[0].ID, "bbb")
	}
	if !fp.IsFallback() {
		t.Error("IsFallback() = false, want true when primary fails")
	}
	if fp.FallbackReason() == "" {
		t.Error("FallbackReason() is empty, want populated reason")
	}
}

func TestFallbackProvider_SaveDelegatesToFallbackWhenActive(t *testing.T) {
	primary := &MockProvider{LoadErr: fmt.Errorf("note not found")}
	fallback := &MockProvider{Tasks: []*Task{}}

	fp := NewFallbackProvider(primary, fallback)

	// Trigger fallback via LoadTasks
	_, _ = fp.LoadTasks()
	if !fp.IsFallback() {
		t.Fatal("expected fallback to be active after primary load failure")
	}

	// SaveTasks should delegate to fallback
	task := newTestTask("ccc", "New Task", StatusTodo, baseTime)
	err := fp.SaveTasks([]*Task{task})
	if err != nil {
		t.Fatalf("SaveTasks() unexpected error: %v", err)
	}
	if len(fallback.SavedTasks) != 1 {
		t.Errorf("fallback.SavedTasks has %d items, want 1", len(fallback.SavedTasks))
	}
}

func TestFallbackProvider_SaveTask_PrimaryReadOnly_DelegatesToFallback(t *testing.T) {
	// Simulate: primary loads OK but save returns ErrReadOnly
	primary := &MockProvider{
		Tasks:   []*Task{newTestTask("aaa", "Task A", StatusTodo, baseTime)},
		SaveErr: ErrReadOnly,
	}
	fallback := &MockProvider{Tasks: []*Task{}}

	fp := NewFallbackProvider(primary, fallback)

	// Load from primary (succeeds)
	_, err := fp.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}

	// SaveTask to primary returns ErrReadOnly — should delegate to fallback
	task := newTestTask("aaa", "Updated Task", StatusTodo, baseTime)
	err = fp.SaveTask(task)
	if err != nil {
		t.Fatalf("SaveTask() should succeed via fallback, got: %v", err)
	}
	if len(fallback.SavedTasks) != 1 {
		t.Errorf("fallback.SavedTasks has %d items, want 1", len(fallback.SavedTasks))
	}
}

func TestFallbackProvider_BothFail_ReturnsError(t *testing.T) {
	primary := &MockProvider{LoadErr: fmt.Errorf("primary failed")}
	fallback := &MockProvider{LoadErr: fmt.Errorf("fallback also failed")}

	fp := NewFallbackProvider(primary, fallback)

	_, err := fp.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error when both providers fail, got nil")
		return
	}
}

func TestFallbackProvider_DeleteTask_DelegatesToCorrectProvider(t *testing.T) {
	t.Run("primary active - delegates to primary", func(t *testing.T) {
		primary := &MockProvider{Tasks: []*Task{}}
		fallback := &MockProvider{Tasks: []*Task{}}

		fp := NewFallbackProvider(primary, fallback)
		_, _ = fp.LoadTasks() // primary succeeds

		err := fp.DeleteTask("aaa")
		if err != nil {
			t.Fatalf("DeleteTask() unexpected error: %v", err)
		}
		if len(primary.DeletedIDs) != 1 {
			t.Errorf("primary.DeletedIDs has %d items, want 1", len(primary.DeletedIDs))
		}
	})

	t.Run("primary failed - delegates to fallback", func(t *testing.T) {
		primary := &MockProvider{LoadErr: fmt.Errorf("failed")}
		fallback := &MockProvider{Tasks: []*Task{}}

		fp := NewFallbackProvider(primary, fallback)
		_, _ = fp.LoadTasks() // triggers fallback

		err := fp.DeleteTask("aaa")
		if err != nil {
			t.Fatalf("DeleteTask() unexpected error: %v", err)
		}
		if len(fallback.DeletedIDs) != 1 {
			t.Errorf("fallback.DeletedIDs has %d items, want 1", len(fallback.DeletedIDs))
		}
	})

	t.Run("primary returns ErrReadOnly - delegates to fallback", func(t *testing.T) {
		primary := &MockProvider{
			Tasks:     []*Task{},
			DeleteErr: ErrReadOnly,
		}
		fallback := &MockProvider{Tasks: []*Task{}}

		fp := NewFallbackProvider(primary, fallback)
		_, _ = fp.LoadTasks()

		err := fp.DeleteTask("aaa")
		if err != nil && !errors.Is(err, ErrReadOnly) {
			t.Fatalf("DeleteTask() unexpected error: %v", err)
		}
	})
}

func TestFallbackProvider_MarkComplete_DelegatesToCorrectProvider(t *testing.T) {
	t.Run("primary active - delegates to primary", func(t *testing.T) {
		primary := &MockProvider{Tasks: []*Task{}}
		fallback := &MockProvider{Tasks: []*Task{}}

		fp := NewFallbackProvider(primary, fallback)
		_, _ = fp.LoadTasks()

		err := fp.MarkComplete("aaa")
		if err != nil {
			t.Fatalf("MarkComplete() unexpected error: %v", err)
		}
		if len(primary.CompletedIDs) != 1 {
			t.Errorf("primary.CompletedIDs has %d items, want 1", len(primary.CompletedIDs))
		}
	})

	t.Run("primary failed - delegates to fallback", func(t *testing.T) {
		primary := &MockProvider{LoadErr: fmt.Errorf("failed")}
		fallback := &MockProvider{Tasks: []*Task{}}

		fp := NewFallbackProvider(primary, fallback)
		_, _ = fp.LoadTasks()

		err := fp.MarkComplete("aaa")
		if err != nil {
			t.Fatalf("MarkComplete() unexpected error: %v", err)
		}
		if len(fallback.CompletedIDs) != 1 {
			t.Errorf("fallback.CompletedIDs has %d items, want 1", len(fallback.CompletedIDs))
		}
	})

	t.Run("primary returns ErrReadOnly - delegates to fallback", func(t *testing.T) {
		primary := &MockProvider{
			Tasks:       []*Task{},
			CompleteErr: ErrReadOnly,
		}
		fallback := &MockProvider{Tasks: []*Task{}}

		fp := NewFallbackProvider(primary, fallback)
		_, _ = fp.LoadTasks()

		err := fp.MarkComplete("aaa")
		if err != nil {
			t.Fatalf("MarkComplete() should succeed via fallback, got: %v", err)
		}
		if len(fallback.CompletedIDs) != 1 {
			t.Errorf("fallback.CompletedIDs has %d items, want 1", len(fallback.CompletedIDs))
		}
	})
}
