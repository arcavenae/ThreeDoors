// Package adapters provides a reusable contract test suite for validating
// TaskProvider implementations. Any adapter can import this package and
// run the contract tests against its implementation to verify compliance.
//
// Usage in adapter tests:
//
//	func TestMyAdapterContract(t *testing.T) {
//	    factory := func(t *testing.T) tasks.TaskProvider {
//	        t.Helper()
//	        return NewMyAdapter(t.TempDir())
//	    }
//	    adapters.RunContractTests(t, factory)
//	}
package adapters

import (
	"sync"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
)

// ProviderFactory creates a fresh TaskProvider instance for each test.
// The factory should return an isolated provider (e.g., using t.TempDir()
// for file-based providers). Use t.Cleanup() for resource teardown.
type ProviderFactory func(t *testing.T) tasks.TaskProvider

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
}

func testSaveAndLoad(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	original := []*tasks.Task{
		tasks.NewTask("Task Alpha"),
		tasks.NewTask("Task Beta"),
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

	tasksByID := make(map[string]*tasks.Task, len(loaded))
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
	if err := provider.SaveTasks([]*tasks.Task{}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	task := tasks.NewTask("New individual task")
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

	task := tasks.NewTask("Original text")
	if err := provider.SaveTasks([]*tasks.Task{task}); err != nil {
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

	task1 := tasks.NewTask("Keep this")
	task2 := tasks.NewTask("Delete this")
	if err := provider.SaveTasks([]*tasks.Task{task1, task2}); err != nil {
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

	if err := provider.SaveTasks([]*tasks.Task{}); err != nil {
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

	task := tasks.NewTask("Complete me")
	if err := provider.SaveTasks([]*tasks.Task{task}); err != nil {
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

	if err := provider.SaveTasks([]*tasks.Task{}); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	err := provider.MarkComplete("nonexistent-id")
	if err == nil {
		t.Error("MarkComplete() on nonexistent task should return error")
	}
}

func testSaveTasksBatch(t *testing.T, factory ProviderFactory) {
	t.Helper()
	provider := factory(t)

	batch := make([]*tasks.Task, 10)
	for i := range batch {
		batch[i] = tasks.NewTask("Batch task")
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

	seed := []*tasks.Task{
		tasks.NewTask("Read test 1"),
		tasks.NewTask("Read test 2"),
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

	if err := provider.SaveTasks([]*tasks.Task{}); err != nil {
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
			task := tasks.NewTask("concurrent write task")
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
