package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"

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

// buildDailyNotePath joins a sanitized relative path with the vault and optional folder.
func (a *ObsidianAdapter) buildDailyNotePath(sanitized string) string {
	if a.dailyNotes.Folder == "" {
		return filepath.Join(a.vaultPath, sanitized)
	}
	return filepath.Join(a.vaultPath, a.dailyNotes.Folder, sanitized)
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

	return a.buildDailyNotePath(sanitized), nil
}

// findHeadingIndex returns the index of the first line matching the heading, or -1.
func findHeadingIndex(lines []string, heading string) int {
	trimmedHeading := strings.TrimSpace(heading)
	for i, line := range lines {
		if strings.TrimSpace(line) == trimmedHeading {
			return i
		}
	}
	return -1
}

// commonDateFormats lists common date formats to try when auto-detecting daily notes.
// The first match wins. These use Go's reference time: Mon Jan 2 15:04:05 MST 2006.
var commonDateFormats = []string{
	"2006-01-02.md",  // YYYY-MM-DD.md (ISO 8601, Obsidian default)
	"01-02-2006.md",  // MM-DD-YYYY.md (US format)
	"02-01-2006.md",  // DD-MM-YYYY.md (European format)
	"2006.01.02.md",  // YYYY.MM.DD.md (dot separator)
	"2006_01_02.md",  // YYYY_MM_DD.md (underscore separator)
	"20060102.md",    // YYYYMMDD.md (compact)
	"2006/01/02.md",  // YYYY/MM/DD.md (subdirectory structure)
	"2006/01-02.md",  // YYYY/MM-DD.md (year subfolder)
	"Jan 2, 2006.md", // Month Day, Year.md (long format)
	"2 Jan 2006.md",  // Day Month Year.md (UK long format)
}

// ValidateDateFormat checks whether a date format string produces a valid filename.
// Returns an error if the format is empty, contains path traversal, or produces
// a filename with null bytes.
func ValidateDateFormat(format string) error {
	if format == "" {
		return fmt.Errorf("date format is empty")
	}
	// Test with a reference date to see if the format produces a sane filename
	testDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	filename := testDate.Format(format)
	_, err := sanitizeDailyNotePath(filename)
	if err != nil {
		return fmt.Errorf("date format %q produces invalid path: %w", format, err)
	}
	if !strings.HasSuffix(filename, ".md") {
		return fmt.Errorf("date format %q does not end with .md", format)
	}
	return nil
}

// loadDailyNoteTasks reads tasks from the daily note for the given date.
// Only tasks under the configured heading section are returned.
// Returns an empty slice (not an error) if the daily note file does not exist.
func (a *ObsidianAdapter) loadDailyNoteTasks(date time.Time) ([]*core.Task, error) {
	path, err := a.resolveDailyNotePath(date)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*core.Task{}, nil
		}
		return nil, fmt.Errorf("read daily note %q: %w", filepath.Base(path), err)
	}

	heading := a.dailyNotes.Heading
	if heading == "" {
		heading = defaultDailyNotesHeading
	}

	now := time.Now().UTC()
	lines := strings.Split(string(data), "\n")
	tasks := a.parseTasksUnderHeading(lines, heading, now)

	return tasks, nil
}

// parseTasksUnderHeading extracts checkbox tasks from lines that fall under
// the specified heading section. If no heading is found, returns an empty slice.
func (a *ObsidianAdapter) parseTasksUnderHeading(lines []string, heading string, now time.Time) []*core.Task {
	headingIdx := findHeadingIndex(lines, heading)
	if headingIdx == -1 {
		// No heading found — no tasks to extract
		return nil
	}

	// Determine heading level to know when the section ends
	headingLevel := headingMarkdownLevel(heading)

	var tasks []*core.Task
	for i := headingIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Stop at next heading of same or higher level
		if strings.HasPrefix(trimmed, "#") {
			level := headingMarkdownLevel(trimmed)
			if level > 0 && level <= headingLevel {
				break
			}
		}

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

		task := &core.Task{
			ID:        id,
			Text:      cleaned,
			Status:    status,
			Effort:    effort,
			Notes:     []core.TaskNote{},
			CreatedAt: now,
			UpdatedAt: now,
		}

		if len(tags) > 0 {
			task.Context = strings.Join(tags, ", ")
		}

		if status == core.StatusComplete {
			task.CompletedAt = &now
		}

		tasks = append(tasks, task)
	}

	return tasks
}

// headingMarkdownLevel returns the heading level (1-6) for a markdown heading,
// or 0 if the line is not a heading.
func headingMarkdownLevel(line string) int {
	trimmed := strings.TrimSpace(line)
	level := 0
	for _, ch := range trimmed {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level > 0 && level <= 6 && len(trimmed) > level && trimmed[level] == ' ' {
		return level
	}
	return 0
}

// resolveDailyNotePath finds the daily note file for the given date.
// It first tries the configured format, then falls back to common date formats.
// Returns the path from dailyNotePath() if no auto-detection is needed.
func (a *ObsidianAdapter) resolveDailyNotePath(date time.Time) (string, error) {
	// Try the configured (or default) format first
	path, err := a.dailyNotePath(date)
	if err != nil {
		return "", err
	}

	if _, statErr := os.Stat(path); statErr == nil {
		return path, nil
	}

	// File not found with configured format — try common formats
	for _, format := range commonDateFormats {
		filename := date.Format(format)
		sanitized, sanitizeErr := sanitizeDailyNotePath(filename)
		if sanitizeErr != nil {
			continue
		}

		candidate := a.buildDailyNotePath(sanitized)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}
	}

	// No file found with any format — return the configured path
	// so callers get the expected "not found" behavior
	return path, nil
}

// QuickAddToDailyNote creates a new task from text and appends it to today's daily note
// under the configured heading. Returns the created task.
func (a *ObsidianAdapter) QuickAddToDailyNote(text string) (*core.Task, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.dailyNotes == nil || !a.dailyNotes.Enabled {
		return nil, fmt.Errorf("daily notes not enabled")
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("task text is empty")
	}

	task := core.NewTask(text)
	if err := a.appendTaskToDailyNote(task, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("quick add to daily note: %w", err)
	}

	return task, nil
}

// appendTaskToDailyNote appends a task under the configured heading in today's daily note.
// Creates the file and heading if they don't exist.
func (a *ObsidianAdapter) appendTaskToDailyNote(task *core.Task, date time.Time) error {
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

	headingIdx := findHeadingIndex(lines, heading)

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
