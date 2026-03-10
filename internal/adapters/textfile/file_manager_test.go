package textfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"

	"gopkg.in/yaml.v3"
)

func TestLoadTasks_NoFileExists(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() failed: %v", err)
	}

	if len(tasks) != len(defaultTaskTexts) {
		t.Errorf("Expected %d default tasks, got %d", len(defaultTaskTexts), len(tasks))
	}

	for i, task := range tasks {
		if task.Text != defaultTaskTexts[i] {
			t.Errorf("Expected task %d text to be %q, got %q", i, defaultTaskTexts[i], task.Text)
		}
		if task.Status != core.StatusTodo {
			t.Errorf("Expected default status %q, got %q", core.StatusTodo, task.Status)
		}
		if task.ID == "" {
			t.Errorf("Expected task %d to have a UUID", i)
		}
	}

	// Verify YAML file was created
	configPath := filepath.Join(tempDir, ".threedoors")
	yamlPath := filepath.Join(configPath, tasksYAMLFile)
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Errorf("tasks.yaml was not created at %s", yamlPath)
	}
}

func TestLoadTasks_YAMLFileExists(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	configPath := filepath.Join(tempDir, ".threedoors")
	_ = os.MkdirAll(configPath, 0o755)

	// Write a YAML tasks file
	task1 := core.NewTask("Task A")
	task2 := core.NewTask("Task B")
	tf := TasksFile{Tasks: []*core.Task{task1, task2}}
	data, _ := yaml.Marshal(&tf)
	_ = os.WriteFile(filepath.Join(configPath, tasksYAMLFile), data, 0o644)

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Text != "Task A" {
		t.Errorf("Expected first task text %q, got %q", "Task A", tasks[0].Text)
	}
}

func TestLoadTasks_MigratesFromText(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	configPath := filepath.Join(tempDir, ".threedoors")
	_ = os.MkdirAll(configPath, 0o755)

	// Write old-style text file
	txtContent := "Task One\nTask Two\nTask Three\n"
	txtPath := filepath.Join(configPath, tasksTextFile)
	_ = os.WriteFile(txtPath, []byte(txtContent), 0o644)

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() failed: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify YAML file exists
	yamlPath := filepath.Join(configPath, tasksYAMLFile)
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Error("tasks.yaml was not created after migration")
	}

	// Verify txt was renamed to .bak
	if _, err := os.Stat(txtPath + ".bak"); os.IsNotExist(err) {
		t.Error("tasks.txt was not renamed to .bak after migration")
	}
}

func TestSaveTasks_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	original := []*core.Task{
		core.NewTask("Alpha task"),
		core.NewTask("Beta task"),
	}
	_ = original[1].UpdateStatus(core.StatusInProgress)
	original[0].AddNote("Test note")

	if err := SaveTasks(original); err != nil {
		t.Fatalf("SaveTasks() failed: %v", err)
	}

	loaded, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() failed: %v", err)
	}

	if len(loaded) != len(original) {
		t.Fatalf("Expected %d tasks, got %d", len(original), len(loaded))
	}

	for i := range original {
		if loaded[i].ID != original[i].ID {
			t.Errorf("Task %d ID mismatch: %q vs %q", i, original[i].ID, loaded[i].ID)
		}
		if loaded[i].Text != original[i].Text {
			t.Errorf("Task %d Text mismatch: %q vs %q", i, original[i].Text, loaded[i].Text)
		}
		if loaded[i].Status != original[i].Status {
			t.Errorf("Task %d Status mismatch: %q vs %q", i, original[i].Status, loaded[i].Status)
		}
	}
}

func TestAppendCompleted(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	task := core.NewTask("Completed task")
	_ = task.UpdateStatus(core.StatusComplete)

	if err := AppendCompleted(task); err != nil {
		t.Fatalf("AppendCompleted() failed: %v", err)
	}

	configPath := filepath.Join(tempDir, ".threedoors")
	completedPath := filepath.Join(configPath, completedFile)
	content, err := os.ReadFile(completedPath)
	if err != nil {
		t.Fatalf("Failed to read completed file: %v", err)
	}

	line := string(content)
	if !strings.Contains(line, task.ID) {
		t.Errorf("Completed file should contain task ID %q, got: %s", task.ID, line)
	}
	if !strings.Contains(line, task.Text) {
		t.Errorf("Completed file should contain task text %q, got: %s", task.Text, line)
	}
	if !strings.HasPrefix(line, "[") {
		t.Errorf("Completed file line should start with timestamp, got: %s", line)
	}
}

func TestSaveTasks_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	tasks := []*core.Task{core.NewTask("Test task")}
	if err := SaveTasks(tasks); err != nil {
		t.Fatalf("SaveTasks() error: %v", err)
	}

	configPath := filepath.Join(tempDir, ".threedoors")
	yamlPath := filepath.Join(configPath, tasksYAMLFile)

	info, err := os.Stat(yamlPath)
	if err != nil {
		t.Fatalf("stat tasks.yaml: %v", err)
	}

	perm := info.Mode().Perm()
	if perm&0o077 != 0 {
		t.Errorf("tasks.yaml permissions = %04o, want no group/other access (got group/other bits: %04o)", perm, perm&0o077)
	}
}

func TestAppendCompleted_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	task := core.NewTask("Perm check task")
	if err := AppendCompleted(task); err != nil {
		t.Fatalf("AppendCompleted() error: %v", err)
	}

	completedPath := filepath.Join(tempDir, ".threedoors", completedFile)
	info, err := os.Stat(completedPath)
	if err != nil {
		t.Fatalf("stat completed.txt: %v", err)
	}

	perm := info.Mode().Perm()
	if perm&0o077 != 0 {
		t.Errorf("completed.txt permissions = %04o, want no group/other access", perm)
	}
}

func TestLoadTasks_EmptyYAML(t *testing.T) {
	tempDir := t.TempDir()
	core.SetHomeDir(tempDir)
	defer core.SetHomeDir("")

	configPath := filepath.Join(tempDir, ".threedoors")
	_ = os.MkdirAll(configPath, 0o755)

	tf := TasksFile{Tasks: []*core.Task{}}
	data, _ := yaml.Marshal(&tf)
	_ = os.WriteFile(filepath.Join(configPath, tasksYAMLFile), data, 0o644)

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks from empty YAML, got %d", len(tasks))
	}
}
