package dispatch

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateFollowUpTasks_ReviewTask(t *testing.T) {
	t.Parallel()

	item := QueueItem{
		ID:       "dq-test01",
		TaskID:   "task-abc",
		TaskText: "implement feature X",
		Status:   QueueItemCompleted,
		PRNumber: 42,
	}

	tasks := GenerateFollowUpTasks(item, nil)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	got := tasks[0]
	if got.Text != "Review PR #42: implement feature X" {
		t.Errorf("text = %q, want %q", got.Text, "Review PR #42: implement feature X")
	}
	if got.Context != "Auto-generated from task task-abc" {
		t.Errorf("context = %q, want traceability context", got.Context)
	}
	if got.DevDispatch == nil || got.DevDispatch.PRNumber != 42 {
		t.Errorf("DevDispatch.PRNumber = %v, want 42", got.DevDispatch)
	}
}

func TestGenerateFollowUpTasks_NoPRNumber(t *testing.T) {
	t.Parallel()

	item := QueueItem{
		ID:       "dq-nopr",
		TaskID:   "task-xyz",
		TaskText: "broken task",
		Status:   QueueItemFailed,
	}

	tasks := GenerateFollowUpTasks(item, nil)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks for no PR, got %d", len(tasks))
	}
}

func TestGenerateFollowUpTasks_Deduplication(t *testing.T) {
	t.Parallel()

	item := QueueItem{
		ID:       "dq-dedup",
		TaskID:   "task-dup",
		TaskText: "fix bug",
		Status:   QueueItemCompleted,
		PRNumber: 55,
	}

	existing := map[string]bool{
		"Review PR #55: fix bug": true,
	}

	tasks := GenerateFollowUpTasks(item, existing)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks (dedup), got %d", len(tasks))
	}
}

func TestGenerateFollowUpTasks_DeduplicationByPrefix(t *testing.T) {
	t.Parallel()

	item := QueueItem{
		ID:       "dq-prefix",
		TaskID:   "task-pfx",
		TaskText: "add tests",
		Status:   QueueItemCompleted,
		PRNumber: 77,
	}

	existing := map[string]bool{
		"Review PR #77: old description": true,
	}

	tasks := GenerateFollowUpTasks(item, existing)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks (prefix dedup), got %d", len(tasks))
	}
}

func TestGenerateFollowUpTasks_CIFailure(t *testing.T) {
	t.Parallel()

	item := QueueItem{
		ID:       "dq-cifail",
		TaskID:   "task-ci",
		TaskText: "refactor auth",
		Status:   QueueItemFailed,
		PRNumber: 99,
		Error:    "lint check failed",
	}

	tasks := GenerateFollowUpTasks(item, nil)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks (review + CI fix), got %d", len(tasks))
	}

	if !strings.HasPrefix(tasks[0].Text, "Review PR #99:") {
		t.Errorf("first task should be review, got %q", tasks[0].Text)
	}
	if tasks[1].Text != "Fix CI on PR #99: lint check failed" {
		t.Errorf("second task text = %q, want CI fix text", tasks[1].Text)
	}
	if tasks[1].DevDispatch == nil || tasks[1].DevDispatch.PRNumber != 99 {
		t.Error("CI fix task should have DevDispatch with PRNumber")
	}
}

func TestGenerateFollowUpTasks_CIFailureDedup(t *testing.T) {
	t.Parallel()

	item := QueueItem{
		ID:       "dq-cidedup",
		TaskID:   "task-cid",
		TaskText: "deploy",
		Status:   QueueItemFailed,
		PRNumber: 88,
		Error:    "test failure",
	}

	existing := map[string]bool{
		"Fix CI on PR #88: previous error": true,
	}

	tasks := GenerateFollowUpTasks(item, existing)
	// Review task should still be created, CI fix deduped
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task (review only, CI deduped), got %d", len(tasks))
	}
	if !strings.HasPrefix(tasks[0].Text, "Review PR #88:") {
		t.Errorf("expected review task, got %q", tasks[0].Text)
	}
}

func TestHasTaskWithPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		texts  map[string]bool
		prefix string
		want   bool
	}{
		{"nil map", nil, "Review PR #1:", false},
		{"empty map", map[string]bool{}, "Review PR #1:", false},
		{"match", map[string]bool{"Review PR #1: foo": true}, "Review PR #1:", true},
		{"no match", map[string]bool{"Review PR #2: bar": true}, "Review PR #1:", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := hasTaskWithPrefix(tt.texts, tt.prefix)
			if got != tt.want {
				t.Errorf("hasTaskWithPrefix(%v, %q) = %v, want %v", tt.texts, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestGenerateFollowUpTasks_FailedNoError(t *testing.T) {
	t.Parallel()

	item := QueueItem{
		ID:       "dq-failnoerr",
		TaskID:   "task-fne",
		TaskText: "build widget",
		Status:   QueueItemFailed,
		PRNumber: 33,
	}

	tasks := GenerateFollowUpTasks(item, nil)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task (review only, no CI fix without error), got %d", len(tasks))
	}
	if want := fmt.Sprintf("Review PR #%d: %s", 33, "build widget"); tasks[0].Text != want {
		t.Errorf("text = %q, want %q", tasks[0].Text, want)
	}
}
