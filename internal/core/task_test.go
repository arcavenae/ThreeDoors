package core

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/dispatch"
)

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
	err := task.UpdateStatus(StatusInProgress)
	if err == nil {
		t.Error("Expected error for invalid transition complete -> in-progress")
	}
}

func TestTask_UpdateStatus_CompleteToTodo(t *testing.T) {
	task := NewTask("Test undo")
	task.AddNote("important note")

	if err := task.UpdateStatus(StatusComplete); err != nil {
		t.Fatalf("UpdateStatus to complete failed: %v", err)
	}
	if task.CompletedAt == nil {
		t.Fatal("Expected CompletedAt to be set after completion")
		return
	}

	if err := task.UpdateStatus(StatusTodo); err != nil {
		t.Fatalf("UpdateStatus complete->todo failed: %v", err)
	}
	if task.Status != StatusTodo {
		t.Errorf("Expected status %q, got %q", StatusTodo, task.Status)
	}
	if task.CompletedAt != nil {
		t.Error("Expected CompletedAt to be nil after undo")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set after undo")
	}
	if task.Blocker != "" {
		t.Errorf("Expected empty blocker after undo, got %q", task.Blocker)
	}
	if len(task.Notes) != 1 || task.Notes[0].Text != "important note" {
		t.Error("Expected notes to be preserved through undo")
	}
}

func TestTask_UpdateStatus_CompleteToInProgress_Invalid(t *testing.T) {
	task := NewTask("Test")
	_ = task.UpdateStatus(StatusComplete)
	if err := task.UpdateStatus(StatusInProgress); err == nil {
		t.Error("Expected error for complete -> in-progress")
	}
}

func TestTask_UpdateStatus_CompleteToBlocked_Invalid(t *testing.T) {
	task := NewTask("Test")
	_ = task.UpdateStatus(StatusComplete)
	if err := task.UpdateStatus(StatusBlocked); err == nil {
		t.Error("Expected error for complete -> blocked")
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

func TestForkTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Task
		check func(t *testing.T, original, forked *Task)
	}{
		{
			name: "preserves text",
			setup: func() *Task {
				return NewTask("Original task text")
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.Text != original.Text {
					t.Errorf("text: got %q, want %q", forked.Text, original.Text)
				}
			},
		},
		{
			name: "preserves context",
			setup: func() *Task {
				return NewTaskWithContext("Task", "Important context")
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.Context != "Important context" {
					t.Errorf("context: got %q, want %q", forked.Context, "Important context")
				}
			},
		},
		{
			name: "preserves effort",
			setup: func() *Task {
				task := NewTask("Task")
				task.Effort = EffortMedium
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.Effort != EffortMedium {
					t.Errorf("effort: got %q, want %q", forked.Effort, EffortMedium)
				}
			},
		},
		{
			name: "preserves type",
			setup: func() *Task {
				task := NewTask("Task")
				task.Type = TypeCreative
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.Type != TypeCreative {
					t.Errorf("type: got %q, want %q", forked.Type, TypeCreative)
				}
			},
		},
		{
			name: "preserves location",
			setup: func() *Task {
				task := NewTask("Task")
				task.Location = LocationHome
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.Location != LocationHome {
					t.Errorf("location: got %q, want %q", forked.Location, LocationHome)
				}
			},
		},
		{
			name: "resets status to todo",
			setup: func() *Task {
				task := NewTask("Task")
				_ = task.UpdateStatus(StatusInProgress)
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.Status != StatusTodo {
					t.Errorf("status: got %q, want %q", forked.Status, StatusTodo)
				}
			},
		},
		{
			name: "resets blocker to empty",
			setup: func() *Task {
				task := NewTask("Task")
				_ = task.UpdateStatus(StatusBlocked)
				_ = task.SetBlocker("something blocking")
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.Blocker != "" {
					t.Errorf("blocker: got %q, want empty", forked.Blocker)
				}
			},
		},
		{
			name: "has fresh timestamps",
			setup: func() *Task {
				return NewTask("Task")
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.CreatedAt.IsZero() {
					t.Error("expected non-zero CreatedAt")
				}
				if forked.CompletedAt != nil {
					t.Error("expected nil CompletedAt")
				}
			},
		},
		{
			name: "has new unique ID",
			setup: func() *Task {
				return NewTask("Task")
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.ID == original.ID {
					t.Error("forked task should have a different ID")
				}
				if forked.ID == "" {
					t.Error("forked task should have a non-empty ID")
				}
			},
		},
		{
			name: "adds provenance note",
			setup: func() *Task {
				return NewTask("Short task")
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if len(forked.Notes) != 1 {
					t.Fatalf("notes count: got %d, want 1", len(forked.Notes))
				}
				if forked.Notes[0].Text != "Forked from: Short task" {
					t.Errorf("note text: got %q, want %q", forked.Notes[0].Text, "Forked from: Short task")
				}
			},
		},
		{
			name: "truncates long text in provenance note at 60 chars",
			setup: func() *Task {
				return NewTask("This is a very long task description that exceeds sixty characters in total length")
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if len(forked.Notes) != 1 {
					t.Fatalf("notes count: got %d, want 1", len(forked.Notes))
				}
				// Text is 81 chars, truncated[:57] + "..."
				expected := "Forked from: This is a very long task description that exceeds sixty c..."
				if forked.Notes[0].Text != expected {
					t.Errorf("note text: got %q, want %q", forked.Notes[0].Text, expected)
				}
			},
		},
		{
			name: "does not copy DependsOn",
			setup: func() *Task {
				task := NewTask("Task")
				task.DependsOn = []string{"dep-1", "dep-2"}
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if len(forked.DependsOn) != 0 {
					t.Errorf("DependsOn: got %v, want empty", forked.DependsOn)
				}
			},
		},
		{
			name: "does not copy DevDispatch",
			setup: func() *Task {
				task := NewTask("Task")
				task.DevDispatch = &dispatch.DevDispatch{Queued: true}
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if forked.DevDispatch != nil {
					t.Error("DevDispatch: expected nil")
				}
			},
		},
		{
			name: "does not copy SourceRefs",
			setup: func() *Task {
				task := NewTask("Task")
				task.SourceRefs = []SourceRef{{Provider: "github", NativeID: "123"}}
				return task
			},
			check: func(t *testing.T, original, forked *Task) {
				t.Helper()
				if len(forked.SourceRefs) != 0 {
					t.Errorf("SourceRefs: got %v, want empty", forked.SourceRefs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			original := tt.setup()
			forked := ForkTask(original)
			tt.check(t, original, forked)
		})
	}
}

func TestForkTask_ExactTruncationBoundary(t *testing.T) {
	t.Parallel()

	// Exactly 60 chars — should NOT be truncated
	text60 := "123456789012345678901234567890123456789012345678901234567890"
	task60 := NewTask(text60)
	forked60 := ForkTask(task60)
	if forked60.Notes[0].Text != "Forked from: "+text60 {
		t.Errorf("60-char text should not be truncated, got %q", forked60.Notes[0].Text)
	}

	// 61 chars — should be truncated
	text61 := text60 + "x"
	task61 := NewTask(text61)
	forked61 := ForkTask(task61)
	expected := "Forked from: " + text60[:57] + "..."
	if forked61.Notes[0].Text != expected {
		t.Errorf("61-char text should be truncated, got %q, want %q", forked61.Notes[0].Text, expected)
	}
}

func TestForkTask_OriginalNotMutated(t *testing.T) {
	t.Parallel()
	original := NewTask("Original")
	original.Context = "my context"
	original.Effort = EffortQuickWin
	original.Type = TypePhysical
	origNotes := len(original.Notes)

	_ = ForkTask(original)

	if original.Context != "my context" {
		t.Errorf("original context mutated: %q", original.Context)
	}
	if original.Effort != EffortQuickWin {
		t.Errorf("original effort mutated: %q", original.Effort)
	}
	if original.Type != TypePhysical {
		t.Errorf("original type mutated: %q", original.Type)
	}
	if len(original.Notes) != origNotes {
		t.Errorf("original notes mutated: got %d, want %d", len(original.Notes), origNotes)
	}
}
