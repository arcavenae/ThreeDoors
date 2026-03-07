package cli

import (
	"encoding/json"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestCompleteOneTask_NotFound(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	ctx := &cliContext{pool: pool}

	result := completeOneTask(ctx, "nonexistent")

	if result.Success {
		t.Error("expected failure for nonexistent task")
	}
	if result.ExitCode != ExitNotFound {
		t.Errorf("exit code = %d, want %d", result.ExitCode, ExitNotFound)
	}
	if result.Error != "task not found" {
		t.Errorf("error = %q, want %q", result.Error, "task not found")
	}
}

func TestCompleteOneTask_AmbiguousPrefix(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task1 := core.NewTask("Task one")
	task1.ID = "abc-111"
	task2 := core.NewTask("Task two")
	task2.ID = "abc-222"
	pool.AddTask(task1)
	pool.AddTask(task2)

	ctx := &cliContext{pool: pool}

	result := completeOneTask(ctx, "abc")

	if result.Success {
		t.Error("expected failure for ambiguous prefix")
	}
	if result.ExitCode != ExitAmbiguousInput {
		t.Errorf("exit code = %d, want %d", result.ExitCode, ExitAmbiguousInput)
	}
}

func TestCompleteOneTask_InvalidTransition(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	// Deferred tasks cannot transition directly to complete
	task := core.NewTask("Deferred task")
	task.ID = "deferred-task-id"
	_ = task.UpdateStatus(core.StatusDeferred)
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := completeOneTask(ctx, "deferred-task-id")

	if result.Success {
		t.Error("expected failure for invalid transition")
	}
	if result.ExitCode != ExitValidation {
		t.Errorf("exit code = %d, want %d", result.ExitCode, ExitValidation)
	}
}

func TestCompleteOneTask_AlreadyComplete(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Already done")
	task.ID = "done-task-id"
	_ = task.UpdateStatus(core.StatusComplete)
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := completeOneTask(ctx, "done-task-id")

	// Completing an already-complete task is a no-op (same status returns nil)
	if !result.Success {
		t.Errorf("expected success for already-complete task, got error: %s", result.Error)
	}
}

func TestCompleteOneTask_Success(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Do something")
	task.ID = "unique-task-id"
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := completeOneTask(ctx, "unique")

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("exit code = %d, want %d", result.ExitCode, ExitSuccess)
	}
	if task.Status != core.StatusComplete {
		t.Errorf("task status = %q, want %q", task.Status, core.StatusComplete)
	}
	if task.CompletedAt == nil {
		t.Error("task CompletedAt should be set after completion")
	}
}

func TestCompleteOneTask_PrefixMatch(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Prefix match test")
	task.ID = "abcdef12-3456-7890-abcd-ef1234567890"
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := completeOneTask(ctx, "abcdef12")

	if !result.Success {
		t.Errorf("expected success with prefix match, got error: %s", result.Error)
	}
	if result.ID != task.ID {
		t.Errorf("result ID = %q, want %q", result.ID, task.ID)
	}
}

func TestCompleteResult_JSON(t *testing.T) {
	t.Parallel()

	r := completeResult{
		ID:       "abc-123",
		ShortID:  "abc-123",
		Success:  true,
		ExitCode: 0,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["id"] != "abc-123" {
		t.Errorf("id = %v, want abc-123", decoded["id"])
	}
	if decoded["success"] != true {
		t.Errorf("success = %v, want true", decoded["success"])
	}
	if _, ok := decoded["error"]; ok {
		t.Error("error field should be omitted when empty")
	}
}

func TestTaskAddCreation(t *testing.T) {
	t.Parallel()

	t.Run("basic task", func(t *testing.T) {
		t.Parallel()
		task := core.NewTask("Buy groceries")
		if task.Text != "Buy groceries" {
			t.Errorf("text = %q, want %q", task.Text, "Buy groceries")
		}
		if task.Status != core.StatusTodo {
			t.Errorf("status = %q, want %q", task.Status, core.StatusTodo)
		}
	})

	t.Run("task with context", func(t *testing.T) {
		t.Parallel()
		task := core.NewTaskWithContext("Buy groceries", "Need food for the week")
		if task.Context != "Need food for the week" {
			t.Errorf("context = %q, want %q", task.Context, "Need food for the week")
		}
	})

	t.Run("task with type and effort", func(t *testing.T) {
		t.Parallel()
		task := core.NewTask("Write tests")
		task.Type = core.TypeTechnical
		task.Effort = core.EffortMedium
		if err := task.Validate(); err != nil {
			t.Errorf("validate: %v", err)
		}
	})

	t.Run("task with invalid type", func(t *testing.T) {
		t.Parallel()
		task := core.NewTask("Write tests")
		task.Type = core.TaskType("invalid")
		if err := task.Validate(); err == nil {
			t.Error("expected validation error for invalid type")
		}
	})

	t.Run("task with invalid effort", func(t *testing.T) {
		t.Parallel()
		task := core.NewTask("Write tests")
		task.Effort = core.TaskEffort("invalid")
		if err := task.Validate(); err == nil {
			t.Error("expected validation error for invalid effort")
		}
	})
}

func TestTaskPool_FindByPrefix(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	t1 := core.NewTask("Task 1")
	t1.ID = "abc-111"
	t2 := core.NewTask("Task 2")
	t2.ID = "abc-222"
	t3 := core.NewTask("Task 3")
	t3.ID = "def-333"
	pool.AddTask(t1)
	pool.AddTask(t2)
	pool.AddTask(t3)

	tests := []struct {
		name      string
		prefix    string
		wantCount int
	}{
		{"exact match", "abc-111", 1},
		{"partial match multiple", "abc", 2},
		{"partial match single", "def", 1},
		{"no match", "xyz", 0},
		{"empty prefix matches all", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matches := pool.FindByPrefix(tt.prefix)
			if len(matches) != tt.wantCount {
				t.Errorf("FindByPrefix(%q) returned %d matches, want %d", tt.prefix, len(matches), tt.wantCount)
			}
		})
	}
}

func TestNewTaskCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := newTaskCmd()
	if cmd.Use != "task" {
		t.Errorf("Use = %q, want %q", cmd.Use, "task")
	}

	subCmds := cmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subCmds {
		names[sub.Name()] = true
	}

	if !names["add"] {
		t.Error("missing 'add' subcommand")
	}
	if !names["complete"] {
		t.Error("missing 'complete' subcommand")
	}
}

func TestNewTaskAddCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := newTaskAddCmd()

	flags := []string{"context", "type", "effort"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag %q", name)
		}
	}
}

// fakeProvider implements core.TaskProvider for testing.
type fakeProvider struct {
	saved []*core.Task
}

func (f *fakeProvider) Name() string                        { return "fake" }
func (f *fakeProvider) LoadTasks() ([]*core.Task, error)    { return nil, nil }
func (f *fakeProvider) SaveTask(task *core.Task) error      { f.saved = append(f.saved, task); return nil }
func (f *fakeProvider) SaveTasks(_ []*core.Task) error      { return nil }
func (f *fakeProvider) DeleteTask(_ string) error           { return nil }
func (f *fakeProvider) MarkComplete(_ string) error         { return nil }
func (f *fakeProvider) Watch() <-chan core.ChangeEvent      { return nil }
func (f *fakeProvider) HealthCheck() core.HealthCheckResult { return core.HealthCheckResult{} }
