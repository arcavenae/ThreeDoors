package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadTasks_NoFileExists(t *testing.T) {
	tempDir := t.TempDir()
	SetHomeDir(tempDir)
	defer SetHomeDir("")

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
		if task.Status != StatusTodo {
			t.Errorf("Expected default status %q, got %q", StatusTodo, task.Status)
		}
		if task.ID == "" {
			t.Errorf("Expected task %d to have a UUID", i)
		}
	}

	// Verify YAML file was created
	configPath := filepath.Join(tempDir, configDir)
	yamlPath := filepath.Join(configPath, tasksYAMLFile)
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Errorf("tasks.yaml was not created at %s", yamlPath)
	}
}

func TestLoadTasks_YAMLFileExists(t *testing.T) {
	tempDir := t.TempDir()
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	configPath := filepath.Join(tempDir, configDir)
	os.MkdirAll(configPath, 0o755)

	// Write a YAML tasks file
	task1 := NewTask("Task A")
	task2 := NewTask("Task B")
	tf := TasksFile{Tasks: []*Task{task1, task2}}
	data, _ := yaml.Marshal(&tf)
	os.WriteFile(filepath.Join(configPath, tasksYAMLFile), data, 0o644)

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
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	configPath := filepath.Join(tempDir, configDir)
	os.MkdirAll(configPath, 0o755)

	// Write old-style text file
	txtContent := "Task One\nTask Two\nTask Three\n"
	txtPath := filepath.Join(configPath, tasksTextFile)
	os.WriteFile(txtPath, []byte(txtContent), 0o644)

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
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	original := []*Task{
		NewTask("Alpha task"),
		NewTask("Beta task"),
	}
	original[1].UpdateStatus(StatusInProgress)
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
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	task := NewTask("Completed task")
	task.UpdateStatus(StatusComplete)

	if err := AppendCompleted(task); err != nil {
		t.Fatalf("AppendCompleted() failed: %v", err)
	}

	configPath := filepath.Join(tempDir, configDir)
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

func TestLoadTasks_EmptyYAML(t *testing.T) {
	tempDir := t.TempDir()
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	configPath := filepath.Join(tempDir, configDir)
	os.MkdirAll(configPath, 0o755)

	tf := TasksFile{Tasks: []*Task{}}
	data, _ := yaml.Marshal(&tf)
	os.WriteFile(filepath.Join(configPath, tasksYAMLFile), data, 0o644)

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks from empty YAML, got %d", len(tasks))
	}
}
