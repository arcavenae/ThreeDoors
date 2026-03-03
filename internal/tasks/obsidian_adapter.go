package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ObsidianAdapter reads and writes tasks from Obsidian vault Markdown files.
// Tasks are identified by checkbox syntax: - [ ] (todo), - [x] (complete), - [/] (in-progress).
// Task IDs are embedded as HTML comments (invisible in Obsidian preview): <!-- td:uuid -->
type ObsidianAdapter struct {
	vaultPath   string
	tasksFolder string
	filePattern string
	mu          sync.Mutex
}

// NewObsidianAdapter creates a new ObsidianAdapter for the given vault path.
// The tasksFolder is relative to vaultPath; empty string means the vault root.
// The filePattern is a glob pattern for matching task files; empty string defaults to "*.md".
func NewObsidianAdapter(vaultPath, tasksFolder, filePattern string) *ObsidianAdapter {
	if filePattern == "" {
		filePattern = "*.md"
	}
	return &ObsidianAdapter{
		vaultPath:   vaultPath,
		tasksFolder: tasksFolder,
		filePattern: filePattern,
	}
}

// ValidateVaultPath checks that the vault path exists, is a directory, and is readable/writable.
func ValidateVaultPath(vaultPath string) error {
	info, err := os.Stat(vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("obsidian vault path %q does not exist", vaultPath)
		}
		return fmt.Errorf("obsidian vault path %q: %w", vaultPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("obsidian vault path %q is not a directory", vaultPath)
	}

	// Check readable by listing directory
	f, err := os.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("obsidian vault path %q is not readable: %w", vaultPath, err)
	}
	_ = f.Close()

	// Check writable by attempting to create and remove a temp file
	tmpFile := filepath.Join(vaultPath, ".threedoors-write-test")
	testFile, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("obsidian vault path %q is not writable: %w", vaultPath, err)
	}
	_ = testFile.Close()
	_ = os.Remove(tmpFile)

	return nil
}

// taskDir returns the absolute path to the directory containing task files.
func (a *ObsidianAdapter) taskDir() string {
	if a.tasksFolder == "" {
		return a.vaultPath
	}
	return filepath.Join(a.vaultPath, a.tasksFolder)
}

// idRe matches embedded task IDs: <!-- td:uuid -->
var idRe = regexp.MustCompile(`<!--\s*td:([^\s]+?)\s*-->`)

// dueDateRe matches Obsidian due date emoji format: 📅 YYYY-MM-DD
var dueDateRe = regexp.MustCompile(`📅\s*(\d{4}-\d{2}-\d{2})`)

// priorityRe matches priority indicators: ⏫ (high), 🔼 (medium), 🔽 (low)
var priorityRe = regexp.MustCompile(`[⏫🔼🔽]`)

// tagRe matches #tags in task text
var tagRe = regexp.MustCompile(`#(\w+)`)

// parseCheckboxLineObsidian extracts task text, status, and embedded ID from an Obsidian checkbox line.
// Returns the cleaned text, status, embedded ID (if any), and whether the line is a checkbox line.
func parseCheckboxLineObsidian(line string) (text string, status TaskStatus, embeddedID string, isCheckbox bool) {
	type prefix struct {
		p string
		s TaskStatus
	}
	prefixes := []prefix{
		{"- [x] ", StatusComplete},
		{"- [X] ", StatusComplete},
		{"* [x] ", StatusComplete},
		{"* [X] ", StatusComplete},
		{"- [/] ", StatusInProgress},
		{"* [/] ", StatusInProgress},
		{"- [ ] ", StatusTodo},
		{"* [ ] ", StatusTodo},
	}

	trimmed := strings.TrimLeft(line, " \t")
	for _, p := range prefixes {
		if strings.HasPrefix(trimmed, p.p) {
			raw := strings.TrimSpace(trimmed[len(p.p):])

			// Extract embedded ID if present
			if m := idRe.FindStringSubmatch(raw); len(m) > 1 {
				embeddedID = m[1]
				raw = strings.TrimSpace(idRe.ReplaceAllString(raw, ""))
			}

			return raw, p.s, embeddedID, true
		}
	}
	return line, StatusTodo, "", false
}

// extractMetadata parses Obsidian metadata from task text and returns cleaned text.
func extractMetadata(text string) (cleaned string, tags []string, dueDate string, effort TaskEffort) {
	// Extract due date
	if m := dueDateRe.FindStringSubmatch(text); len(m) > 1 {
		dueDate = m[1]
	}
	cleaned = dueDateRe.ReplaceAllString(text, "")

	// Extract tags
	for _, m := range tagRe.FindAllStringSubmatch(cleaned, -1) {
		if len(m) > 1 {
			tags = append(tags, m[1])
		}
	}

	// Extract priority and map to effort
	priorities := priorityRe.FindAllString(cleaned, -1)
	if len(priorities) > 0 {
		switch priorities[0] {
		case "⏫":
			effort = EffortDeepWork
		case "🔼":
			effort = EffortMedium
		case "🔽":
			effort = EffortQuickWin
		}
	}
	cleaned = priorityRe.ReplaceAllString(cleaned, "")

	// Clean up extra whitespace from removed metadata
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned, tags, dueDate, effort
}

// LoadTasks reads all Markdown files in the configured folder and parses checkbox tasks.
func (a *ObsidianAdapter) LoadTasks() ([]*Task, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.loadTasksLocked()
}

// loadTasksLocked reads tasks without acquiring the mutex (caller must hold it).
func (a *ObsidianAdapter) loadTasksLocked() ([]*Task, error) {
	dir := a.taskDir()

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("obsidian vault folder %q not found: %w", dir, err)
		}
		return nil, fmt.Errorf("obsidian vault folder %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("obsidian vault path %q is not a directory", dir)
	}

	files, err := filepath.Glob(filepath.Join(dir, a.filePattern))
	if err != nil {
		return nil, fmt.Errorf("obsidian glob files with pattern %q: %w", a.filePattern, err)
	}

	var tasks []*Task
	now := time.Now().UTC()

	for _, file := range files {
		fileTasks, err := a.parseFile(file, now)
		if err != nil {
			return nil, fmt.Errorf("obsidian parse %q: %w", filepath.Base(file), err)
		}
		tasks = append(tasks, fileTasks...)
	}

	return tasks, nil
}

// parseFile reads a single Markdown file and extracts checkbox tasks.
func (a *ObsidianAdapter) parseFile(filePath string, now time.Time) ([]*Task, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var tasks []*Task

	for _, line := range lines {
		text, status, embeddedID, isCheckbox := parseCheckboxLineObsidian(line)
		if !isCheckbox {
			continue
		}

		cleaned, tags, _, effort := extractMetadata(text)
		if cleaned == "" {
			continue
		}

		id := embeddedID
		if id == "" {
			id = uuid.New().String()
		}

		task := &Task{
			ID:        id,
			Text:      cleaned,
			Status:    status,
			Effort:    effort,
			Notes:     []TaskNote{},
			CreatedAt: now,
			UpdatedAt: now,
		}

		if len(tags) > 0 {
			task.Context = strings.Join(tags, ", ")
		}

		if status == StatusComplete {
			task.CompletedAt = &now
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// SaveTask updates a single task in its source file via read-modify-write.
// If the task ID doesn't match any existing task, it appends to the first .md file.
func (a *ObsidianAdapter) SaveTask(task *Task) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	dir := a.taskDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("obsidian create dir: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, a.filePattern))
	if err != nil {
		return fmt.Errorf("obsidian glob files: %w", err)
	}

	// Search for the task in existing files by embedded ID
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("obsidian read %q: %w", filepath.Base(file), err)
		}

		lines := strings.Split(string(data), "\n")
		found := false

		for i, line := range lines {
			_, _, embeddedID, isCheckbox := parseCheckboxLineObsidian(line)
			if !isCheckbox {
				continue
			}
			if embeddedID == task.ID {
				lines[i] = taskToObsidianLine(task)
				found = true
				break
			}
		}

		if found {
			content := strings.Join(lines, "\n")
			return atomicWriteFile(file, content)
		}
	}

	// Task not found — append to first file, or create tasks.md
	targetFile := filepath.Join(dir, "tasks.md")
	if len(files) > 0 {
		targetFile = files[0]
	}

	existing := ""
	if data, err := os.ReadFile(targetFile); err == nil {
		existing = string(data)
	}

	newLine := taskToObsidianLine(task)
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	existing += newLine + "\n"

	return atomicWriteFile(targetFile, existing)
}

// SaveTasks replaces all tasks. Writes tasks grouped by their source file.
// New tasks (no matching file) go to tasks.md.
func (a *ObsidianAdapter) SaveTasks(tasks []*Task) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	dir := a.taskDir()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("obsidian create dir: %w", err)
	}

	taskByID := make(map[string]*Task, len(tasks))
	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	files, _ := filepath.Glob(filepath.Join(dir, a.filePattern))

	matched := make(map[string]bool)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("obsidian read %q: %w", filepath.Base(file), err)
		}

		lines := strings.Split(string(data), "\n")
		var newLines []string

		for _, line := range lines {
			_, _, embeddedID, isCheckbox := parseCheckboxLineObsidian(line)
			if !isCheckbox {
				newLines = append(newLines, line)
				continue
			}

			if embeddedID != "" {
				if t, ok := taskByID[embeddedID]; ok {
					newLines = append(newLines, taskToObsidianLine(t))
					matched[embeddedID] = true
					continue
				}
			}
			// Checkbox line not in our set — remove it
		}

		content := strings.Join(newLines, "\n")
		if err := atomicWriteFile(file, content); err != nil {
			return fmt.Errorf("obsidian save %q: %w", filepath.Base(file), err)
		}
	}

	// Write unmatched tasks to tasks.md
	var unmatched []*Task
	for _, t := range tasks {
		if !matched[t.ID] {
			unmatched = append(unmatched, t)
		}
	}

	if len(unmatched) > 0 {
		targetFile := filepath.Join(dir, "tasks.md")
		var lines []string

		// Preserve existing non-checkbox content
		if data, err := os.ReadFile(targetFile); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				_, _, _, isCheckbox := parseCheckboxLineObsidian(line)
				if !isCheckbox {
					lines = append(lines, line)
				}
			}
		}

		for _, t := range unmatched {
			lines = append(lines, taskToObsidianLine(t))
		}

		content := strings.Join(lines, "\n")
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if err := atomicWriteFile(targetFile, content); err != nil {
			return fmt.Errorf("obsidian save tasks.md: %w", err)
		}
	}

	return nil
}

// DeleteTask removes a task line from its source file.
func (a *ObsidianAdapter) DeleteTask(targetID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	dir := a.taskDir()
	files, err := filepath.Glob(filepath.Join(dir, a.filePattern))
	if err != nil {
		return fmt.Errorf("obsidian glob files: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("obsidian read %q: %w", filepath.Base(file), err)
		}

		lines := strings.Split(string(data), "\n")
		var newLines []string
		found := false

		for _, line := range lines {
			_, _, embeddedID, isCheckbox := parseCheckboxLineObsidian(line)
			if isCheckbox && embeddedID == targetID {
				found = true
				continue // skip deleted line
			}
			newLines = append(newLines, line)
		}

		if found {
			content := strings.Join(newLines, "\n")
			return atomicWriteFile(file, content)
		}
	}

	// Non-existent task deletion is idempotent
	return nil
}

// MarkComplete marks a task as complete in its source file.
func (a *ObsidianAdapter) MarkComplete(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	dir := a.taskDir()
	files, err := filepath.Glob(filepath.Join(dir, a.filePattern))
	if err != nil {
		return fmt.Errorf("obsidian glob files: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("obsidian read %q: %w", filepath.Base(file), err)
		}

		lines := strings.Split(string(data), "\n")

		for i, line := range lines {
			text, status, embeddedID, isCheckbox := parseCheckboxLineObsidian(line)
			if !isCheckbox || embeddedID != id {
				continue
			}

			if !IsValidTransition(status, StatusComplete) {
				return fmt.Errorf("obsidian mark complete: invalid transition from %s to complete", status)
			}

			task := &Task{
				ID:     id,
				Text:   text,
				Status: StatusComplete,
			}
			lines[i] = taskToObsidianLine(task)

			content := strings.Join(lines, "\n")
			return atomicWriteFile(file, content)
		}
	}

	return fmt.Errorf("obsidian mark complete: task %q not found", id)
}

// taskToObsidianLine converts a Task back to Obsidian checkbox syntax with embedded ID.
func taskToObsidianLine(task *Task) string {
	var checkbox string
	switch task.Status {
	case StatusComplete:
		checkbox = "- [x] "
	case StatusInProgress:
		checkbox = "- [/] "
	default:
		checkbox = "- [ ] "
	}
	return fmt.Sprintf("%s%s <!-- td:%s -->", checkbox, task.Text, task.ID)
}

// atomicWriteFile writes content to a file using the atomic tmp-sync-rename pattern.
func atomicWriteFile(path, content string) error {
	tmpPath := path + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync temp file: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
