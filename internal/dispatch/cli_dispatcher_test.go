package dispatch

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRunner implements CommandRunner for testing.
type mockRunner struct {
	output []byte
	err    error
	calls  []mockCall
}

type mockCall struct {
	name string
	args []string
}

func (m *mockRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	m.calls = append(m.calls, mockCall{name: name, args: args})
	return m.output, m.err
}

func TestCLIDispatcher_CreateWorker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		output   string
		err      error
		wantName string
		wantErr  bool
		wantArgs []string
	}{
		{
			name:     "parses Created worker prefix",
			output:   "Created worker: happy-fox\n",
			wantName: "happy-fox",
			wantArgs: []string{"worker", "create", "implement feature X"},
		},
		{
			name:     "parses Worker created prefix",
			output:   "Worker created: brave-lion\n",
			wantName: "brave-lion",
		},
		{
			name:     "parses bare name",
			output:   "clever-squirrel\n",
			wantName: "clever-squirrel",
		},
		{
			name:    "command error",
			err:     errors.New("connection refused"),
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &mockRunner{output: []byte(tt.output), err: tt.err}
			d := NewCLIDispatcher(m)

			got, err := d.CreateWorker(context.Background(), "implement feature X")
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantName {
				t.Errorf("CreateWorker() = %q, want %q", got, tt.wantName)
			}
			if tt.wantArgs != nil && len(m.calls) > 0 {
				call := m.calls[0]
				if call.name != multiclaude {
					t.Errorf("called %q, want %q", call.name, multiclaude)
				}
				if len(call.args) != len(tt.wantArgs) {
					t.Fatalf("args len = %d, want %d", len(call.args), len(tt.wantArgs))
				}
				for i, a := range call.args {
					if a != tt.wantArgs[i] {
						t.Errorf("arg[%d] = %q, want %q", i, a, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestCLIDispatcher_ListWorkers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		output  string
		err     error
		want    int
		wantErr bool
	}{
		{
			name: "parses worker list with header",
			output: `NAME        STATUS    BRANCH              TASK
happy-fox   running   work/happy-fox      implement auth
brave-lion  idle      work/brave-lion     fix bug #42`,
			want: 2,
		},
		{
			name:   "empty output",
			output: "",
			want:   0,
		},
		{
			name:    "command error",
			err:     errors.New("not available"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &mockRunner{output: []byte(tt.output), err: tt.err}
			d := NewCLIDispatcher(m)

			got, err := d.ListWorkers(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListWorkers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("ListWorkers() returned %d workers, want %d", len(got), tt.want)
			}
		})
	}
}

func TestCLIDispatcher_ListWorkers_Fields(t *testing.T) {
	t.Parallel()

	output := `NAME        STATUS    BRANCH              TASK
happy-fox   running   work/happy-fox      implement auth module`

	m := &mockRunner{output: []byte(output)}
	d := NewCLIDispatcher(m)

	workers, err := d.ListWorkers(context.Background())
	if err != nil {
		t.Fatalf("ListWorkers() error = %v", err)
	}
	if len(workers) != 1 {
		t.Fatalf("got %d workers, want 1", len(workers))
	}

	w := workers[0]
	if w.Name != "happy-fox" {
		t.Errorf("Name = %q, want %q", w.Name, "happy-fox")
	}
	if w.Status != "running" {
		t.Errorf("Status = %q, want %q", w.Status, "running")
	}
	if w.Branch != "work/happy-fox" {
		t.Errorf("Branch = %q, want %q", w.Branch, "work/happy-fox")
	}
	if w.Task != "implement auth module" {
		t.Errorf("Task = %q, want %q", w.Task, "implement auth module")
	}
}

func TestCLIDispatcher_GetHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		output  string
		err     error
		want    int
		wantErr bool
	}{
		{
			name: "parses history entries",
			output: `WORKER       STATUS     PR
happy-fox    completed  #42
brave-lion   failed     #43`,
			want: 2,
		},
		{
			name:   "empty output",
			output: "",
			want:   0,
		},
		{
			name:    "command error",
			err:     errors.New("timeout"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &mockRunner{output: []byte(tt.output), err: tt.err}
			d := NewCLIDispatcher(m)

			got, err := d.GetHistory(context.Background(), 10)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetHistory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("GetHistory() returned %d entries, want %d", len(got), tt.want)
			}
		})
	}
}

func TestCLIDispatcher_GetHistory_Fields(t *testing.T) {
	t.Parallel()

	output := `WORKER       STATUS     PR
happy-fox    completed  #42 https://github.com/example/repo/pull/42 2025-01-15T10:30:00Z`

	m := &mockRunner{output: []byte(output)}
	d := NewCLIDispatcher(m)

	entries, err := d.GetHistory(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}

	e := entries[0]
	if e.WorkerName != "happy-fox" {
		t.Errorf("WorkerName = %q, want %q", e.WorkerName, "happy-fox")
	}
	if e.Status != "completed" {
		t.Errorf("Status = %q, want %q", e.Status, "completed")
	}
	if e.PRNumber != 42 {
		t.Errorf("PRNumber = %d, want %d", e.PRNumber, 42)
	}
	if e.PRURL != "https://github.com/example/repo/pull/42" {
		t.Errorf("PRURL = %q, want %q", e.PRURL, "https://github.com/example/repo/pull/42")
	}
	if e.CompletedAt == nil {
		t.Fatal("CompletedAt is nil, want non-nil")
	}
	expected := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	if !e.CompletedAt.Equal(expected) {
		t.Errorf("CompletedAt = %v, want %v", e.CompletedAt, expected)
	}
}

func TestCLIDispatcher_RemoveWorker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		output  string
		err     error
		wantErr bool
	}{
		{
			name:   "success",
			output: "Worker removed\n",
		},
		{
			name:    "command error",
			err:     errors.New("worker not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &mockRunner{output: []byte(tt.output), err: tt.err}
			d := NewCLIDispatcher(m)

			err := d.RemoveWorker(context.Background(), "happy-fox")
			if (err != nil) != tt.wantErr {
				t.Fatalf("RemoveWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(m.calls) != 1 {
				t.Fatalf("expected 1 call, got %d", len(m.calls))
			}
			call := m.calls[0]
			if call.name != multiclaude {
				t.Errorf("called %q, want %q", call.name, multiclaude)
			}
			wantArgs := []string{"worker", "rm", "happy-fox"}
			if len(call.args) != len(wantArgs) {
				t.Fatalf("args len = %d, want %d", len(call.args), len(wantArgs))
			}
			for i, a := range call.args {
				if a != wantArgs[i] {
					t.Errorf("arg[%d] = %q, want %q", i, a, wantArgs[i])
				}
			}
		})
	}
}

func TestBuildTaskDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		item     QueueItem
		contains []string
	}{
		{
			name: "full item",
			item: QueueItem{
				TaskText:           "implement auth module",
				Context:            "uses JWT tokens",
				AcceptanceCriteria: []string{"login works", "tokens expire"},
				Scope:              "internal/auth package only",
			},
			contains: []string{
				"Task: implement auth module",
				"Context: uses JWT tokens",
				"- login works",
				"- tokens expire",
				"Scope: internal/auth package only",
				"Sign all commits",
				"fork workflow",
			},
		},
		{
			name: "minimal item",
			item: QueueItem{
				TaskText: "fix bug",
			},
			contains: []string{
				"Task: fix bug",
				"Sign all commits",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildTaskDescription(tt.item)
			for _, s := range tt.contains {
				if !containsString(got, s) {
					t.Errorf("BuildTaskDescription() missing %q in:\n%s", s, got)
				}
			}
		})
	}
}

func TestBuildTaskDescription_NoOptionalSections(t *testing.T) {
	t.Parallel()

	item := QueueItem{TaskText: "simple task"}
	got := BuildTaskDescription(item)

	if containsString(got, "Context:") {
		t.Error("should not contain Context: for empty context")
	}
	if containsString(got, "Acceptance Criteria:") {
		t.Error("should not contain Acceptance Criteria: for empty AC")
	}
	if containsString(got, "Scope:") {
		t.Error("should not contain Scope: for empty scope")
	}
}

func TestCLIDispatcher_ErrorWrapping(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("connection refused")
	m := &mockRunner{err: baseErr}
	d := NewCLIDispatcher(m)
	ctx := context.Background()

	_, err := d.CreateWorker(ctx, "task")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, baseErr) {
		t.Errorf("CreateWorker error should wrap base error, got: %v", err)
	}

	_, err = d.ListWorkers(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, baseErr) {
		t.Errorf("ListWorkers error should wrap base error, got: %v", err)
	}

	_, err = d.GetHistory(ctx, 5)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, baseErr) {
		t.Errorf("GetHistory error should wrap base error, got: %v", err)
	}

	err = d.RemoveWorker(ctx, "test")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, baseErr) {
		t.Errorf("RemoveWorker error should wrap base error, got: %v", err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
