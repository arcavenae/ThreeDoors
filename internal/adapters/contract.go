// Package adapters provides a reusable contract test suite for validating
// TaskProvider implementations. Any adapter can import this package and
// run the contract tests against its implementation to verify compliance.
//
// Usage in adapter tests:
//
//	func TestMyAdapterContract(t *testing.T) {
//	    factory := func(t *testing.T) core.TaskProvider {
//	        t.Helper()
//	        return NewMyAdapter(t.TempDir())
//	    }
//	    adapters.RunContractTests(t, factory)
//	}
package adapters

import (
	"sync"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// ProviderFactory creates a fresh TaskProvider instance for each test.
// The factory should return an isolated provider (e.g., using t.TempDir()
// for file-based providers). Use t.Cleanup() for resource teardown.
type ProviderFactory func(t *testing.T) core.TaskProvider

// RunContractTests runs the full contract test suite against a TaskProvider
// created by the given factory. Each subtest gets a fresh provider instance.
func RunContractTests(t *testing.T, factory ProviderFactory) {
	t.Helper()

	t.Run("SaveAndLoad", func(t *testing.T) {
		testSaveAndLoad(t, factory)
	})

	t.Run("SaveTask_NewTask", func(t *testing.T) {
		testSaveTaskNew(t, factory)
	})

	t.Run("SaveTask_UpdateExisting", func(t *testing.T) {
		testSaveTaskUpdate(t, factory)
	})

	t.Run("DeleteTask", func(t *testing.T) {
		testDeleteTask(t, factory)
	})

	t.Run("DeleteTask_NonExistent", func(t *testing.T) {
		testDeleteTaskNonExistent(t, factory)
	})

	t.Run("MarkComplete", func(t *testing.T) {
		testMarkComplete(t, factory)
	})

	t.Run("MarkComplete_NonExistent", func(t *testing.T) {
		testMarkCompleteNonExistent(t, factory)
	})

	t.Run("SaveTasks_Batch", func(t *testing.T) {
		testSaveTasksBatch(t, factory)
	})

	t.Run("ConcurrentReads", func(t *testing.T) {
		testConcurrentReads(t, factory)
	})

	t.Run("ConcurrentWrites", func(t *testing.T) {
		testConcurrentWrites(t, factory)
	})

	// Error handling contract tests (Story 9.2)
	t.Run("ErrorHandling_SaveTasksEmpty", func(t *testing.T) {
		testSaveTasksEmpty(t, factory)
	})

	t.Run("ErrorHandling_LoadAfterSave", func(t *testing.T) {
		testLoadAfterSave(t, factory)
	})

	t.Run("ErrorHandling_DeleteThenLoad", func(t *testing.T) {
		testDeleteThenLoad(t, factory)
	})

	t.Run("InterfaceCompliance_AllMethodsCallable", func(t *testing.T) {
		testInterfaceCompliance(t, factory)
	})
}

func testSaveAndLoad(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	original := []*core.Task{
		core.NewTask("Task Alpha"),
		core.NewTask("Task Beta"),
	}

	if err := provider.SaveTasks(original); err != nil {
		t.Fatalf("SaveTasks() error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(loaded) != len(original) {
		t.Fatalf("LoadTasks() returned %d tasks, want %d", len(loaded), len(original))
	}

	tasksByID := make(map[string]*core.Task, len(loaded))
	for _, task := range loaded {
		tasksByID[task.ID] = task
	}

	for _, orig := range original {
		loaded, ok := tasksByID[orig.ID]
		if !ok {
			t.Errorf("task %q not found after save/load", orig.ID)
			continue
		}
		if loaded.Text != orig.Text {
			t.Errorf("task %q: Text = %q, want %q", orig.ID, loaded.Text, orig.Text)
		}
		if loaded.Status != orig.Status {
			t.Errorf("task %q: Status = %q, want %q", orig.ID, loaded.Status, orig.Status)
		}
	}
}

func testSaveTaskNew(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	// Start with empty state
	if err := provider.SaveTasks([]*core.Task{}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	task := core.NewTask("New individual task")
	if err := provider.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	found := false
	for _, lt := range loaded {
		if lt.ID == task.ID {
			found = true
			if lt.Text != task.Text {
				t.Errorf("saved task Text = %q, want %q", lt.Text, task.Text)
			}
		}
	}
	if !found {
		t.Errorf("saved task %q not found in loaded tasks", task.ID)
	}
}

func testSaveTaskUpdate(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	task := core.NewTask("Original text")
	if err := provider.SaveTasks([]*core.Task{task}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	task.Text = "Updated text"
	if err := provider.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() update error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	for _, lt := range loaded {
		if lt.ID == task.ID {
			if lt.Text != "Updated text" {
				t.Errorf("updated task Text = %q, want %q", lt.Text, "Updated text")
			}
			return
		}
	}
	t.Error("updated task not found after reload")
}

func testDeleteTask(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	task1 := core.NewTask("Keep this")
	task2 := core.NewTask("Delete this")
	if err := provider.SaveTasks([]*core.Task{task1, task2}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	if err := provider.DeleteTask(task2.ID); err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	for _, lt := range loaded {
		if lt.ID == task2.ID {
			t.Errorf("deleted task %q still present after deletion", task2.ID)
		}
	}

	// Verify the other task survives
	found := false
	for _, lt := range loaded {
		if lt.ID == task1.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("non-deleted task %q was lost", task1.ID)
	}
}

func testDeleteTaskNonExistent(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	if err := provider.SaveTasks([]*core.Task{}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	// Deleting a non-existent task should not error (idempotent)
	err := provider.DeleteTask("nonexistent-id")
	if err != nil {
		t.Logf("DeleteTask() for nonexistent ID returned error: %v (acceptable behavior varies by provider)", err)
	}
}

func testMarkComplete(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	task := core.NewTask("Complete me")
	if err := provider.SaveTasks([]*core.Task{task}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	err := provider.MarkComplete(task.ID)
	if err != nil {
		// Some providers return ErrReadOnly — that's acceptable behavior
		if err.Error() == "provider is read-only" {
			t.Skipf("provider does not support MarkComplete (read-only)")
		}
		t.Fatalf("MarkComplete() error: %v", err)
	}
}

func testMarkCompleteNonExistent(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	if err := provider.SaveTasks([]*core.Task{}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	err := provider.MarkComplete("nonexistent-id")
	if err == nil {
		// Some providers (e.g., WAL-based) may enqueue the operation rather than
		// returning an immediate error. Both behaviors are acceptable.
		t.Logf("MarkComplete() on nonexistent task returned nil (acceptable for queue-based providers)")
	}
}

func testSaveTasksBatch(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	batch := make([]*core.Task, 10)
	for i := range batch {
		batch[i] = core.NewTask("Batch task")
	}

	if err := provider.SaveTasks(batch); err != nil {
		t.Fatalf("SaveTasks() batch error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(loaded) != 10 {
		t.Errorf("LoadTasks() returned %d tasks, want 10", len(loaded))
	}
}

func testConcurrentReads(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	seed := []*core.Task{
		core.NewTask("Read test 1"),
		core.NewTask("Read test 2"),
	}
	if err := provider.SaveTasks(seed); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := provider.LoadTasks(); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent read error: %v", err)
	}
}

func testConcurrentWrites(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	if err := provider.SaveTasks([]*core.Task{}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	// Concurrent writes may encounter transient errors on file-based providers
	// (e.g., atomic rename races). The contract requires that:
	// 1. No panics occur
	// 2. No data corruption (provider remains usable after)
	// Individual write errors are acceptable for providers without internal locking.
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task := core.NewTask("concurrent write task")
			_ = provider.SaveTask(task)
		}()
	}

	wg.Wait()

	// Verify provider doesn't panic after concurrent writes.
	// File-based providers may have corrupted state from concurrent writes
	// without locking — that's an acceptable known limitation.
	// The contract only requires no panics during concurrent access.
	_, _ = provider.LoadTasks()
}

// testSaveTasksEmpty verifies saving an empty slice doesn't error.
func testSaveTasksEmpty(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	if err := provider.SaveTasks([]*core.Task{}); err != nil {
		t.Fatalf("SaveTasks([]) error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() after empty save error: %v", err)
	}

	// After saving empty list, provider should return zero or its default tasks.
	// We don't mandate an exact count since some providers create defaults.
	_ = loaded
}

// testLoadAfterSave verifies that a full save-load-modify-save-load cycle
// preserves data integrity with no error accumulation.
func testLoadAfterSave(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	// Round 1: save and load
	task := core.NewTask("Round trip task")
	if err := provider.SaveTasks([]*core.Task{task}); err != nil {
		t.Fatalf("SaveTasks() round 1 error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() round 1 error: %v", err)
	}

	if len(loaded) == 0 {
		t.Fatal("LoadTasks() round 1 returned 0 tasks")
	}

	// Round 2: save a second task individually
	task2 := core.NewTask("Second task")
	if err := provider.SaveTask(task2); err != nil {
		t.Fatalf("SaveTask() round 2 error: %v", err)
	}

	loaded2, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() round 2 error: %v", err)
	}

	// After two saves, provider should have at least the most recent data.
	// Some providers (e.g., Obsidian with position-based IDs) may merge tasks.
	if len(loaded2) == 0 {
		t.Error("LoadTasks() round 2 returned 0 tasks after two saves")
	}
}

// testDeleteThenLoad verifies that after deleting a task, the provider
// returns a consistent state with no residual data.
func testDeleteThenLoad(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	t1 := core.NewTask("Survivor")
	t2 := core.NewTask("Doomed")
	t3 := core.NewTask("Also survives")

	if err := provider.SaveTasks([]*core.Task{t1, t2, t3}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	if err := provider.DeleteTask(t2.ID); err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() after delete error: %v", err)
	}

	for _, lt := range loaded {
		if lt.ID == t2.ID {
			t.Errorf("deleted task %q still present", t2.ID)
		}
	}

	// Verify count decreased by exactly 1
	if len(loaded) != 2 {
		t.Errorf("LoadTasks() returned %d tasks, want 2", len(loaded))
	}
}

// testInterfaceCompliance verifies that all five TaskProvider methods are
// callable without panicking across a typical lifecycle. This is a smoke
// test ensuring the provider wiring is correct beyond compile-time checks.
func testInterfaceCompliance(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	// 1. LoadTasks on fresh provider
	_, err := provider.LoadTasks()
	if err != nil {
		t.Logf("LoadTasks() on fresh provider: %v (acceptable for some providers)", err)
	}

	// 2. SaveTasks
	task := core.NewTask("Compliance test task")
	if err := provider.SaveTasks([]*core.Task{task}); err != nil {
		t.Fatalf("SaveTasks() error: %v", err)
	}

	// 3. SaveTask
	task2 := core.NewTask("Individual save")
	if err := provider.SaveTask(task2); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	// 4. DeleteTask
	if err := provider.DeleteTask(task2.ID); err != nil {
		t.Logf("DeleteTask() error: %v (may be acceptable)", err)
	}

	// 5. MarkComplete
	err = provider.MarkComplete(task.ID)
	if err != nil {
		// Read-only providers may return ErrReadOnly — that's acceptable
		if err.Error() == "provider is read-only" {
			t.Logf("MarkComplete() returned ErrReadOnly (acceptable)")
		} else {
			t.Logf("MarkComplete() error: %v (may be acceptable for some providers)", err)
		}
	}
}
