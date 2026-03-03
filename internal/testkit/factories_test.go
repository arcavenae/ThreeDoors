package testkit

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestNewTask_Defaults(t *testing.T) {
	t.Parallel()
	task := NewTask("id-1", "Do something")

	if task.ID != "id-1" {
		t.Errorf("ID = %q, want %q", task.ID, "id-1")
	}
	if task.Text != "Do something" {
		t.Errorf("Text = %q, want %q", task.Text, "Do something")
	}
	if task.Status != core.StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusTodo)
	}
	if !task.CreatedAt.Equal(BaseTime) {
		t.Errorf("CreatedAt = %v, want %v", task.CreatedAt, BaseTime)
	}
}

func TestNewTask_WithOptions(t *testing.T) {
	t.Parallel()
	task := NewTask("id-2", "Blocked task",
		WithStatus(core.StatusBlocked),
		WithContext("waiting on review"),
	)

	if task.Status != core.StatusBlocked {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusBlocked)
	}
	if task.Context != "waiting on review" {
		t.Errorf("Context = %q, want %q", task.Context, "waiting on review")
	}
}

func TestNewTask_WithCompleted(t *testing.T) {
	t.Parallel()
	at := BaseTime.Add(2 * 3600_000_000_000) // +2h
	task := NewTask("id-3", "Done", WithCompleted(at))

	if task.Status != core.StatusComplete {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusComplete)
	}
	if task.CompletedAt == nil || !task.CompletedAt.Equal(at) {
		t.Errorf("CompletedAt = %v, want %v", task.CompletedAt, at)
	}
}

func TestNewTasks(t *testing.T) {
	t.Parallel()
	tasks := NewTasks(3)

	if len(tasks) != 3 {
		t.Fatalf("NewTasks(3) returned %d tasks, want 3", len(tasks))
	}

	// Verify sequential IDs
	for i, task := range tasks {
		wantID := idFromIndex(i)
		if task.ID != wantID {
			t.Errorf("tasks[%d].ID = %q, want %q", i, task.ID, wantID)
		}
	}
}
