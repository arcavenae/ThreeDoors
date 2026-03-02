package tasks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	tasksYAMLFile = "tasks.yaml"
	tasksTextFile = "tasks.txt"
	completedFile = "completed.txt"
	configDir     = ".threedoors"
)

var testHomeDir string

// SetHomeDir sets the home directory for testing purposes.
func SetHomeDir(dir string) {
	testHomeDir = dir
}

// TasksFile represents the YAML structure for task persistence.
type TasksFile struct {
	Tasks []*Task `yaml:"tasks"`
}

// GetConfigDirPath returns the path to ~/.threedoors/.
func GetConfigDirPath() (string, error) {
	var homeDir string
	if testHomeDir != "" {
		homeDir = testHomeDir
	} else {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
	}
	return filepath.Join(homeDir, configDir), nil
}

// GetTasksFilePath returns the path to tasks.yaml.
func GetTasksFilePath() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, tasksYAMLFile), nil
}

var defaultTaskTexts = []string{
	"Learn Go",
	"Build a TUI app",
	"Explore Bubbletea",
	"Write tests",
	"Deploy application",
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return configPath, nil
}

// LoadTasks reads tasks from tasks.yaml. If it doesn't exist, checks for
// tasks.txt migration or creates default tasks.
func LoadTasks() ([]*Task, error) {
	configPath, err := EnsureConfigDir()
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
	tasks := make([]*Task, len(defaultTaskTexts))
	for i, text := range defaultTaskTexts {
		tasks[i] = NewTask(text)
	}
	if err := SaveTasks(tasks); err != nil {
		return nil, fmt.Errorf("failed to save default tasks: %w", err)
	}
	return tasks, nil
}

func loadFromYAML(path string) ([]*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var tf TasksFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return tf.Tasks, nil
}

func migrateFromText(path string) ([]*Task, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []*Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			tasks = append(tasks, NewTask(text))
		}
	}
	return tasks, scanner.Err()
}

// SaveTasks persists tasks to tasks.yaml using atomic write.
func SaveTasks(tasks []*Task) error {
	configPath, err := EnsureConfigDir()
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

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	f.Close()

	if err := os.Rename(tmpPath, yamlPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

// AppendCompleted appends a completed task to completed.txt.
func AppendCompleted(task *Task) error {
	configPath, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	completedPath := filepath.Join(configPath, completedFile)
	f, err := os.OpenFile(completedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open completed file: %w", err)
	}
	defer f.Close()

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
