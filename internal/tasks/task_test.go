package tasks

import "testing"

func TestNewTask(t *testing.T) {
	task := NewTask("  Test task  ")
	if task.Text != "Test task" {
		t.Errorf("Expected trimmed text %q, got %q", "Test task", task.Text)
	}
	if task.ID == "" {
		t.Error("Expected non-empty UUID")
	}
	if task.Status != StatusTodo {
		t.Errorf("Expected status %q, got %q", StatusTodo, task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
	if task.CompletedAt != nil {
		t.Error("Expected nil CompletedAt")
	}
}

func TestTask_UpdateStatus(t *testing.T) {
	task := NewTask("Test")
	if err := task.UpdateStatus(StatusInProgress); err != nil {
		t.Fatalf("UpdateStatus to in-progress failed: %v", err)
	}
	if task.Status != StatusInProgress {
		t.Errorf("Expected %q, got %q", StatusInProgress, task.Status)
	}
}

func TestTask_UpdateStatus_Complete(t *testing.T) {
	task := NewTask("Test")
	if err := task.UpdateStatus(StatusComplete); err != nil {
		t.Fatalf("UpdateStatus to complete failed: %v", err)
	}
	if task.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set on complete")
	}
}

func TestTask_UpdateStatus_Invalid(t *testing.T) {
	task := NewTask("Test")
	task.Status = StatusComplete
	err := task.UpdateStatus(StatusTodo)
	if err == nil {
		t.Error("Expected error for invalid transition complete -> todo")
	}
}

func TestTask_AddNote(t *testing.T) {
	task := NewTask("Test")
	task.AddNote("  Progress update  ")
	if len(task.Notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(task.Notes))
	}
	if task.Notes[0].Text != "Progress update" {
		t.Errorf("Expected trimmed note text, got %q", task.Notes[0].Text)
	}
}

func TestTask_SetBlocker(t *testing.T) {
	task := NewTask("Test")
	_ = task.UpdateStatus(StatusBlocked)
	if err := task.SetBlocker("Waiting on API"); err != nil {
		t.Fatalf("SetBlocker failed: %v", err)
	}
	if task.Blocker != "Waiting on API" {
		t.Errorf("Expected blocker %q, got %q", "Waiting on API", task.Blocker)
	}
}

func TestTask_SetBlocker_WrongStatus(t *testing.T) {
	task := NewTask("Test")
	err := task.SetBlocker("should fail")
	if err == nil {
		t.Error("Expected error when setting blocker on non-blocked task")
	}
}

func TestTask_UpdateStatus_ClearBlocker(t *testing.T) {
	task := NewTask("Test")
	_ = task.UpdateStatus(StatusBlocked)
	_ = task.SetBlocker("blocker reason")
	_ = task.UpdateStatus(StatusInProgress)
	if task.Blocker != "" {
		t.Errorf("Expected blocker cleared after status change, got %q", task.Blocker)
	}
}

func TestNewTaskWithContext(t *testing.T) {
	task := NewTaskWithContext("  Buy groceries  ", "  Need healthy food for the week  ")
	if task.Text != "Buy groceries" {
		t.Errorf("Expected trimmed text %q, got %q", "Buy groceries", task.Text)
	}
	if task.Context != "Need healthy food for the week" {
		t.Errorf("Expected trimmed context %q, got %q", "Need healthy food for the week", task.Context)
	}
	if task.ID == "" {
		t.Error("Expected non-empty UUID")
	}
	if task.Status != StatusTodo {
		t.Errorf("Expected status %q, got %q", StatusTodo, task.Status)
	}
}

func TestNewTaskWithContext_EmptyContext(t *testing.T) {
	task := NewTaskWithContext("Buy groceries", "")
	if task.Context != "" {
		t.Errorf("Expected empty context, got %q", task.Context)
	}
	if task.Text != "Buy groceries" {
		t.Errorf("Expected text %q, got %q", "Buy groceries", task.Text)
	}
}

func TestNewTask_HasNoContext(t *testing.T) {
	task := NewTask("Simple task")
	if task.Context != "" {
		t.Errorf("Expected empty context for NewTask, got %q", task.Context)
	}
}

func TestTask_Validate(t *testing.T) {
	task := NewTask("Valid task")
	if err := task.Validate(); err != nil {
		t.Errorf("Expected valid task, got error: %v", err)
	}

	// Empty ID
	task2 := NewTask("test")
	task2.ID = ""
	if err := task2.Validate(); err == nil {
		t.Error("Expected error for empty ID")
	}

	// Empty text
	task3 := NewTask("test")
	task3.Text = ""
	if err := task3.Validate(); err == nil {
		t.Error("Expected error for empty text")
	}
}
