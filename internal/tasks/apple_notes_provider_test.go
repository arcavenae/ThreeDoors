package tasks

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockExecutor returns a CommandExecutor that returns canned output.
func mockExecutor(output string, err error) CommandExecutor {
	return func(ctx context.Context, script string) (string, error) {
		return output, err
	}
}

// --- Interface compliance ---

func TestAppleNotesProvider_ImplementsTaskProvider(t *testing.T) {
	var _ TaskProvider = (*AppleNotesProvider)(nil)
}

// --- parseNoteBody table-driven tests (AC: 4) ---

func TestAppleNotesProvider_ParseNoteBody(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))

	tests := []struct {
		name       string
		body       string
		wantCount  int
		wantTexts  []string
		wantStatus []TaskStatus
	}{
		{
			name:       "basic checkboxes",
			body:       "- [ ] Buy milk\n- [x] Buy eggs",
			wantCount:  2,
			wantTexts:  []string{"Buy milk", "Buy eggs"},
			wantStatus: []TaskStatus{StatusTodo, StatusComplete},
		},
		{
			name:       "plain text lines become todo tasks",
			body:       "Plain task line",
			wantCount:  1,
			wantTexts:  []string{"Plain task line"},
			wantStatus: []TaskStatus{StatusTodo},
		},
		{
			name:      "empty note returns no tasks",
			body:      "",
			wantCount: 0,
		},
		{
			name:      "only empty lines returns no tasks",
			body:      "\n\n\n",
			wantCount: 0,
		},
		{
			name:      "whitespace-only lines returns no tasks",
			body:      "  \n  \t  \n",
			wantCount: 0,
		},
		{
			name:       "capital X is complete",
			body:       "- [X] Capital X",
			wantCount:  1,
			wantTexts:  []string{"Capital X"},
			wantStatus: []TaskStatus{StatusComplete},
		},
		{
			name:       "asterisk prefix works",
			body:       "* [ ] Asterisk item",
			wantCount:  1,
			wantTexts:  []string{"Asterisk item"},
			wantStatus: []TaskStatus{StatusTodo},
		},
		{
			name:       "indented checkbox",
			body:       "  - [ ] Indented",
			wantCount:  1,
			wantTexts:  []string{"Indented"},
			wantStatus: []TaskStatus{StatusTodo},
		},
		{
			name:       "empty lines between tasks are skipped",
			body:       "Buy milk\n\nBuy bread",
			wantCount:  2,
			wantTexts:  []string{"Buy milk", "Buy bread"},
			wantStatus: []TaskStatus{StatusTodo, StatusTodo},
		},
		{
			name:       "unicode characters pass through",
			body:       "Task with 🎯 emoji",
			wantCount:  1,
			wantTexts:  []string{"Task with 🎯 emoji"},
			wantStatus: []TaskStatus{StatusTodo},
		},
		{
			name:       "mixed checkboxes and plain text",
			body:       "Shopping List\n- [ ] Buy milk\n- [x] Buy eggs\n- [ ] Buy bread\n\nWork Tasks\n- [ ] Review PR\n- [ ] Update docs",
			wantCount:  7,
			wantTexts:  []string{"Shopping List", "Buy milk", "Buy eggs", "Buy bread", "Work Tasks", "Review PR", "Update docs"},
			wantStatus: []TaskStatus{StatusTodo, StatusTodo, StatusComplete, StatusTodo, StatusTodo, StatusTodo, StatusTodo},
		},
		{
			name:       "asterisk completed checkbox",
			body:       "* [x] Done item\n* [X] Also done",
			wantCount:  2,
			wantTexts:  []string{"Done item", "Also done"},
			wantStatus: []TaskStatus{StatusComplete, StatusComplete},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := provider.parseNoteBody(tt.body)

			if len(tasks) != tt.wantCount {
				t.Fatalf("parseNoteBody() returned %d tasks, want %d", len(tasks), tt.wantCount)
			}

			for i := 0; i < tt.wantCount; i++ {
				if i < len(tt.wantTexts) && tasks[i].Text != tt.wantTexts[i] {
					t.Errorf("task[%d].Text = %q, want %q", i, tasks[i].Text, tt.wantTexts[i])
				}
				if i < len(tt.wantStatus) && tasks[i].Status != tt.wantStatus[i] {
					t.Errorf("task[%d].Status = %q, want %q", i, tasks[i].Status, tt.wantStatus[i])
				}
			}
		})
	}
}

// --- Deterministic ID generation ---

func TestAppleNotesProvider_DeterministicIDs(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))

	t.Run("same input produces same IDs", func(t *testing.T) {
		body := "- [ ] Task A\n- [ ] Task B"
		tasks1 := provider.parseNoteBody(body)
		tasks2 := provider.parseNoteBody(body)

		if len(tasks1) != len(tasks2) {
			t.Fatalf("different task counts: %d vs %d", len(tasks1), len(tasks2))
		}
		for i := range tasks1 {
			if tasks1[i].ID != tasks2[i].ID {
				t.Errorf("task[%d] ID mismatch: %q vs %q", i, tasks1[i].ID, tasks2[i].ID)
			}
		}
	})

	t.Run("different note titles produce different IDs", func(t *testing.T) {
		providerA := NewAppleNotesProviderWithExecutor("NoteA", mockExecutor("", nil))
		providerB := NewAppleNotesProviderWithExecutor("NoteB", mockExecutor("", nil))

		body := "- [ ] Same task"
		tasksA := providerA.parseNoteBody(body)
		tasksB := providerB.parseNoteBody(body)

		if len(tasksA) == 0 || len(tasksB) == 0 {
			t.Fatal("expected at least 1 task from each provider")
		}
		if tasksA[0].ID == tasksB[0].ID {
			t.Error("different note titles should produce different IDs")
		}
	})
}

// --- LoadTasks with mock executor (AC: 1, 6) ---

func TestAppleNotesProvider_LoadTasks_Success(t *testing.T) {
	plaintext := "- [ ] Buy milk\n- [x] Buy eggs\n- [ ] Buy bread"
	provider := NewAppleNotesProviderWithExecutor("ThreeDoors Tasks", mockExecutor(plaintext, nil))

	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("LoadTasks() returned %d tasks, want 3", len(tasks))
	}
	if tasks[0].Text != "Buy milk" {
		t.Errorf("tasks[0].Text = %q, want %q", tasks[0].Text, "Buy milk")
	}
	if tasks[1].Status != StatusComplete {
		t.Errorf("tasks[1].Status = %q, want %q", tasks[1].Status, StatusComplete)
	}
}

func TestAppleNotesProvider_LoadTasks_NoteNotFound(t *testing.T) {
	notFoundErr := errors.New("execution error: Can't get note \"ThreeDoors Tasks\"")
	provider := NewAppleNotesProviderWithExecutor("ThreeDoors Tasks", mockExecutor("", notFoundErr))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error for note not found, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found', got: %v", err)
	}
}

func TestAppleNotesProvider_LoadTasks_PermissionDenied(t *testing.T) {
	permErr := errors.New("execution error: Not authorized to send Apple events to Notes")
	provider := NewAppleNotesProviderWithExecutor("ThreeDoors Tasks", mockExecutor("", permErr))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error for permission denied, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should contain 'permission denied', got: %v", err)
	}
}

func TestAppleNotesProvider_LoadTasks_Timeout(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("ThreeDoors Tasks", mockExecutor("", context.DeadlineExceeded))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error for timeout, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error should contain 'timed out', got: %v", err)
	}
}

func TestAppleNotesProvider_LoadTasks_EmptyNote(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("ThreeDoors Tasks", mockExecutor("", nil))

	tasks, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("LoadTasks() returned %d tasks for empty note, want 0", len(tasks))
	}
}

// --- Read-only sentinel error tests ---

func TestAppleNotesProvider_SaveTask_ReturnsErrReadOnly(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	task := newTestTask("aaa", "Test", StatusTodo, baseTime)

	err := provider.SaveTask(task)
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("SaveTask() error = %v, want ErrReadOnly", err)
	}
}

func TestAppleNotesProvider_SaveTasks_ReturnsErrReadOnly(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))

	err := provider.SaveTasks([]*Task{})
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("SaveTasks() error = %v, want ErrReadOnly", err)
	}
}

func TestAppleNotesProvider_DeleteTask_ReturnsErrReadOnly(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))

	err := provider.DeleteTask("aaa")
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("DeleteTask() error = %v, want ErrReadOnly", err)
	}
}
