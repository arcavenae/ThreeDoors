package textfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"gopkg.in/yaml.v3"
)

const (
	tasksYAMLFile = "tasks.yaml"
	tasksTextFile = "tasks.txt"
	completedFile = "completed.txt"
)

// TasksFile represents the YAML structure for task persistence.
type TasksFile struct {
	Tasks []*core.Task `yaml:"tasks"`
}

var defaultTaskTexts = []string{
	"Learn Go",
	"Build a TUI app",
	"Explore Bubbletea",
	"Write tests",
	"Deploy application",
}

// LoadTasks reads tasks from tasks.yaml. If it doesn't exist, checks for
// tasks.txt migration or creates default tasks.
func LoadTasks() ([]*core.Task, error) {
	configPath, err := core.EnsureConfigDir()
	if err != nil {
		return nil, err
	}

	yamlPath := filepath.Join(configPath, tasksYAMLFile)
	txtPath := filepath.Join(configPath, tasksTextFile)

	// Try YAML first
	if _, err := os.Stat(yamlPath); err == nil {
		return loadFromYAML(yamlPath)
	}

	// Try migrating from text file
	if _, err := os.Stat(txtPath); err == nil {
		tasks, migrateErr := migrateFromText(txtPath)
		if migrateErr != nil {
			return nil, fmt.Errorf("failed to migrate tasks.txt: %w", migrateErr)
		}
		if saveErr := SaveTasks(tasks); saveErr != nil {
			return nil, fmt.Errorf("failed to save migrated tasks: %w", saveErr)
		}
		_ = os.Rename(txtPath, txtPath+".bak")
		return tasks, nil
	}

	// Create defaults
	tasks := make([]*core.Task, len(defaultTaskTexts))
	for i, text := range defaultTaskTexts {
		tasks[i] = core.NewTask(text)
	}
	if err := SaveTasks(tasks); err != nil {
		return nil, fmt.Errorf("failed to save default tasks: %w", err)
	}
	return tasks, nil
}

func loadFromYAML(path string) ([]*core.Task, error) {
	data, err := core.ReadFileWithLimit(path, core.MaxTaskFileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var tf TasksFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return tf.Tasks, nil
}

func migrateFromText(path string) ([]*core.Task, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck // best-effort close on read-only file

	var tasks []*core.Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			tasks = append(tasks, core.NewTask(text))
		}
	}
	return tasks, scanner.Err()
}

// SaveTasks persists tasks to tasks.yaml using atomic write.
func SaveTasks(tasks []*core.Task) error {
	configPath, err := core.EnsureConfigDir()
	if err != nil {
		return err
	}

	yamlPath := filepath.Join(configPath, tasksYAMLFile)
	tmpPath := yamlPath + ".tmp"

	tf := TasksFile{Tasks: tasks}
	data, err := yaml.Marshal(&tf)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, yamlPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

// AppendCompleted appends a completed task to completed.txt.
func AppendCompleted(task *core.Task) error {
	configPath, err := core.EnsureConfigDir()
	if err != nil {
		return err
	}

	completedPath := filepath.Join(configPath, completedFile)
	f, err := os.OpenFile(completedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open completed file: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on append file

	line := fmt.Sprintf("[%s] %s | %s\n",
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		task.ID,
		task.Text,
	)
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to completed file: %w", err)
	}
	return nil
}
