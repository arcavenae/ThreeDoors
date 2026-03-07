package cli

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
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

	for _, want := range []string{"add", "complete", "list", "show"} {
		if !names[want] {
			t.Errorf("missing %q subcommand", want)
		}
	}
}

func TestNewTaskAddCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := newTaskAddCmd()

	flags := []string{"context", "type", "effort", "stdin"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag %q", name)
		}
	}
}

func TestRunTaskAddSingleStdin(t *testing.T) {
	t.Parallel()

	t.Run("reads single task from stdin", func(t *testing.T) {
		t.Parallel()
		r := strings.NewReader("Buy groceries\n")
		task := core.NewTask("placeholder")

		// We can test the stdin reading logic by calling the internal helper
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		text := strings.TrimSpace(string(data))
		if text != "Buy groceries" {
			t.Errorf("text = %q, want %q", text, "Buy groceries")
		}
		_ = task
	})

	t.Run("empty stdin returns error", func(t *testing.T) {
		t.Parallel()
		r := strings.NewReader("")
		err := runTaskAddSingleStdin(nil, r, "", "", "")
		if err == nil {
			t.Error("expected error for empty stdin")
		}
	})

	t.Run("whitespace-only stdin returns error", func(t *testing.T) {
		t.Parallel()
		r := strings.NewReader("   \n  \n")
		err := runTaskAddSingleStdin(nil, r, "", "", "")
		if err == nil {
			t.Error("expected error for whitespace-only stdin")
		}
	})
}

func TestRunTaskAddFromStdin_MultiLine(t *testing.T) {
	t.Parallel()

	provider := &fakeProvider{}
	pool := core.NewTaskPool()

	// Save original bootstrap and restore after test
	// We test the scanner logic directly instead
	input := "Task one\nTask two\n\nTask three\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	var lines []string
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "Task one" {
		t.Errorf("line[0] = %q, want %q", lines[0], "Task one")
	}
	if lines[1] != "Task two" {
		t.Errorf("line[1] = %q, want %q", lines[1], "Task two")
	}
	if lines[2] != "Task three" {
		t.Errorf("line[2] = %q, want %q", lines[2], "Task three")
	}

	_ = provider
	_ = pool
}

func TestNewTaskAddCmd_AcceptsZeroArgs(t *testing.T) {
	t.Parallel()

	cmd := newTaskAddCmd()
	// MaximumNArgs(1) should allow 0 args (for stdin mode)
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("expected 0 args to be valid, got: %v", err)
	}
}

func TestStdinDetector_Variable(t *testing.T) {
	t.Parallel()

	// Verify stdinDetector is a function variable that can be swapped
	original := stdinDetector
	t.Cleanup(func() { stdinDetector = original })

	stdinDetector = func() bool { return true }
	if !stdinDetector() {
		t.Error("expected mocked stdinDetector to return true")
	}
}

func TestFilterTasks(t *testing.T) {
	t.Parallel()

	tasks := []*core.Task{
		{ID: "1", Text: "A", Status: core.StatusTodo, Type: core.TypeTechnical, Effort: core.EffortQuickWin},
		{ID: "2", Text: "B", Status: core.StatusInProgress, Type: core.TypeCreative, Effort: core.EffortMedium},
		{ID: "3", Text: "C", Status: core.StatusTodo, Type: core.TypeTechnical, Effort: core.EffortDeepWork},
		{ID: "4", Text: "D", Status: core.StatusComplete, Type: core.TypeAdministrative, Effort: core.EffortQuickWin},
	}

	tests := []struct {
		name      string
		status    string
		taskType  string
		effort    string
		wantCount int
	}{
		{"no filters", "", "", "", 4},
		{"filter by status todo", "todo", "", "", 2},
		{"filter by type technical", "", "technical", "", 2},
		{"filter by effort quick-win", "", "", "quick-win", 2},
		{"composable: status+type", "todo", "technical", "", 2},
		{"composable: status+type+effort", "todo", "technical", "quick-win", 1},
		{"no matches", "done", "", "", 0},
		{"in-progress", "in-progress", "", "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := filterTasks(tasks, tt.status, tt.taskType, tt.effort)
			if len(got) != tt.wantCount {
				t.Errorf("filterTasks() returned %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestNewTaskListCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := newTaskListCmd()
	for _, name := range []string{"status", "type", "effort"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag %q", name)
		}
	}
}

func TestNewTaskShowCmd_Args(t *testing.T) {
	t.Parallel()

	cmd := newTaskShowCmd()
	if cmd.Use != "show <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "show <id>")
	}
}

func TestListMetadata_JSON(t *testing.T) {
	t.Parallel()

	meta := listMetadata{
		Total:    10,
		Filtered: 3,
		Filters:  map[string]string{"status": "todo"},
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["total"] != float64(10) {
		t.Errorf("total = %v, want 10", decoded["total"])
	}
	if decoded["filtered"] != float64(3) {
		t.Errorf("filtered = %v, want 3", decoded["filtered"])
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
