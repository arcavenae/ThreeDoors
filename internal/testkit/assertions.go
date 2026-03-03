package testkit

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// RequireTaskCount fails the test immediately if the task count doesn't match.
func RequireTaskCount(t *testing.T, tasks []*core.Task, want int) {
	t.Helper()
	if len(tasks) != want {
		t.Fatalf("got %d tasks, want %d", len(tasks), want)
	}
}

// AssertTaskText checks that a task with the given ID has the expected text.
func AssertTaskText(t *testing.T, tasks []*core.Task, id, wantText string) {
	t.Helper()
	for _, task := range tasks {
		if task.ID == id {
			if task.Text != wantText {
				t.Errorf("task %q: Text = %q, want %q", id, task.Text, wantText)
			}
			return
		}
	}
	t.Errorf("task %q not found in %d tasks", id, len(tasks))
}

// AssertTaskStatus checks that a task with the given ID has the expected status.
func AssertTaskStatus(t *testing.T, tasks []*core.Task, id string, wantStatus core.TaskStatus) {
	t.Helper()
	for _, task := range tasks {
		if task.ID == id {
			if task.Status != wantStatus {
				t.Errorf("task %q: Status = %q, want %q", id, task.Status, wantStatus)
			}
			return
		}
	}
	t.Errorf("task %q not found in %d tasks", id, len(tasks))
}

// AssertTaskExists checks that a task with the given ID exists in the slice.
func AssertTaskExists(t *testing.T, tasks []*core.Task, id string) {
	t.Helper()
	for _, task := range tasks {
		if task.ID == id {
			return
		}
	}
	t.Errorf("task %q not found in %d tasks", id, len(tasks))
}

// AssertTaskAbsent checks that no task with the given ID exists in the slice.
func AssertTaskAbsent(t *testing.T, tasks []*core.Task, id string) {
	t.Helper()
	for _, task := range tasks {
		if task.ID == id {
			t.Errorf("task %q should not be present, but was found", id)
			return
		}
	}
}
