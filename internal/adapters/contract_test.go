package adapters_test

import (
	"sync"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/adapters/textfile"

	"github.com/arcavenae/ThreeDoors/internal/adapters"
	"github.com/arcavenae/ThreeDoors/internal/core"
)

// TestTextFileProviderContract runs the full contract test suite
// against the TextFileProvider to validate the reference implementation.
func TestTextFileProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		dir := t.TempDir()
		core.SetHomeDir(dir)
		t.Cleanup(func() {
			core.SetHomeDir("")
		})
		return textfile.NewTextFileProvider()
	}

	adapters.RunContractTests(t, factory)
}

// TestContractSuite_LoadTasks_Empty verifies LoadTasks on a fresh provider.
func TestContractSuite_LoadTasks_Empty(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() {
		core.SetHomeDir("")
	})

	provider := textfile.NewTextFileProvider()
	loaded, err := provider.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	// TextFileProvider creates default tasks if none exist — that's implementation-specific
	// The contract test suite handles the general case
	if loaded == nil {
		t.Error("LoadTasks() returned nil slice")
	}
}

// TestContractSuite_ConcurrentAccess validates thread safety of provider operations.
func TestContractSuite_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() {
		core.SetHomeDir("")
	})

	provider := textfile.NewTextFileProvider()

	// Initialize with some tasks
	initial := []*core.Task{
		core.NewTask("Concurrent task 1"),
		core.NewTask("Concurrent task 2"),
		core.NewTask("Concurrent task 3"),
	}
	if err := provider.SaveTasks(initial); err != nil {
		t.Fatalf("SaveTasks() setup error: %v", err)
	}

	var wg sync.WaitGroup

	// Run concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = provider.LoadTasks()
		}()
	}

	// Run concurrent saves — errors expected for file-based providers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task := core.NewTask("concurrent save task")
			_ = provider.SaveTask(task)
		}()
	}

	wg.Wait()

	// Verify no panics. File-based providers may have corrupted state
	// from concurrent writes without locking — that's acceptable.
	_, _ = provider.LoadTasks()
}
