package applenotes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// mockExecutor returns a CommandExecutor that returns canned output.
func mockExecutor(output string, err error) CommandExecutor {
	return func(ctx context.Context, script string) (string, error) {
		return output, err
	}
}

// --- Interface compliance ---

func TestAppleNotesProvider_ImplementsTaskProvider(t *testing.T) {
	var _ core.TaskProvider = (*AppleNotesProvider)(nil)
}

// --- parseNoteBody table-driven tests (AC: 4) ---

func TestAppleNotesProvider_ParseNoteBody(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))

	tests := []struct {
		name       string
		body       string
		wantCount  int
		wantTexts  []string
		wantStatus []core.TaskStatus
	}{
		{
			name:       "basic checkboxes",
			body:       "- [ ] Buy milk\n- [x] Buy eggs",
			wantCount:  2,
			wantTexts:  []string{"Buy milk", "Buy eggs"},
			wantStatus: []core.TaskStatus{core.StatusTodo, core.StatusComplete},
		},
		{
			name:       "plain text lines become todo tasks",
			body:       "Plain task line",
			wantCount:  1,
			wantTexts:  []string{"Plain task line"},
			wantStatus: []core.TaskStatus{core.StatusTodo},
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
			wantStatus: []core.TaskStatus{core.StatusComplete},
		},
		{
			name:       "asterisk prefix works",
			body:       "* [ ] Asterisk item",
			wantCount:  1,
			wantTexts:  []string{"Asterisk item"},
			wantStatus: []core.TaskStatus{core.StatusTodo},
		},
		{
			name:       "indented checkbox",
			body:       "  - [ ] Indented",
			wantCount:  1,
			wantTexts:  []string{"Indented"},
			wantStatus: []core.TaskStatus{core.StatusTodo},
		},
		{
			name:       "empty lines between tasks are skipped",
			body:       "Buy milk\n\nBuy bread",
			wantCount:  2,
			wantTexts:  []string{"Buy milk", "Buy bread"},
			wantStatus: []core.TaskStatus{core.StatusTodo, core.StatusTodo},
		},
		{
			name:       "unicode characters pass through",
			body:       "Task with 🎯 emoji",
			wantCount:  1,
			wantTexts:  []string{"Task with 🎯 emoji"},
			wantStatus: []core.TaskStatus{core.StatusTodo},
		},
		{
			name:       "mixed checkboxes and plain text",
			body:       "Shopping List\n- [ ] Buy milk\n- [x] Buy eggs\n- [ ] Buy bread\n\nWork Tasks\n- [ ] Review PR\n- [ ] Update docs",
			wantCount:  7,
			wantTexts:  []string{"Shopping List", "Buy milk", "Buy eggs", "Buy bread", "Work Tasks", "Review PR", "Update docs"},
			wantStatus: []core.TaskStatus{core.StatusTodo, core.StatusTodo, core.StatusComplete, core.StatusTodo, core.StatusTodo, core.StatusTodo, core.StatusTodo},
		},
		{
			name:       "asterisk completed checkbox",
			body:       "* [x] Done item\n* [X] Also done",
			wantCount:  2,
			wantTexts:  []string{"Done item", "Also done"},
			wantStatus: []core.TaskStatus{core.StatusComplete, core.StatusComplete},
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
	if tasks[1].Status != core.StatusComplete {
		t.Errorf("tasks[1].Status = %q, want %q", tasks[1].Status, core.StatusComplete)
	}
}

func TestAppleNotesProvider_LoadTasks_NoteNotFound(t *testing.T) {
	notFoundErr := errors.New("execution error: Can't get note \"ThreeDoors Tasks\"")
	provider := NewAppleNotesProviderWithExecutor("ThreeDoors Tasks", mockExecutor("", notFoundErr))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("LoadTasks() expected error for note not found, got nil")
		return
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
		return
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
		return
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

// =============================================================================
// Story 2.4: Write Task Updates to Apple Notes — TDD Tests
// These tests define the expected behavior for write operations.
// =============================================================================

// --- scriptRecord tracks osascript calls for write verification ---

type scriptRecord struct {
	scripts []string
}

// recordingExecutor tracks all scripts executed and responds based on prefix matching.
func recordingExecutor(responses map[string]struct {
	output string
	err    error
},
) (CommandExecutor, *scriptRecord) {
	rec := &scriptRecord{}
	executor := func(ctx context.Context, script string) (string, error) {
		rec.scripts = append(rec.scripts, script)
		for prefix, resp := range responses {
			if strings.Contains(script, prefix) {
				return resp.output, resp.err
			}
		}
		return "", fmt.Errorf("unexpected script: %s", script)
	}
	return executor, rec
}

// --- readRawNoteBody tests ---

func TestAppleNotesProvider_ReadRawNoteBody_Success(t *testing.T) {
	expected := "- [ ] Buy milk\n- [x] Buy eggs\n- [ ] Buy bread"
	provider := NewAppleNotesProviderWithExecutor("ThreeDoors Tasks", mockExecutor(expected, nil))

	raw, err := provider.readRawNoteBody()
	if err != nil {
		t.Fatalf("readRawNoteBody() unexpected error: %v", err)
	}
	if raw != expected {
		t.Errorf("readRawNoteBody() = %q, want %q", raw, expected)
	}
}

func TestAppleNotesProvider_ReadRawNoteBody_Error(t *testing.T) {
	notFoundErr := errors.New("execution error: Can't get note \"Missing Note\"")
	provider := NewAppleNotesProviderWithExecutor("Missing Note", mockExecutor("", notFoundErr))

	_, err := provider.readRawNoteBody()
	if err == nil {
		t.Fatal("readRawNoteBody() expected error, got nil")
		return
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found', got: %v", err)
	}
}

// --- taskToNoteLine tests ---

func TestTaskToNoteLine(t *testing.T) {
	tests := []struct {
		name string
		task *core.Task
		want string
	}{
		{
			name: "complete task gets checked checkbox",
			task: newTestTask("id1", "Buy milk", core.StatusComplete, baseTime),
			want: "- [x] Buy milk",
		},
		{
			name: "todo task gets unchecked checkbox",
			task: newTestTask("id2", "Buy bread", core.StatusTodo, baseTime),
			want: "- [ ] Buy bread",
		},
		{
			name: "blocked task gets unchecked checkbox",
			task: newTestTask("id3", "Fix bug", core.StatusBlocked, baseTime),
			want: "- [ ] Fix bug",
		},
		{
			name: "in-progress task gets unchecked checkbox",
			task: newTestTask("id4", "Review PR", core.StatusInProgress, baseTime),
			want: "- [ ] Review PR",
		},
	}

	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.taskToNoteLine(tt.task)
			if got != tt.want {
				t.Errorf("taskToNoteLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- plaintextToHTML tests ---

func TestPlaintextToHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single unchecked line",
			input: "- [ ] Buy milk",
			want:  "<div>- [ ] Buy milk</div>",
		},
		{
			name:  "single checked line",
			input: "- [x] Done task",
			want:  "<div>- [x] Done task</div>",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "two lines",
			input: "Line 1\nLine 2",
			want:  "<div>Line 1</div>\n<div>Line 2</div>",
		},
		{
			name:  "empty line between content becomes br",
			input: "Line 1\n\nLine 3",
			want:  "<div>Line 1</div>\n<div><br></div>\n<div>Line 3</div>",
		},
		{
			name:  "html special characters are escaped",
			input: "Task with <html> chars",
			want:  "<div>Task with &lt;html&gt; chars</div>",
		},
		{
			name:  "quotes are escaped",
			input: `Task with "quotes"`,
			want:  "<div>Task with &#34;quotes&#34;</div>",
		},
		{
			name:  "whitespace-only input",
			input: "  ",
			want:  "",
		},
		{
			name:  "ampersand is escaped",
			input: "Tom & Jerry",
			want:  "<div>Tom &amp; Jerry</div>",
		},
	}

	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.plaintextToHTML(tt.input)
			if got != tt.want {
				t.Errorf("plaintextToHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- SaveTask write operation tests ---

func TestAppleNotesProvider_SaveTask_Success(t *testing.T) {
	noteBody := "- [ ] Buy milk\n- [x] Buy eggs\n- [ ] Buy bread"
	// Compute the ID for line 0 ("Buy milk") using the same algorithm as parseNoteBody
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	parsed := provider.parseNoteBody(noteBody)
	if len(parsed) != 3 {
		t.Fatalf("setup: expected 3 tasks, got %d", len(parsed))
	}
	taskToUpdate := parsed[0] // "Buy milk" at line 0
	taskToUpdate.Status = core.StatusComplete

	// Create recording executor: read returns noteBody, write accepts
	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider = NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.SaveTask(taskToUpdate)
	if err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	// Verify exactly 2 osascript calls: read then write
	if len(rec.scripts) != 2 {
		t.Fatalf("expected 2 osascript calls, got %d", len(rec.scripts))
	}
	if !strings.Contains(rec.scripts[0], "get plaintext text") {
		t.Errorf("first call should be read, got: %s", rec.scripts[0])
	}
	if !strings.Contains(rec.scripts[1], "set body") {
		t.Errorf("second call should be write, got: %s", rec.scripts[1])
	}
	// Verify the write contains the updated checkbox
	if !strings.Contains(rec.scripts[1], "- [x] Buy milk") {
		t.Errorf("write script should contain '- [x] Buy milk', got: %s", rec.scripts[1])
	}
}

func TestAppleNotesProvider_SaveTask_NotFound_Appends(t *testing.T) {
	noteBody := "- [ ] Buy milk"
	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	// Create a task with an ID that doesn't match any existing line
	newTask := newTestTask("nonexistent-id", "New task appended", core.StatusTodo, baseTime)
	err := provider.SaveTask(newTask)
	if err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	// Verify write contains both original and appended task
	if len(rec.scripts) < 2 {
		t.Fatalf("expected at least 2 osascript calls, got %d", len(rec.scripts))
	}
	writeScript := rec.scripts[1]
	if !strings.Contains(writeScript, "Buy milk") {
		t.Error("write should preserve existing 'Buy milk' line")
	}
	if !strings.Contains(writeScript, "New task appended") {
		t.Error("write should contain appended 'New task appended' line")
	}
}

func TestAppleNotesProvider_SaveTask_Complete_UpdatesCheckbox(t *testing.T) {
	noteBody := "- [ ] Buy milk\n- [ ] Buy eggs"
	// Get the ID for "Buy eggs" (line index 1)
	tempProvider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	parsed := tempProvider.parseNoteBody(noteBody)
	eggTask := parsed[1]
	eggTask.Status = core.StatusComplete

	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.SaveTask(eggTask)
	if err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	// Verify the write replaces the correct line with checked checkbox
	writeScript := rec.scripts[1]
	if !strings.Contains(writeScript, "- [x] Buy eggs") {
		t.Errorf("write should contain '- [x] Buy eggs', got: %s", writeScript)
	}
	// Original first task should be unchanged
	if !strings.Contains(writeScript, "- [ ] Buy milk") {
		t.Errorf("write should preserve '- [ ] Buy milk', got: %s", writeScript)
	}
}

func TestAppleNotesProvider_SaveTask_CorrectLineMatched(t *testing.T) {
	// 5-line note with a target at line index 2
	noteBody := "Header\n- [ ] Task A\n- [ ] Task B\n- [ ] Task C\n- [ ] Task D"
	tempProvider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	parsed := tempProvider.parseNoteBody(noteBody)
	if len(parsed) != 5 {
		t.Fatalf("setup: expected 5 tasks, got %d", len(parsed))
	}

	// Modify Task B (index 2 in the original body)
	taskB := parsed[2]
	taskB.Status = core.StatusComplete

	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.SaveTask(taskB)
	if err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	writeScript := rec.scripts[1]
	// Line 2 (Task B) should be updated
	if !strings.Contains(writeScript, "- [x] Task B") {
		t.Errorf("write should contain '- [x] Task B', got: %s", writeScript)
	}
	// All other lines should be unchanged
	for _, unchanged := range []string{"Header", "- [ ] Task A", "- [ ] Task C", "- [ ] Task D"} {
		if !strings.Contains(writeScript, unchanged) {
			t.Errorf("write should preserve %q, got: %s", unchanged, writeScript)
		}
	}
}

func TestAppleNotesProvider_SaveTask_SpecialCharacters(t *testing.T) {
	noteBody := `- [ ] Buy "fancy" milk`
	tempProvider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	parsed := tempProvider.parseNoteBody(noteBody)
	if len(parsed) != 1 {
		t.Fatalf("setup: expected 1 task, got %d", len(parsed))
	}

	taskWithQuotes := parsed[0]
	taskWithQuotes.Status = core.StatusComplete

	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.SaveTask(taskWithQuotes)
	if err != nil {
		t.Fatalf("SaveTask() unexpected error: %v", err)
	}

	// Verify the write script properly handles the quotes
	if len(rec.scripts) < 2 {
		t.Fatalf("expected at least 2 scripts, got %d", len(rec.scripts))
	}
	writeScript := rec.scripts[1]
	// The HTML body should contain the task with quotes (escaped for HTML)
	if !strings.Contains(writeScript, "fancy") {
		t.Errorf("write should contain task text with 'fancy', got: %s", writeScript)
	}
}

func TestAppleNotesProvider_SaveTask_WriteFails(t *testing.T) {
	noteBody := "- [ ] Buy milk"
	writeErr := errors.New("write failed: permission denied")

	executor, _ := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: writeErr},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	task := newTestTask("id1", "Buy milk", core.StatusComplete, baseTime)
	err := provider.SaveTask(task)
	if err == nil {
		t.Fatal("SaveTask() expected error when write fails, got nil")
	}
}

func TestAppleNotesProvider_SaveTask_ReadFails(t *testing.T) {
	readErr := errors.New("execution error: Can't get note \"TestNote\"")
	executor, _ := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: "", err: readErr},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	task := newTestTask("id1", "Buy milk", core.StatusComplete, baseTime)
	err := provider.SaveTask(task)
	if err == nil {
		t.Fatal("SaveTask() expected error when read fails, got nil")
	}
}

// --- SaveTasks batch tests ---

func TestAppleNotesProvider_SaveTasks_BatchUpdate(t *testing.T) {
	noteBody := "- [ ] Buy milk\n- [ ] Buy eggs\n- [ ] Buy bread"
	tempProvider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	parsed := tempProvider.parseNoteBody(noteBody)

	// Mark first and third tasks as complete
	parsed[0].Status = core.StatusComplete
	parsed[2].Status = core.StatusComplete

	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.SaveTasks([]*core.Task{parsed[0], parsed[2]})
	if err != nil {
		t.Fatalf("SaveTasks() unexpected error: %v", err)
	}

	// Should be exactly 2 osascript calls (1 read + 1 write)
	if len(rec.scripts) != 2 {
		t.Fatalf("SaveTasks() should do single read-modify-write, got %d calls", len(rec.scripts))
	}

	writeScript := rec.scripts[1]
	if !strings.Contains(writeScript, "- [x] Buy milk") {
		t.Error("write should contain '- [x] Buy milk'")
	}
	if !strings.Contains(writeScript, "- [ ] Buy eggs") {
		t.Error("write should preserve '- [ ] Buy eggs' (not in update list)")
	}
	if !strings.Contains(writeScript, "- [x] Buy bread") {
		t.Error("write should contain '- [x] Buy bread'")
	}
}

func TestAppleNotesProvider_SaveTasks_EmptyInput(t *testing.T) {
	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.SaveTasks([]*core.Task{})
	if err != nil {
		t.Fatalf("SaveTasks() with empty input unexpected error: %v", err)
	}

	// Empty input should be a no-op — no osascript calls
	if len(rec.scripts) != 0 {
		t.Errorf("SaveTasks([]) should make 0 osascript calls, got %d", len(rec.scripts))
	}
}

// --- DeleteTask tests ---

func TestAppleNotesProvider_DeleteTask_RemovesLine(t *testing.T) {
	noteBody := "- [ ] Buy milk\n- [x] Buy eggs\n- [ ] Buy bread"
	tempProvider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))
	parsed := tempProvider.parseNoteBody(noteBody)
	eggTaskID := parsed[1].ID // "Buy eggs"

	executor, rec := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.DeleteTask(eggTaskID)
	if err != nil {
		t.Fatalf("DeleteTask() unexpected error: %v", err)
	}

	writeScript := rec.scripts[1]
	if strings.Contains(writeScript, "Buy eggs") {
		t.Error("write should NOT contain 'Buy eggs' after deletion")
	}
	if !strings.Contains(writeScript, "Buy milk") {
		t.Error("write should preserve 'Buy milk'")
	}
	if !strings.Contains(writeScript, "Buy bread") {
		t.Error("write should preserve 'Buy bread'")
	}
}

// --- Verify core.ErrReadOnly is no longer returned (Story 2.4 replaces read-only with real writes) ---

func TestAppleNotesProvider_SaveTask_NoLongerReadOnly(t *testing.T) {
	noteBody := "- [ ] Test task"
	executor, _ := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	task := newTestTask("any-id", "Test task", core.StatusTodo, baseTime)
	err := provider.SaveTask(task)
	if errors.Is(err, core.ErrReadOnly) {
		t.Error("SaveTask() should no longer return core.ErrReadOnly after Story 2.4")
	}
}

func TestAppleNotesProvider_SaveTasks_NoLongerReadOnly(t *testing.T) {
	noteBody := "- [ ] Test task"
	executor, _ := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	tasks := []*core.Task{newTestTask("any-id", "Test task", core.StatusTodo, baseTime)}
	err := provider.SaveTasks(tasks)
	if errors.Is(err, core.ErrReadOnly) {
		t.Error("SaveTasks() should no longer return core.ErrReadOnly after Story 2.4")
	}
}

func TestAppleNotesProvider_DeleteTask_NoLongerReadOnly(t *testing.T) {
	noteBody := "- [ ] Test task"
	executor, _ := recordingExecutor(map[string]struct {
		output string
		err    error
	}{
		"get plaintext text": {output: noteBody, err: nil},
		"set body":           {output: "", err: nil},
	})
	provider := NewAppleNotesProviderWithExecutor("TestNote", executor)

	err := provider.DeleteTask("some-id")
	if errors.Is(err, core.ErrReadOnly) {
		t.Error("DeleteTask() should no longer return core.ErrReadOnly after Story 2.4")
	}
}

func TestAppleNotesProvider_MarkComplete_ReturnsErrReadOnly(t *testing.T) {
	provider := NewAppleNotesProviderWithExecutor("TestNote", mockExecutor("", nil))

	err := provider.MarkComplete("aaa")
	if !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("MarkComplete() error = %v, want core.ErrReadOnly", err)
	}
}
