package cli

import (
	"encoding/json"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestBlockOneTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *core.Task
		reason   string
		wantOK   bool
		wantExit int
		wantErr  string
	}{
		{
			name: "block todo task",
			setup: func() *core.Task {
				task := core.NewTask("Do something")
				task.ID = "block-todo-id"
				return task
			},
			reason:   "Waiting on API",
			wantOK:   true,
			wantExit: ExitSuccess,
		},
		{
			name: "block in-progress task",
			setup: func() *core.Task {
				task := core.NewTask("Working on it")
				task.ID = "block-inprog-id"
				_ = task.UpdateStatus(core.StatusInProgress)
				return task
			},
			reason:   "Blocked by dependency",
			wantOK:   true,
			wantExit: ExitSuccess,
		},
		{
			name: "block complete task fails",
			setup: func() *core.Task {
				task := core.NewTask("Done task")
				task.ID = "block-done-id"
				_ = task.UpdateStatus(core.StatusComplete)
				return task
			},
			reason:   "Should fail",
			wantOK:   false,
			wantExit: ExitValidation,
			wantErr:  `invalid transition from "complete" to "blocked"`,
		},
		{
			name:     "block nonexistent task",
			setup:    nil,
			reason:   "no task",
			wantOK:   false,
			wantExit: ExitNotFound,
			wantErr:  "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := core.NewTaskPool()
			var idPrefix string
			if tt.setup != nil {
				task := tt.setup()
				pool.AddTask(task)
				idPrefix = task.ID
			} else {
				idPrefix = "nonexistent"
			}

			provider := &fakeProvider{}
			ctx := &cliContext{pool: pool, provider: provider}

			result := blockOneTask(ctx, idPrefix, tt.reason)

			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v (error: %s)", result.Success, tt.wantOK, result.Error)
			}
			if result.ExitCode != tt.wantExit {
				t.Errorf("ExitCode = %d, want %d", result.ExitCode, tt.wantExit)
			}
			if tt.wantErr != "" && result.Error != tt.wantErr {
				t.Errorf("Error = %q, want %q", result.Error, tt.wantErr)
			}
		})
	}
}

func TestBlockOneTask_SetsBlocker(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Blockable task")
	task.ID = "blocker-set-id"
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := blockOneTask(ctx, "blocker-set-id", "Waiting on API")

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if task.Blocker != "Waiting on API" {
		t.Errorf("Blocker = %q, want %q", task.Blocker, "Waiting on API")
	}
	if task.Status != core.StatusBlocked {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusBlocked)
	}
}

func TestBlockOneTask_PrefixMatch(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Prefix task")
	task.ID = "abcdef12-3456-7890-abcd-ef1234567890"
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := blockOneTask(ctx, "abcdef12", "prefix reason")

	if !result.Success {
		t.Errorf("expected success with prefix match, got error: %s", result.Error)
	}
	if result.ID != task.ID {
		t.Errorf("result ID = %q, want %q", result.ID, task.ID)
	}
}

func TestBlockOneTask_AmbiguousPrefix(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	t1 := core.NewTask("Task 1")
	t1.ID = "abc-111"
	t2 := core.NewTask("Task 2")
	t2.ID = "abc-222"
	pool.AddTask(t1)
	pool.AddTask(t2)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := blockOneTask(ctx, "abc", "reason")

	if result.Success {
		t.Error("expected failure for ambiguous prefix")
	}
	if result.ExitCode != ExitAmbiguousInput {
		t.Errorf("ExitCode = %d, want %d", result.ExitCode, ExitAmbiguousInput)
	}
}

func TestUnblockOneTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *core.Task
		wantOK   bool
		wantExit int
		wantErr  string
	}{
		{
			name: "unblock blocked task",
			setup: func() *core.Task {
				task := core.NewTask("Blocked task")
				task.ID = "unblock-ok-id"
				_ = task.UpdateStatus(core.StatusBlocked)
				_ = task.SetBlocker("Some reason")
				return task
			},
			wantOK:   true,
			wantExit: ExitSuccess,
		},
		{
			name: "unblock non-blocked task fails",
			setup: func() *core.Task {
				task := core.NewTask("Todo task")
				task.ID = "unblock-todo-id"
				return task
			},
			wantOK:   false,
			wantExit: ExitValidation,
			wantErr:  `task is not blocked (current status: "todo")`,
		},
		{
			name: "unblock in-progress task fails",
			setup: func() *core.Task {
				task := core.NewTask("In progress task")
				task.ID = "unblock-inprog-id"
				_ = task.UpdateStatus(core.StatusInProgress)
				return task
			},
			wantOK:   false,
			wantExit: ExitValidation,
			wantErr:  `task is not blocked (current status: "in-progress")`,
		},
		{
			name:     "unblock nonexistent task",
			setup:    nil,
			wantOK:   false,
			wantExit: ExitNotFound,
			wantErr:  "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := core.NewTaskPool()
			var idPrefix string
			if tt.setup != nil {
				task := tt.setup()
				pool.AddTask(task)
				idPrefix = task.ID
			} else {
				idPrefix = "nonexistent"
			}

			provider := &fakeProvider{}
			ctx := &cliContext{pool: pool, provider: provider}

			result := unblockOneTask(ctx, idPrefix)

			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v (error: %s)", result.Success, tt.wantOK, result.Error)
			}
			if result.ExitCode != tt.wantExit {
				t.Errorf("ExitCode = %d, want %d", result.ExitCode, tt.wantExit)
			}
			if tt.wantErr != "" && result.Error != tt.wantErr {
				t.Errorf("Error = %q, want %q", result.Error, tt.wantErr)
			}
		})
	}
}

func TestUnblockOneTask_ClearsBlocker(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Blocked task")
	task.ID = "unblock-clear-id"
	_ = task.UpdateStatus(core.StatusBlocked)
	_ = task.SetBlocker("Old reason")
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := unblockOneTask(ctx, "unblock-clear-id")

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if task.Status != core.StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, core.StatusTodo)
	}
	if task.Blocker != "" {
		t.Errorf("Blocker = %q, want empty", task.Blocker)
	}
}

func TestUnblockOneTask_PrefixMatch(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Prefix task")
	task.ID = "xyzdef12-3456-7890-abcd-ef1234567890"
	_ = task.UpdateStatus(core.StatusBlocked)
	_ = task.SetBlocker("reason")
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	result := unblockOneTask(ctx, "xyzdef12")

	if !result.Success {
		t.Errorf("expected success with prefix match, got error: %s", result.Error)
	}
}

func TestChangeOneTaskStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *core.Task
		target   core.TaskStatus
		wantOK   bool
		wantExit int
		wantErr  string
	}{
		{
			name: "todo to in-progress",
			setup: func() *core.Task {
				task := core.NewTask("Start work")
				task.ID = "status-inprog-id"
				return task
			},
			target:   core.StatusInProgress,
			wantOK:   true,
			wantExit: ExitSuccess,
		},
		{
			name: "in-progress to in-review",
			setup: func() *core.Task {
				task := core.NewTask("Review work")
				task.ID = "status-review-id"
				_ = task.UpdateStatus(core.StatusInProgress)
				return task
			},
			target:   core.StatusInReview,
			wantOK:   true,
			wantExit: ExitSuccess,
		},
		{
			name: "invalid transition complete to todo",
			setup: func() *core.Task {
				task := core.NewTask("Done task")
				task.ID = "status-invalid-id"
				_ = task.UpdateStatus(core.StatusComplete)
				return task
			},
			target:   core.StatusTodo,
			wantOK:   false,
			wantExit: ExitValidation,
			wantErr:  `invalid transition from "complete" to "todo"`,
		},
		{
			name: "same status is no-op",
			setup: func() *core.Task {
				task := core.NewTask("Same task")
				task.ID = "status-noop-id"
				return task
			},
			target:   core.StatusTodo,
			wantOK:   true,
			wantExit: ExitSuccess,
		},
		{
			name:     "not found",
			setup:    nil,
			target:   core.StatusInProgress,
			wantOK:   false,
			wantExit: ExitNotFound,
			wantErr:  "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := core.NewTaskPool()
			var idPrefix string
			if tt.setup != nil {
				task := tt.setup()
				pool.AddTask(task)
				idPrefix = task.ID
			} else {
				idPrefix = "nonexistent"
			}

			provider := &fakeProvider{}
			ctx := &cliContext{pool: pool, provider: provider}

			result := changeOneTaskStatus(ctx, idPrefix, tt.target)

			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v (error: %s)", result.Success, tt.wantOK, result.Error)
			}
			if result.ExitCode != tt.wantExit {
				t.Errorf("ExitCode = %d, want %d", result.ExitCode, tt.wantExit)
			}
			if tt.wantErr != "" && result.Error != tt.wantErr {
				t.Errorf("Error = %q, want %q", result.Error, tt.wantErr)
			}
		})
	}
}

func TestChangeOneTaskStatus_BatchSupport(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	t1 := core.NewTask("Task 1")
	t1.ID = "batch-aaa-111"
	t2 := core.NewTask("Task 2")
	t2.ID = "batch-bbb-222"
	t3 := core.NewTask("Task 3")
	t3.ID = "batch-ccc-333"
	pool.AddTask(t1)
	pool.AddTask(t2)
	pool.AddTask(t3)

	provider := &fakeProvider{}
	ctx := &cliContext{pool: pool, provider: provider}

	ids := []string{"batch-aaa", "batch-bbb", "batch-ccc"}
	results := make([]statusResult, 0, len(ids))
	for _, id := range ids {
		results = append(results, changeOneTaskStatus(ctx, id, core.StatusInProgress))
	}

	for i, r := range results {
		if !r.Success {
			t.Errorf("task %d: expected success, got error: %s", i, r.Error)
		}
		if r.OldStatus != "todo" {
			t.Errorf("task %d: OldStatus = %q, want %q", i, r.OldStatus, "todo")
		}
		if r.NewStatus != "in-progress" {
			t.Errorf("task %d: NewStatus = %q, want %q", i, r.NewStatus, "in-progress")
		}
	}

	if t1.Status != core.StatusInProgress {
		t.Errorf("t1 status = %q, want %q", t1.Status, core.StatusInProgress)
	}
	if t2.Status != core.StatusInProgress {
		t.Errorf("t2 status = %q, want %q", t2.Status, core.StatusInProgress)
	}
	if t3.Status != core.StatusInProgress {
		t.Errorf("t3 status = %q, want %q", t3.Status, core.StatusInProgress)
	}
}

func TestStatusResult_JSON(t *testing.T) {
	t.Parallel()

	r := statusResult{
		ID:        "abc-123",
		ShortID:   "abc-123",
		OldStatus: "todo",
		NewStatus: "in-progress",
		Success:   true,
		ExitCode:  0,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["old_status"] != "todo" {
		t.Errorf("old_status = %v, want todo", decoded["old_status"])
	}
	if decoded["new_status"] != "in-progress" {
		t.Errorf("new_status = %v, want in-progress", decoded["new_status"])
	}
	if _, ok := decoded["error"]; ok {
		t.Error("error field should be omitted when empty")
	}
}

func TestNewTaskBlockCmd_ReasonRequired(t *testing.T) {
	t.Parallel()

	cmd := newTaskBlockCmd()
	if cmd.Flags().Lookup("reason") == nil {
		t.Error("missing --reason flag")
	}
}

func TestNewTaskStatusCmd_Args(t *testing.T) {
	t.Parallel()

	cmd := newTaskStatusCmd()
	if cmd.Use != "status <id> [id...] <new-status>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status <id> [id...] <new-status>")
	}
}

func TestNewTaskCmd_IncludesNewSubcommands(t *testing.T) {
	t.Parallel()

	cmd := newTaskCmd()
	subCmds := cmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subCmds {
		names[sub.Name()] = true
	}

	for _, want := range []string{"block", "unblock", "status"} {
		if !names[want] {
			t.Errorf("missing %q subcommand", want)
		}
	}
}

func TestResolveTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		taskIDs  []string
		wantTask bool
		wantExit int
	}{
		{"found", "unique-id", []string{"unique-id"}, true, 0},
		{"not found", "missing", []string{"other-id"}, false, ExitNotFound},
		{"ambiguous", "abc", []string{"abc-111", "abc-222"}, false, ExitAmbiguousInput},
		{"prefix match", "uni", []string{"unique-id"}, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := core.NewTaskPool()
			for _, id := range tt.taskIDs {
				task := core.NewTask("task " + id)
				task.ID = id
				pool.AddTask(task)
			}

			ctx := &cliContext{pool: pool}
			task, result := resolveTask(ctx, tt.prefix)

			if tt.wantTask && task == nil {
				t.Error("expected task, got nil")
			}
			if !tt.wantTask && task != nil {
				t.Error("expected nil task")
			}
			if !tt.wantTask && result.ExitCode != tt.wantExit {
				t.Errorf("ExitCode = %d, want %d", result.ExitCode, tt.wantExit)
			}
		})
	}
}
