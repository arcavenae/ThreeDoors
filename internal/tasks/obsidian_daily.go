package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DailyNotesConfig holds configuration for Obsidian daily note integration.
type DailyNotesConfig struct {
	// Enabled activates daily note reading/writing.
	Enabled bool
	// Folder is the daily notes folder relative to vault root (e.g. "Daily").
	Folder string
	// Heading is the Markdown heading under which tasks are appended (e.g. "## Tasks").
	Heading string
	// DateFormat is a Go time format for the daily note filename (default "2006-01-02.md").
	DateFormat string
}

// defaultDailyNotesDateFormat is the Go time layout for YYYY-MM-DD.md.
const defaultDailyNotesDateFormat = "2006-01-02.md"

// defaultDailyNotesHeading is the default heading under which tasks are appended.
const defaultDailyNotesHeading = "## Tasks"

// sanitizeDailyNotePath validates and sanitizes a date-formatted path.
// It allows subdirectory structures (e.g. "2026/03/15.md") but rejects
// path traversal attempts (..) and null bytes.
func sanitizeDailyNotePath(name string) (string, error) {
	if strings.ContainsAny(name, "\x00") {
		return "", fmt.Errorf("daily note path contains null byte")
	}
	// Clean the path to normalize separators
	cleaned := filepath.Clean(name)
	if cleaned == "." || cleaned == ".." {
		return "", fmt.Errorf("daily note path %q is invalid", name)
	}
	// Reject path traversal: no component should be ".."
	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == ".." {
			return "", fmt.Errorf("daily note path %q contains path traversal", name)
		}
	}
	// Reject absolute paths
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("daily note path %q is absolute", name)
	}
	return cleaned, nil
}

// dailyNotePath computes the absolute path to a daily note for the given date.
func (a *ObsidianAdapter) dailyNotePath(date time.Time) (string, error) {
	if a.dailyNotes == nil || !a.dailyNotes.Enabled {
		return "", fmt.Errorf("daily notes not enabled")
	}

	dateFormat := a.dailyNotes.DateFormat
	if dateFormat == "" {
		dateFormat = defaultDailyNotesDateFormat
	}

	filename := date.Format(dateFormat)
	sanitized, err := sanitizeDailyNotePath(filename)
	if err != nil {
		return "", fmt.Errorf("daily note path: %w", err)
	}

	folder := a.dailyNotes.Folder
	if folder == "" {
		return filepath.Join(a.vaultPath, sanitized), nil
	}

	return filepath.Join(a.vaultPath, folder, sanitized), nil
}

// loadDailyNoteTasks reads tasks from the daily note for the given date.
// Returns an empty slice (not an error) if the daily note file does not exist.
func (a *ObsidianAdapter) loadDailyNoteTasks(date time.Time) ([]*Task, error) {
	path, err := a.dailyNotePath(date)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Task{}, nil
		}
		return nil, fmt.Errorf("read daily note %q: %w", filepath.Base(path), err)
	}

	now := time.Now().UTC()
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

// appendTaskToDailyNote appends a task under the configured heading in today's daily note.
// Creates the file and heading if they don't exist.
func (a *ObsidianAdapter) appendTaskToDailyNote(task *Task, date time.Time) error {
	path, err := a.dailyNotePath(date)
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create daily notes dir: %w", err)
	}

	heading := a.dailyNotes.Heading
	if heading == "" {
		heading = defaultDailyNotesHeading
	}

	taskLine := taskToObsidianLine(task)

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read daily note %q: %w", filepath.Base(path), err)
		}
		// File doesn't exist — create with heading and task
		content := heading + "\n\n" + taskLine + "\n"
		return atomicWriteFile(path, content)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Find the heading line
	headingIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(heading) {
			headingIdx = i
			break
		}
	}

	if headingIdx == -1 {
		// Heading not found — append heading + task at end
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + heading + "\n\n" + taskLine + "\n"
		return atomicWriteFile(path, content)
	}

	// Insert task after the heading section: find the right insertion point.
	// Skip blank lines after heading, then insert after the last checkbox line
	// in this section (or right after the heading if no checkboxes yet).
	insertIdx := headingIdx + 1

	// Skip blank lines after heading
	for insertIdx < len(lines) && strings.TrimSpace(lines[insertIdx]) == "" {
		insertIdx++
	}

	// Find the end of checkbox block under this heading
	for insertIdx < len(lines) {
		trimmed := strings.TrimSpace(lines[insertIdx])
		// Stop at next heading or non-checkbox, non-blank content
		if strings.HasPrefix(trimmed, "#") {
			break
		}
		_, _, _, isCheckbox := parseCheckboxLineObsidian(lines[insertIdx])
		if !isCheckbox && trimmed != "" {
			break
		}
		insertIdx++
	}

	// Insert the task line
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, taskLine)
	newLines = append(newLines, lines[insertIdx:]...)

	return atomicWriteFile(path, strings.Join(newLines, "\n"))
}
