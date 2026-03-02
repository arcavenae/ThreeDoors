package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTasks_CreatesDefaultFileWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	fm := NewFileManager(tmpDir)

	taskList, err := fm.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(taskList) != len(defaultTasks) {
		t.Errorf("LoadTasks() returned %d tasks, want %d", len(taskList), len(defaultTasks))
	}

	// Verify file was created
	tasksPath := filepath.Join(tmpDir, "tasks.txt")
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		t.Error("tasks.txt was not created")
	}
}

func TestLoadTasks_ReadsExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.txt")

	content := "Task one\nTask two\nTask three\n"
	if err := os.WriteFile(tasksPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	fm := NewFileManager(tmpDir)
	taskList, err := fm.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(taskList) != 3 {
		t.Fatalf("LoadTasks() returned %d tasks, want 3", len(taskList))
	}

	expected := []string{"Task one", "Task two", "Task three"}
	for i, want := range expected {
		if taskList[i].Text != want {
			t.Errorf("task[%d].Text = %q, want %q", i, taskList[i].Text, want)
		}
	}
}

func TestLoadTasks_SkipsBlankLines(t *testing.T) {
	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.txt")

	content := "Task one\n\n  \nTask two\n"
	if err := os.WriteFile(tasksPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	fm := NewFileManager(tmpDir)
	taskList, err := fm.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(taskList) != 2 {
		t.Errorf("LoadTasks() returned %d tasks, want 2", len(taskList))
	}
}

func TestLoadTasks_SkipsCommentLines(t *testing.T) {
	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.txt")

	content := "# This is a comment\nTask one\n# Another comment\nTask two\n"
	if err := os.WriteFile(tasksPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	fm := NewFileManager(tmpDir)
	taskList, err := fm.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(taskList) != 2 {
		t.Errorf("LoadTasks() returned %d tasks, want 2", len(taskList))
	}
}

func TestLoadTasks_TrimsWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.txt")

	content := "  Task with spaces  \n\tTask with tabs\t\n"
	if err := os.WriteFile(tasksPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	fm := NewFileManager(tmpDir)
	taskList, err := fm.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(taskList) != 2 {
		t.Fatalf("LoadTasks() returned %d tasks, want 2", len(taskList))
	}

	if taskList[0].Text != "Task with spaces" {
		t.Errorf("task[0].Text = %q, want %q", taskList[0].Text, "Task with spaces")
	}
	if taskList[1].Text != "Task with tabs" {
		t.Errorf("task[1].Text = %q, want %q", taskList[1].Text, "Task with tabs")
	}
}

func TestLoadTasks_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.txt")

	if err := os.WriteFile(tasksPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	fm := NewFileManager(tmpDir)
	taskList, err := fm.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if len(taskList) != 0 {
		t.Errorf("LoadTasks() returned %d tasks, want 0", len(taskList))
	}
}

func TestLoadTasks_CreatesDirectoryWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "subdir", "threedoors")
	fm := NewFileManager(nestedDir)

	_, err := fm.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}

	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}
