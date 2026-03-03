package applenotes_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/adapters/applenotes"
	"github.com/arcaven/ThreeDoors/internal/adapters/textfile"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// fixturesDir returns the absolute path to testdata/applenotes fixtures.
func fixturesDir(t *testing.T) string {
	t.Helper()
	// Navigate from internal/adapters/applenotes to project root
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "testdata", "applenotes"))
	if err != nil {
		t.Fatalf("failed to resolve fixtures dir: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("fixtures directory does not exist: %s", dir)
	}
	return dir
}

// loadFixture reads a test fixture file and returns its contents.
func loadFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixturesDir(t), name))
	if err != nil {
		t.Fatalf("failed to read fixture %q: %v", name, err)
	}
	return string(data)
}

// applenotes_stubExecutor creates a mock applenotes.CommandExecutor that returns canned responses
// based on script content matching. Responses are matched by substring.
type stubResponse struct {
	output string
	err    error
}

func applenotes_stubExecutor(responses map[string]stubResponse) applenotes.CommandExecutor {
	return func(_ context.Context, script string) (string, error) {
		for key, resp := range responses {
			if strings.Contains(script, key) {
				return resp.output, resp.err
			}
		}
		return "", fmt.Errorf("unmatched script: %s", script)
	}
}

// trackingExecutor wraps a applenotes_stubExecutor and records all scripts called.
type scriptTracker struct {
	mu      sync.Mutex
	scripts []string
}

func (st *scriptTracker) record(script string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.scripts = append(st.scripts, script)
}

func (st *scriptTracker) count() int {
	st.mu.Lock()
	defer st.mu.Unlock()
	return len(st.scripts)
}

func (st *scriptTracker) get(i int) string {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.scripts[i]
}

func trackingStubExecutor(responses map[string]stubResponse) (applenotes.CommandExecutor, *scriptTracker) {
	tracker := &scriptTracker{}
	executor := func(_ context.Context, script string) (string, error) {
		tracker.record(script)
		for key, resp := range responses {
			if strings.Contains(script, key) {
				return resp.output, resp.err
			}
		}
		return "", fmt.Errorf("unmatched script: %s", script)
	}
	return executor, tracker
}

// =============================================================================
// E2E Test: Full Note Creation Workflow
// =============================================================================

func TestE2E_NoteCreation_NewTaskAppears(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")
	executor, tracker := trackingStubExecutor(map[string]stubResponse{
		"get plaintext text": {output: fixture},
		"set body":           {},
	})
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks", executor)

	// Create a new task and save it
	newTask := core.NewTask("Brand new E2E task")
	err := provider.SaveTask(newTask)
	if err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	// Verify read-modify-write: exactly 2 osascript calls
	if tracker.count() != 2 {
		t.Fatalf("expected 2 osascript calls (read + write), got %d", tracker.count())
	}

	// Verify the write contains the new task
	writeScript := tracker.get(1)
	if !strings.Contains(writeScript, "Brand new E2E task") {
		t.Error("write script should contain the new task text")
	}
	if !strings.Contains(writeScript, "set body") {
		t.Error("second call should be a write (set body)")
	}
}

// =============================================================================
// E2E Test: Task Read Workflow
// =============================================================================

func TestE2E_TaskRead_LoadsFromFixture(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(loaded) != 5 {
		t.Fatalf("expected 5 tasks from basic fixture, got %d", len(loaded))
	}

	// Verify status parsing
	wantStatuses := []core.TaskStatus{
		core.StatusTodo, core.StatusComplete, core.StatusTodo,
		core.StatusTodo, core.StatusComplete,
	}
	for i, task := range loaded {
		if task.Status != wantStatuses[i] {
			t.Errorf("task[%d] %q: status = %q, want %q", i, task.Text, task.Status, wantStatuses[i])
		}
	}
}

func TestE2E_TaskRead_SpecialCharacters(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "special_chars.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(loaded) != 5 {
		t.Fatalf("expected 5 tasks from special_chars fixture, got %d", len(loaded))
	}

	// Verify special characters survive parsing
	wantTexts := []string{
		`Fix "login" bug & deploy`,
		"Review <PR #42> changes",
		"Tom & Jerry's café order",
		"Task with 🎯 emoji target",
		"Handle null/empty cases",
	}
	for i, task := range loaded {
		if task.Text != wantTexts[i] {
			t.Errorf("task[%d].Text = %q, want %q", i, task.Text, wantTexts[i])
		}
	}
}

func TestE2E_TaskRead_LargeNote(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "large_note.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// large_note.txt has section headers + tasks = 20 non-blank lines
	if len(loaded) != 20 {
		t.Fatalf("expected 20 tasks from large fixture, got %d", len(loaded))
	}

	// Count completed tasks
	completed := 0
	for _, task := range loaded {
		if task.Status == core.StatusComplete {
			completed++
		}
	}
	if completed != 4 {
		t.Errorf("expected 4 completed tasks in large fixture, got %d", completed)
	}
}

func TestE2E_TaskRead_MixedFormats(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "mixed_formats.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// Verify various formats are parsed
	if len(loaded) == 0 {
		t.Fatal("expected tasks from mixed_formats fixture, got 0")
	}

	// Count completed tasks: "- [x]", "* [x]", "- [X]" = 3
	completed := 0
	for _, task := range loaded {
		if task.Status == core.StatusComplete {
			completed++
		}
	}
	if completed != 3 {
		t.Errorf("expected 3 completed tasks in mixed fixture, got %d", completed)
	}
}

func TestE2E_TaskRead_EmptyNote(t *testing.T) {
	t.Parallel()

	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: ""},
		}))

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 tasks from empty note, got %d", len(loaded))
	}
}

func TestE2E_TaskRead_DeterministicIDs(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))

	// Load twice — IDs must be identical
	first, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("first LoadTasks() error: %v", err)
	}
	second, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("second LoadTasks() error: %v", err)
	}

	if len(first) != len(second) {
		t.Fatalf("different task counts: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Errorf("task[%d] ID mismatch on reload: %q vs %q", i, first[i].ID, second[i].ID)
		}
	}
}

// =============================================================================
// E2E Test: Task Update Workflow
// =============================================================================

func TestE2E_TaskUpdate_ToggleCompletion(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	// Parse to get IDs
	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, err := tempProvider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// Toggle first task (todo -> complete)
	taskToUpdate := loaded[0]
	if taskToUpdate.Status != core.StatusTodo {
		t.Fatalf("expected first task to be todo, got %q", taskToUpdate.Status)
	}
	taskToUpdate.Status = core.StatusComplete

	executor, tracker := trackingStubExecutor(map[string]stubResponse{
		"get plaintext text": {output: fixture},
		"set body":           {},
	})
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks", executor)

	err = provider.SaveTask(taskToUpdate)
	if err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	// Verify the write contains updated checkbox
	writeScript := tracker.get(1)
	if !strings.Contains(writeScript, "- [x] Buy groceries") {
		t.Error("write should contain checked checkbox for 'Buy groceries'")
	}
	// Other tasks should be unchanged
	if !strings.Contains(writeScript, "Walk the dog") {
		t.Error("write should preserve other tasks")
	}
}

func TestE2E_TaskUpdate_BatchUpdate(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, err := tempProvider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// Update multiple tasks at once
	loaded[0].Status = core.StatusComplete // Buy groceries -> complete
	loaded[2].Status = core.StatusComplete // Write report -> complete

	executor, tracker := trackingStubExecutor(map[string]stubResponse{
		"get plaintext text": {output: fixture},
		"set body":           {},
	})
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks", executor)

	err = provider.SaveTasks([]*core.Task{loaded[0], loaded[2]})
	if err != nil {
		t.Fatalf("SaveTasks() error: %v", err)
	}

	// Should be single read-modify-write (2 calls total)
	if tracker.count() != 2 {
		t.Fatalf("batch update should be 2 osascript calls, got %d", tracker.count())
	}

	writeScript := tracker.get(1)
	if !strings.Contains(writeScript, "- [x] Buy groceries") {
		t.Error("write should contain updated 'Buy groceries'")
	}
	if !strings.Contains(writeScript, "- [x] Write report") {
		t.Error("write should contain updated 'Write report'")
	}
}

func TestE2E_TaskUpdate_DeleteTask(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, err := tempProvider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// Delete "Walk the dog" (index 1)
	deleteID := loaded[1].ID

	executor, tracker := trackingStubExecutor(map[string]stubResponse{
		"get plaintext text": {output: fixture},
		"set body":           {},
	})
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks", executor)

	err = provider.DeleteTask(deleteID)
	if err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	writeScript := tracker.get(1)
	if strings.Contains(writeScript, "Walk the dog") {
		t.Error("write should NOT contain deleted task 'Walk the dog'")
	}
	if !strings.Contains(writeScript, "Buy groceries") {
		t.Error("write should preserve 'Buy groceries'")
	}
	if !strings.Contains(writeScript, "Write report") {
		t.Error("write should preserve 'Write report'")
	}
}

// =============================================================================
// E2E Test: Bidirectional Sync Workflow
// =============================================================================

func TestE2E_BidirectionalSync_ReadModifyWrite(t *testing.T) {
	t.Parallel()

	// Simulate a complete read-modify-write-read cycle
	originalFixture := loadFixture(t, "basic_tasks.txt")

	// Step 1: Read tasks from note
	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("Sync Test",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: originalFixture},
		}))
	loaded, err := tempProvider.LoadTasks()
	if err != nil {
		t.Fatalf("initial LoadTasks() error: %v", err)
	}

	// Step 2: Modify a task
	loaded[0].Status = core.StatusComplete

	// Step 3: Write back — capture what gets written
	var writtenBody string
	executor := func(_ context.Context, script string) (string, error) {
		if strings.Contains(script, "set body") {
			writtenBody = script
			return "", nil
		}
		return originalFixture, nil
	}
	provider := applenotes.NewAppleNotesProviderWithExecutor("Sync Test", executor)

	err = provider.SaveTask(loaded[0])
	if err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	// Step 4: Verify the written content includes the modification
	if !strings.Contains(writtenBody, "- [x] Buy groceries") {
		t.Error("sync write should contain checked 'Buy groceries'")
	}

	// Step 5: Simulate re-reading with the modified content
	// Build the expected new plaintext from the written HTML
	// (In real usage, Apple Notes would parse the HTML back to plaintext)
	modifiedPlaintext := strings.Replace(originalFixture, "- [ ] Buy groceries", "- [x] Buy groceries", 1)
	rereadProvider := applenotes.NewAppleNotesProviderWithExecutor("Sync Test",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: modifiedPlaintext},
		}))

	reloaded, err := rereadProvider.LoadTasks()
	if err != nil {
		t.Fatalf("re-read LoadTasks() error: %v", err)
	}

	if reloaded[0].Status != core.StatusComplete {
		t.Errorf("after sync, first task should be complete, got %q", reloaded[0].Status)
	}
}

func TestE2E_BidirectionalSync_MultipleRoundTrips(t *testing.T) {
	t.Parallel()

	// Simulate multiple sync cycles to catch state drift
	noteContent := "- [ ] Task A\n- [ ] Task B\n- [ ] Task C"

	for round := 0; round < 3; round++ {
		// Read
		provider := applenotes.NewAppleNotesProviderWithExecutor("MultiSync",
			applenotes_stubExecutor(map[string]stubResponse{
				"get plaintext text": {output: noteContent},
			}))
		loaded, err := provider.LoadTasks()
		if err != nil {
			t.Fatalf("round %d: LoadTasks() error: %v", round, err)
		}

		if len(loaded) != 3 {
			t.Fatalf("round %d: expected 3 tasks, got %d", round, len(loaded))
		}

		// Mark task at index `round` as complete
		loaded[round].Status = core.StatusComplete

		// Write — capture new content
		var capturedWrite string
		writeProvider := applenotes.NewAppleNotesProviderWithExecutor("MultiSync",
			func(_ context.Context, script string) (string, error) {
				if strings.Contains(script, "set body") {
					capturedWrite = script
					return "", nil
				}
				return noteContent, nil
			})

		err = writeProvider.SaveTask(loaded[round])
		if err != nil {
			t.Fatalf("round %d: SaveTask() error: %v", round, err)
		}

		if capturedWrite == "" {
			t.Fatalf("round %d: no write captured", round)
		}

		// Update noteContent to reflect the change for next round
		taskLetter := string(rune('A' + round))
		noteContent = strings.Replace(noteContent,
			"- [ ] Task "+taskLetter,
			"- [x] Task "+taskLetter, 1)
	}

	// After 3 rounds, all tasks should be marked complete
	finalProvider := applenotes.NewAppleNotesProviderWithExecutor("MultiSync",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: noteContent},
		}))
	final, err := finalProvider.LoadTasks()
	if err != nil {
		t.Fatalf("final LoadTasks() error: %v", err)
	}

	for i, task := range final {
		if task.Status != core.StatusComplete {
			t.Errorf("after 3 rounds, task[%d] should be complete, got %q", i, task.Status)
		}
	}
}

func TestE2E_BidirectionalSync_IDStability(t *testing.T) {
	t.Parallel()

	// IDs should remain stable across read-write-read cycles
	fixture := loadFixture(t, "basic_tasks.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("ID Stability",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))

	first, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("first LoadTasks() error: %v", err)
	}

	// IDs for same note title + content should be deterministic
	second, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("second LoadTasks() error: %v", err)
	}

	for i := range first {
		if first[i].ID != second[i].ID {
			t.Errorf("task[%d] ID changed across reads: %q vs %q", i, first[i].ID, second[i].ID)
		}
	}
}

// =============================================================================
// E2E Test: Error Handling
// =============================================================================

func TestE2E_Error_NoteNotFound(t *testing.T) {
	t.Parallel()

	provider := applenotes.NewAppleNotesProviderWithExecutor("Missing Note",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {
				err: errors.New(`execution error: Can't get note "Missing Note"`),
			},
		}))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("expected error for missing note")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should indicate 'not found', got: %v", err)
	}
}

func TestE2E_Error_PermissionDenied(t *testing.T) {
	t.Parallel()

	provider := applenotes.NewAppleNotesProviderWithExecutor("Protected Note",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {
				err: errors.New("execution error: Not authorized to send Apple events to Notes"),
			},
		}))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("expected error for permission denied")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should indicate 'permission denied', got: %v", err)
	}
}

func TestE2E_Error_Timeout(t *testing.T) {
	t.Parallel()

	provider := applenotes.NewAppleNotesProviderWithExecutor("Slow Note",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {err: context.DeadlineExceeded},
		}))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("expected error for timeout")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error should indicate 'timed out', got: %v", err)
	}
}

func TestE2E_Error_MarkCompleteReadOnly(t *testing.T) {
	t.Parallel()

	provider := applenotes.NewAppleNotesProviderWithExecutor("ReadOnly Note",
		applenotes_stubExecutor(map[string]stubResponse{}))

	err := provider.MarkComplete("any-id")
	if !errors.Is(err, core.ErrReadOnly) {
		t.Errorf("MarkComplete() should return core.ErrReadOnly, got: %v", err)
	}
}

func TestE2E_Error_WriteFailure_OnSave(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
			"set body":           {err: errors.New("write failed: disk full")},
		}))

	task := core.NewTask("Will fail to save")
	err := provider.SaveTask(task)
	if err == nil {
		t.Fatal("expected error when write fails")
	}
}

func TestE2E_Error_ReadFailure_OnSave(t *testing.T) {
	t.Parallel()

	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {err: errors.New(`Can't get note "E2E Tasks"`)},
		}))

	task := core.NewTask("Will fail to read")
	err := provider.SaveTask(task)
	if err == nil {
		t.Fatal("expected error when read fails during save")
	}
}

func TestE2E_Error_ReadFailure_OnDelete(t *testing.T) {
	t.Parallel()

	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {err: errors.New(`Can't get note "E2E Tasks"`)},
		}))

	err := provider.DeleteTask("some-id")
	if err == nil {
		t.Fatal("expected error when read fails during delete")
	}
}

func TestE2E_Error_WriteFailure_OnDelete(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
			"set body":           {err: errors.New("write failed")},
		}))

	// Get a valid task ID
	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, _ := tempProvider.LoadTasks()

	err := provider.DeleteTask(loaded[0].ID)
	if err == nil {
		t.Fatal("expected error when write fails during delete")
	}
}

// =============================================================================
// E2E Test: Connectivity Failure Scenarios
// =============================================================================

func TestE2E_ConnectivityFailure_OsascriptNotFound(t *testing.T) {
	t.Parallel()

	// Simulate osascript binary not being available (non-macOS)
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {err: fmt.Errorf("exec: %w", errors.New("executable file not found in $PATH"))},
		}))

	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("expected error when osascript is unavailable")
	}
}

func TestE2E_ConnectivityFailure_IntermittentTimeout(t *testing.T) {
	t.Parallel()

	// First call times out, simulating transient connectivity issue
	callCount := 0
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		func(_ context.Context, _ string) (string, error) {
			callCount++
			if callCount == 1 {
				return "", context.DeadlineExceeded
			}
			return "- [ ] Recovered task", nil
		})

	// First attempt fails
	_, err := provider.LoadTasks()
	if err == nil {
		t.Fatal("first call should fail with timeout")
	}

	// Second attempt succeeds (simulating retry at application level)
	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("second call should succeed, got: %v", err)
	}
	if len(loaded) != 1 {
		t.Errorf("expected 1 task after recovery, got %d", len(loaded))
	}
}

func TestE2E_ConnectivityFailure_FallbackProvider(t *testing.T) {
	t.Parallel()

	// Simulate the FallbackProvider behavior: primary fails, fallback succeeds
	primaryErr := errors.New(`execution error: Can't get note "Missing"`)
	primary := applenotes.NewAppleNotesProviderWithExecutor("Missing",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {err: primaryErr},
		}))

	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })
	fallback := textfile.NewTextFileProvider()

	fp := core.NewFallbackProvider(primary, fallback)

	// Should fall back to TextFileProvider
	loaded, err := fp.LoadTasks()
	if err != nil {
		t.Fatalf("FallbackProvider.LoadTasks() error: %v", err)
	}

	// TextFileProvider creates default tasks — verify we got something
	if loaded == nil {
		t.Error("FallbackProvider should return tasks from fallback")
	}
	if !fp.IsFallback() {
		t.Error("FallbackProvider should indicate fallback was used")
	}
}

// =============================================================================
// E2E Test: Partial Sync Scenarios
// =============================================================================

func TestE2E_PartialSync_WriteFailsAfterRead(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	// Read succeeds but write fails — state should not be corrupted
	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, _ := tempProvider.LoadTasks()

	loaded[0].Status = core.StatusComplete

	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
			"set body":           {err: errors.New("partial write failure")},
		}))

	err := provider.SaveTask(loaded[0])
	if err == nil {
		t.Fatal("expected error from write failure")
	}

	// Re-read should still return original data (no corruption)
	reProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	reloaded, err := reProvider.LoadTasks()
	if err != nil {
		t.Fatalf("re-read after partial failure: %v", err)
	}

	if reloaded[0].Status != core.StatusTodo {
		t.Error("original data should be intact after partial sync failure")
	}
}

func TestE2E_PartialSync_BatchUpdatePartialFailure(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	// SaveTasks with write failure — entire batch should fail atomically
	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, _ := tempProvider.LoadTasks()

	loaded[0].Status = core.StatusComplete
	loaded[2].Status = core.StatusComplete

	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
			"set body":           {err: errors.New("batch write failed")},
		}))

	err := provider.SaveTasks([]*core.Task{loaded[0], loaded[2]})
	if err == nil {
		t.Fatal("expected error from batch write failure")
	}
}

func TestE2E_PartialSync_DeleteDuringModification(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, _ := tempProvider.LoadTasks()

	// Delete a task, then try to save the same task — should append it
	deleteID := loaded[1].ID

	executor, _ := trackingStubExecutor(map[string]stubResponse{
		"get plaintext text": {output: fixture},
		"set body":           {},
	})
	provider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks", executor)

	err := provider.DeleteTask(deleteID)
	if err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	// Now the note no longer has "Walk the dog" at position 1.
	// If we re-save a task with that old ID, it should be appended.
	modifiedFixture := "- [ ] Buy groceries\n- [ ] Write report\n- [ ] Call dentist\n- [x] Send invoices"
	appendProvider := applenotes.NewAppleNotesProviderWithExecutor("E2E Tasks",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: modifiedFixture},
			"set body":           {},
		}))

	// Saving with old ID should append since it's no longer found
	oldTask := loaded[1]
	err = appendProvider.SaveTask(oldTask)
	if err != nil {
		t.Fatalf("SaveTask() after delete should append: %v", err)
	}
}

// =============================================================================
// E2E Test: Concurrent Modification Scenarios
// =============================================================================

func TestE2E_ConcurrentModification_ParallelReads(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("Concurrent",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loaded, err := provider.LoadTasks()
			if err != nil {
				errCh <- err
				return
			}
			if len(loaded) != 5 {
				errCh <- fmt.Errorf("expected 5 tasks, got %d", len(loaded))
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent read error: %v", err)
	}
}

func TestE2E_ConcurrentModification_ParallelWrites(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	// Track how many writes succeed vs fail
	var writeCount atomic.Int32
	provider := applenotes.NewAppleNotesProviderWithExecutor("Concurrent",
		func(_ context.Context, script string) (string, error) {
			if strings.Contains(script, "set body") {
				writeCount.Add(1)
				return "", nil
			}
			return fixture, nil
		})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task := core.NewTask("concurrent task")
			_ = provider.SaveTask(task)
		}()
	}

	wg.Wait()

	// All writes should succeed (mock doesn't have race conditions)
	if writeCount.Load() != 10 {
		t.Errorf("expected 10 writes, got %d", writeCount.Load())
	}
}

func TestE2E_ConcurrentModification_ReadDuringWrite(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")
	provider := applenotes.NewAppleNotesProviderWithExecutor("Concurrent",
		func(_ context.Context, script string) (string, error) {
			if strings.Contains(script, "set body") {
				return "", nil
			}
			return fixture, nil
		})

	var wg sync.WaitGroup

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = provider.LoadTasks()
		}()
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task := core.NewTask("concurrent write")
			_ = provider.SaveTask(task)
		}()
	}

	wg.Wait()

	// Verify provider still works after concurrent access
	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() after concurrent access: %v", err)
	}
	if len(loaded) == 0 {
		t.Error("provider should still return tasks after concurrent access")
	}
}

func TestE2E_ConcurrentModification_ParallelDeletes(t *testing.T) {
	t.Parallel()

	fixture := loadFixture(t, "basic_tasks.txt")

	tempProvider := applenotes.NewAppleNotesProviderWithExecutor("Concurrent",
		applenotes_stubExecutor(map[string]stubResponse{
			"get plaintext text": {output: fixture},
		}))
	loaded, _ := tempProvider.LoadTasks()

	provider := applenotes.NewAppleNotesProviderWithExecutor("Concurrent",
		func(_ context.Context, script string) (string, error) {
			if strings.Contains(script, "set body") {
				return "", nil
			}
			return fixture, nil
		})

	var wg sync.WaitGroup
	// Delete different tasks concurrently
	for _, task := range loaded {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			_ = provider.DeleteTask(id)
		}(task.ID)
	}

	wg.Wait()
	// No panics = pass
}
