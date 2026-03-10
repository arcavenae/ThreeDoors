package textfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func setupTestProvider(t *testing.T) (*TextFileProvider, func()) {
	t.Helper()
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	cleanup := func() { core.SetHomeDir("") }
	return NewTextFileProvider(), cleanup
}

func TestNewTextFileProvider(t *testing.T) {
	t.Parallel()
	p := NewTextFileProvider()
	if p == nil {
		t.Fatal("expected non-nil provider")
		return
	}
}

func TestTextFileProvider_IsTextFileBackend(t *testing.T) {
	t.Parallel()
	p := NewTextFileProvider()
	if !p.IsTextFileBackend() {
		t.Error("expected IsTextFileBackend to return true")
	}
}

func TestTextFileProvider_Name(t *testing.T) {
	t.Parallel()
	p := NewTextFileProvider()
	if p.Name() != "textfile" {
		t.Errorf("expected name %q, got %q", "textfile", p.Name())
	}
}

func TestTextFileProvider_Watch(t *testing.T) {
	t.Parallel()
	p := NewTextFileProvider()
	if p.Watch() != nil {
		t.Error("expected Watch to return nil")
	}
}

func TestTextFileProvider_LoadTasks(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(tasks) != len(defaultTaskTexts) {
		t.Errorf("expected %d default tasks, got %d", len(defaultTaskTexts), len(tasks))
	}
}

func TestTextFileProvider_SaveTask_NewTask(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	// Load defaults first
	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	newTask := core.NewTask("Brand new task")
	if err := p.SaveTask(newTask); err != nil {
		t.Fatalf("SaveTask: %v", err)
	}

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks after save: %v", err)
	}

	found := false
	for _, task := range tasks {
		if task.ID == newTask.ID {
			found = true
			if task.Text != "Brand new task" {
				t.Errorf("expected text %q, got %q", "Brand new task", task.Text)
			}
			break
		}
	}
	if !found {
		t.Error("newly saved task not found in loaded tasks")
	}
}

func TestTextFileProvider_SaveTask_UpdateExisting(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	target := tasks[0]
	_ = target.UpdateStatus(core.StatusInProgress)

	if err := p.SaveTask(target); err != nil {
		t.Fatalf("SaveTask: %v", err)
	}

	reloaded, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks after update: %v", err)
	}

	for _, task := range reloaded {
		if task.ID == target.ID {
			if task.Status != core.StatusInProgress {
				t.Errorf("expected status %q, got %q", core.StatusInProgress, task.Status)
			}
			return
		}
	}
	t.Error("updated task not found")
}

func TestTextFileProvider_SaveTasks(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	tasks := []*core.Task{
		core.NewTask("Task X"),
		core.NewTask("Task Y"),
	}

	if err := p.SaveTasks(tasks); err != nil {
		t.Fatalf("SaveTasks: %v", err)
	}

	loaded, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(loaded))
	}
}

func TestTextFileProvider_DeleteTask(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	deleteID := tasks[0].ID
	originalCount := len(tasks)

	if err := p.DeleteTask(deleteID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	remaining, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks after delete: %v", err)
	}
	if len(remaining) != originalCount-1 {
		t.Errorf("expected %d tasks, got %d", originalCount-1, len(remaining))
	}
	for _, task := range remaining {
		if task.ID == deleteID {
			t.Error("deleted task still present")
		}
	}
}

func TestTextFileProvider_DeleteTask_NonExistent(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	// Deleting a non-existent ID should succeed (no-op)
	if err := p.DeleteTask("nonexistent-id"); err != nil {
		t.Fatalf("DeleteTask with nonexistent ID: %v", err)
	}
}

func TestTextFileProvider_HealthCheck_OK(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	// Create the task file first
	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	result := p.HealthCheck()
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 health check item, got %d", len(result.Items))
	}
	if result.Items[0].Status != core.HealthOK {
		t.Errorf("expected HealthOK, got %v", result.Items[0].Status)
	}
}

func TestTextFileProvider_HealthCheck_Fail(t *testing.T) {
	p := NewTextFileProvider()
	// Set home to a non-existent directory
	core.SetHomeDir("/nonexistent/path/that/does/not/exist")
	defer core.SetHomeDir("")

	result := p.HealthCheck()
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 health check item, got %d", len(result.Items))
	}
	if result.Items[0].Status != core.HealthFail {
		t.Errorf("expected HealthFail, got %v", result.Items[0].Status)
	}
	if result.Items[0].Suggestion == "" {
		t.Error("expected non-empty suggestion on failure")
	}
}

func TestTextFileProvider_MarkComplete(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	target := tasks[0]
	targetID := target.ID
	originalCount := len(tasks)

	if err := p.MarkComplete(targetID); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	// Task should be removed from active tasks
	remaining, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks after complete: %v", err)
	}
	if len(remaining) != originalCount-1 {
		t.Errorf("expected %d active tasks, got %d", originalCount-1, len(remaining))
	}

	// Task should be in completed.txt
	configPath, _ := core.EnsureConfigDir()
	completedPath := filepath.Join(configPath, completedFile)
	data, err := os.ReadFile(completedPath)
	if err != nil {
		t.Fatalf("read completed file: %v", err)
	}
	if len(data) == 0 {
		t.Error("completed file should not be empty")
	}
}

func TestTextFileProvider_MarkComplete_NotFound(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	_, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	err = p.MarkComplete("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
}

func TestTextFileProvider_MarkComplete_InvalidTransition(t *testing.T) {
	p, cleanup := setupTestProvider(t)
	defer cleanup()

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	// Set a task to "deferred" — deferred→complete is not a valid transition
	target := tasks[0]
	_ = target.UpdateStatus(core.StatusDeferred)
	if err := p.SaveTask(target); err != nil {
		t.Fatalf("SaveTask: %v", err)
	}

	err = p.MarkComplete(target.ID)
	if err == nil {
		t.Fatal("expected error for invalid transition from deferred to complete")
	}
}
