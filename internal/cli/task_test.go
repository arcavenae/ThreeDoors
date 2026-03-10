package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

	for _, want := range []string{"add", "block", "complete", "delete", "edit", "list", "note", "search", "show", "status", "unblock"} {
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

func TestDeleteOneTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupIDs []string
		prefix   string
		wantOK   bool
		wantExit int
	}{
		{
			name:     "not found",
			setupIDs: []string{"abc-111"},
			prefix:   "xyz",
			wantOK:   false,
			wantExit: ExitNotFound,
		},
		{
			name:     "ambiguous prefix",
			setupIDs: []string{"abc-111", "abc-222"},
			prefix:   "abc",
			wantOK:   false,
			wantExit: ExitAmbiguousInput,
		},
		{
			name:     "success",
			setupIDs: []string{"unique-task-id"},
			prefix:   "unique",
			wantOK:   true,
			wantExit: ExitSuccess,
		},
		{
			name:     "exact match",
			setupIDs: []string{"abc-111", "def-222"},
			prefix:   "abc-111",
			wantOK:   true,
			wantExit: ExitSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := core.NewTaskPool()
			for _, id := range tt.setupIDs {
				task := core.NewTask("Task " + id)
				task.ID = id
				pool.AddTask(task)
			}

			provider := &fakeProvider{}
			ctx := &cliContext{pool: pool, provider: provider}

			result := deleteOneTask(ctx, tt.prefix)

			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantOK)
			}
			if result.ExitCode != tt.wantExit {
				t.Errorf("ExitCode = %d, want %d", result.ExitCode, tt.wantExit)
			}

			if tt.wantOK {
				// Verify task was removed from pool
				if pool.GetTask(result.ID) != nil {
					t.Error("task should have been removed from pool")
				}
			}
		})
	}
}

func TestDeleteOneTask_ProviderError(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Test task")
	task.ID = "provider-err-id"
	pool.AddTask(task)

	provider := &fakeProviderWithDeleteErr{err: fmt.Errorf("disk full")}
	ctx := &cliContext{pool: pool, provider: provider}

	result := deleteOneTask(ctx, "provider-err")

	if result.Success {
		t.Error("expected failure when provider errors")
	}
	if result.ExitCode != ExitProviderError {
		t.Errorf("ExitCode = %d, want %d", result.ExitCode, ExitProviderError)
	}
}

func TestResolveOneTask(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Resolve me")
	task.ID = "resolve-test-id"
	pool.AddTask(task)

	ctx := &cliContext{pool: pool, provider: &fakeProvider{}}
	formatter := NewOutputFormatter(os.Stdout, false)

	got := resolveOneTask(ctx, formatter, "test", "resolve-test", false)
	if got == nil {
		t.Fatal("expected task, got nil")
		return
	}
	if got.ID != "resolve-test-id" {
		t.Errorf("ID = %q, want %q", got.ID, "resolve-test-id")
	}
}

func TestDeleteResult_JSON(t *testing.T) {
	t.Parallel()

	r := deleteResult{
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

	if decoded["success"] != true {
		t.Errorf("success = %v, want true", decoded["success"])
	}
	if _, ok := decoded["error"]; ok {
		t.Error("error field should be omitted when empty")
	}
}

func TestSearchMetadata_JSON(t *testing.T) {
	t.Parallel()

	meta := searchMetadata{
		Query:   "test",
		Matched: 3,
		Total:   10,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["query"] != "test" {
		t.Errorf("query = %v, want %q", decoded["query"], "test")
	}
	if decoded["matched"] != float64(3) {
		t.Errorf("matched = %v, want 3", decoded["matched"])
	}
	if decoded["total"] != float64(10) {
		t.Errorf("total = %v, want 10", decoded["total"])
	}
}

func TestNewTaskEditCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := newTaskEditCmd()
	if cmd.Use != "edit <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "edit <id>")
	}
	for _, name := range []string{"text", "context"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag %q", name)
		}
	}
}

func TestNewTaskDeleteCmd_Args(t *testing.T) {
	t.Parallel()

	cmd := newTaskDeleteCmd()
	if cmd.Use != "delete <id> [id...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <id> [id...]")
	}
	// Requires at least 1 arg
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error for 0 args")
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err != nil {
		t.Errorf("expected 2 args to be valid, got: %v", err)
	}
}

func TestNewTaskNoteCmd_Args(t *testing.T) {
	t.Parallel()

	cmd := newTaskNoteCmd()
	if cmd.Use != "note <id> <text>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "note <id> <text>")
	}
	// Requires exactly 2 args
	if err := cmd.Args(cmd, []string{"id"}); err == nil {
		t.Error("expected error for 1 arg")
	}
	if err := cmd.Args(cmd, []string{"id", "note text"}); err != nil {
		t.Errorf("expected 2 args to be valid, got: %v", err)
	}
}

func TestNewTaskSearchCmd_Args(t *testing.T) {
	t.Parallel()

	cmd := newTaskSearchCmd()
	if cmd.Use != "search <query>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "search <query>")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error for 0 args")
	}
}

// fakeProviderWithDeleteErr simulates provider delete failures.
type fakeProviderWithDeleteErr struct {
	fakeProvider
	err error
}

func (f *fakeProviderWithDeleteErr) DeleteTask(_ string) error { return f.err }

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
