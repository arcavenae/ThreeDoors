package core

import "testing"

func TestStatusDeferred_Transitions(t *testing.T) {
	tests := []struct {
		name  string
		from  TaskStatus
		to    TaskStatus
		valid bool
	}{
		{"todo to deferred", StatusTodo, StatusDeferred, true},
		{"todo to archived", StatusTodo, StatusArchived, true},
		{"deferred to todo", StatusDeferred, StatusTodo, true},
		{"deferred to complete", StatusDeferred, StatusComplete, false},
		{"archived to todo", StatusArchived, StatusTodo, false},
		{"archived is terminal", StatusArchived, StatusInProgress, false},
		{"in-progress to deferred", StatusInProgress, StatusDeferred, false},
		{"blocked to deferred", StatusBlocked, StatusDeferred, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransition(tt.from, tt.to)
			if got != tt.valid {
				t.Errorf("IsValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.valid)
			}
		})
	}
}

func TestValidateStatus_NewStatuses(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"deferred", true},
		{"archived", true},
		{"invalid", false},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if tt.valid && err != nil {
				t.Errorf("ValidateStatus(%q) returned error: %v", tt.status, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("ValidateStatus(%q) expected error, got nil", tt.status)
			}
		})
	}
}

func TestTaskUpdateStatus_Deferred(t *testing.T) {
	task := NewTask("Test task")

	// Defer from todo
	if err := task.UpdateStatus(StatusDeferred); err != nil {
		t.Fatalf("Failed to defer task: %v", err)
	}
	if task.Status != StatusDeferred {
		t.Errorf("Expected status deferred, got %s", task.Status)
	}
	if task.CompletedAt != nil {
		t.Error("Deferred task should not have CompletedAt set")
	}

	// Un-defer back to todo
	if err := task.UpdateStatus(StatusTodo); err != nil {
		t.Fatalf("Failed to un-defer task: %v", err)
	}
	if task.Status != StatusTodo {
		t.Errorf("Expected status todo, got %s", task.Status)
	}
}

func TestTaskUpdateStatus_Archived(t *testing.T) {
	task := NewTask("Test task")

	if err := task.UpdateStatus(StatusArchived); err != nil {
		t.Fatalf("Failed to archive task: %v", err)
	}
	if task.Status != StatusArchived {
		t.Errorf("Expected status archived, got %s", task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("Archived task should have CompletedAt set")
	}

	// Cannot transition out of archived
	if err := task.UpdateStatus(StatusTodo); err == nil {
		t.Error("Expected error transitioning from archived to todo")
	}
}

func TestTaskValidate_ArchivedWithCompletedAt(t *testing.T) {
	task := NewTask("Test task")
	if err := task.UpdateStatus(StatusArchived); err != nil {
		t.Fatalf("Failed to archive: %v", err)
	}
	if err := task.Validate(); err != nil {
		t.Errorf("Archived task with CompletedAt should validate, got: %v", err)
	}
}
