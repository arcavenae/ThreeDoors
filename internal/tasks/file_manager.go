package tasks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var defaultTasks = []string{
	"Review project documentation",
	"Update meeting notes",
	"Fix login page bug",
	"Refactor database queries",
	"Write unit tests for API",
}

// FileManager handles reading and creating task files.
type FileManager struct {
	baseDir string
}

// NewFileManager creates a FileManager with the given base directory.
// Use "~/.threedoors" for production, t.TempDir() for tests.
func NewFileManager(baseDir string) *FileManager {
	return &FileManager{baseDir: baseDir}
}

// LoadTasks reads tasks from tasks.txt in the base directory.
// If the directory or file does not exist, they are created with default tasks.
// Blank lines, lines with only whitespace, and lines starting with # are skipped.
func (fm *FileManager) LoadTasks() ([]Task, error) {
	tasksPath := filepath.Join(fm.baseDir, "tasks.txt")

	if err := fm.ensureDirectoryExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure directory exists: %w", err)
	}

	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		if err := fm.createDefaultTasksFile(tasksPath); err != nil {
			return nil, fmt.Errorf("failed to create default tasks file: %w", err)
		}
	}

	return fm.readTasksFromFile(tasksPath)
}

func (fm *FileManager) ensureDirectoryExists() error {
	return os.MkdirAll(fm.baseDir, 0o755)
}

func (fm *FileManager) createDefaultTasksFile(tasksPath string) error {
	content := strings.Join(defaultTasks, "\n") + "\n"

	tmpPath := tasksPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	f, err := os.OpenFile(tmpPath, os.O_RDWR, 0o644)
	if err == nil {
		_ = f.Sync()
		_ = f.Close()
	}

	if err := os.Rename(tmpPath, tasksPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to commit default tasks file: %w", err)
	}

	return nil
}

func (fm *FileManager) readTasksFromFile(tasksPath string) ([]Task, error) {
	file, err := os.Open(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tasks file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var taskList []Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		taskList = append(taskList, Task{Text: line})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	return taskList, nil
}
