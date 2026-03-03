package applenotes

import (
	"context"
	"errors"
	"fmt"
	"html"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"

	"github.com/google/uuid"
)

// ErrReadOnly references the core sentinel error for read-only providers.
var ErrReadOnly = core.ErrReadOnly

// CommandExecutor abstracts osascript execution for testability.
type CommandExecutor func(ctx context.Context, script string) (string, error)

// defaultExecutor runs osascript for real.
func defaultExecutor(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// AppleNotesProvider reads tasks from Apple Notes via osascript.
type AppleNotesProvider struct {
	noteTitle string
	executor  CommandExecutor
}

// NewAppleNotesProvider creates an AppleNotesProvider with the default osascript executor.
func NewAppleNotesProvider(noteTitle string) *AppleNotesProvider {
	return &AppleNotesProvider{
		noteTitle: noteTitle,
		executor:  defaultExecutor,
	}
}

// NewAppleNotesProviderWithExecutor creates an AppleNotesProvider with a custom executor for testing.
func NewAppleNotesProviderWithExecutor(noteTitle string, executor CommandExecutor) *AppleNotesProvider {
	return &AppleNotesProvider{
		noteTitle: noteTitle,
		executor:  executor,
	}
}

// LoadTasks retrieves tasks from Apple Notes via osascript.
func (p *AppleNotesProvider) LoadTasks() ([]*core.Task, error) {
	raw, err := p.readRawNoteBody()
	if err != nil {
		return nil, err
	}
	return p.parseNoteBody(raw), nil
}

// escapedNoteTitle returns the note title escaped for AppleScript string embedding.
func (p *AppleNotesProvider) escapedNoteTitle() string {
	t := strings.ReplaceAll(p.noteTitle, `\`, `\\`)
	return strings.ReplaceAll(t, `"`, `\"`)
}

// readRawNoteBody reads the plaintext note body via osascript without parsing.
func (p *AppleNotesProvider) readRawNoteBody() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.readRawNoteBodyWithCtx(ctx)
}

// SaveTask writes a single task update back to Apple Notes via read-modify-write.
func (p *AppleNotesProvider) SaveTask(task *core.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read current note body
	raw, err := p.readRawNoteBodyWithCtx(ctx)
	if err != nil {
		return err
	}

	// Find and replace the matching line
	lines := strings.Split(raw, "\n")
	found := false
	lineIndex := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			lineIndex++
			continue
		}
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(p.noteTitle+":"+strconv.Itoa(lineIndex))).String()
		if id == task.ID {
			lines[i] = p.taskToNoteLine(task)
			found = true
			break
		}
		lineIndex++
	}

	if !found {
		lines = append(lines, p.taskToNoteLine(task))
	}

	newBody := strings.Join(lines, "\n")
	return p.writeNoteBodyWithCtx(ctx, newBody)
}

// SaveTasks writes multiple task updates in a single read-modify-write cycle.
func (p *AppleNotesProvider) SaveTasks(tasks []*core.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := p.readRawNoteBodyWithCtx(ctx)
	if err != nil {
		return err
	}

	// Build update map
	updateMap := make(map[string]*core.Task, len(tasks))
	for _, t := range tasks {
		updateMap[t.ID] = t
	}

	lines := strings.Split(raw, "\n")
	matched := make(map[string]bool)
	lineIndex := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			lineIndex++
			continue
		}
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(p.noteTitle+":"+strconv.Itoa(lineIndex))).String()
		if t, ok := updateMap[id]; ok {
			lines[i] = p.taskToNoteLine(t)
			matched[id] = true
		}
		lineIndex++
	}

	// Append any tasks not found in existing lines
	for _, t := range tasks {
		if !matched[t.ID] {
			lines = append(lines, p.taskToNoteLine(t))
		}
	}

	newBody := strings.Join(lines, "\n")
	return p.writeNoteBodyWithCtx(ctx, newBody)
}

// DeleteTask removes a task line from Apple Notes by ID.
func (p *AppleNotesProvider) DeleteTask(taskID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := p.readRawNoteBodyWithCtx(ctx)
	if err != nil {
		return err
	}

	lines := strings.Split(raw, "\n")
	var result []string
	lineIndex := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
			lineIndex++
			continue
		}
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(p.noteTitle+":"+strconv.Itoa(lineIndex))).String()
		if id != taskID {
			result = append(result, line)
		}
		lineIndex++
	}

	newBody := strings.Join(result, "\n")
	return p.writeNoteBodyWithCtx(ctx, newBody)
}

// readRawNoteBodyWithCtx reads the raw note body using an existing context.
func (p *AppleNotesProvider) readRawNoteBodyWithCtx(ctx context.Context) (string, error) {
	script := fmt.Sprintf(`tell application "Notes" to get plaintext text of note "%s"`, p.escapedNoteTitle())
	output, err := p.executor(ctx, script)
	if err != nil {
		return "", p.wrapError(err)
	}
	return output, nil
}

// escapeForAppleScript escapes a string for embedding inside AppleScript double-quoted strings.
func escapeForAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}

// writeNoteBodyWithCtx writes plaintext back to Apple Notes as HTML using an existing context.
func (p *AppleNotesProvider) writeNoteBodyWithCtx(ctx context.Context, body string) error {
	htmlBody := p.plaintextToHTML(body)
	escapedHTML := escapeForAppleScript(htmlBody)
	script := fmt.Sprintf(`tell application "Notes" to set body of note "%s" to "%s"`, p.escapedNoteTitle(), escapedHTML)
	_, err := p.executor(ctx, script)
	if err != nil {
		return p.wrapError(err)
	}
	return nil
}

// taskToNoteLine converts a Task to a checkbox-format note line.
func (p *AppleNotesProvider) taskToNoteLine(task *core.Task) string {
	if task.Status == core.StatusComplete {
		return "- [x] " + task.Text
	}
	return "- [ ] " + task.Text
}

// plaintextToHTML converts plaintext note body to HTML for Apple Notes body property.
func (p *AppleNotesProvider) plaintextToHTML(body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}

	lines := strings.Split(body, "\n")
	var htmlLines []string
	for _, line := range lines {
		if line == "" {
			htmlLines = append(htmlLines, "<div><br></div>")
		} else {
			escaped := html.EscapeString(line)
			htmlLines = append(htmlLines, "<div>"+escaped+"</div>")
		}
	}
	return strings.Join(htmlLines, "\n")
}

// Name returns the provider identifier.
func (p *AppleNotesProvider) Name() string {
	return "applenotes"
}

// Watch returns nil because Apple Notes does not support external change detection.
func (p *AppleNotesProvider) Watch() <-chan core.ChangeEvent {
	return nil
}

// HealthCheck reports the operational status of the Apple Notes provider.
func (p *AppleNotesProvider) HealthCheck() core.HealthCheckResult {
	result := core.HealthCheckResult{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.executor(ctx, `tell application "Notes" to get name of first note`)
	if err != nil {
		result.Items = append(result.Items, core.HealthCheckItem{
			Name:       "Apple Notes Access",
			Status:     core.HealthFail,
			Message:    fmt.Sprintf("Cannot access Apple Notes: %v", err),
			Suggestion: "Ensure Notes.app is running and accessible",
		})
	} else {
		result.Items = append(result.Items, core.HealthCheckItem{
			Name:    "Apple Notes Access",
			Status:  core.HealthOK,
			Message: "Apple Notes accessible",
		})
	}
	return result
}

// MarkComplete is not supported — Apple Notes is read-only in this story.
func (p *AppleNotesProvider) MarkComplete(_ string) error {
	return ErrReadOnly
}

// wrapError maps osascript errors to meaningful wrapped errors.
func (p *AppleNotesProvider) wrapError(err error) error {
	msg := err.Error()

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("apple notes: osascript timed out: %w", err)
	}
	if strings.Contains(msg, "Can't get note") || strings.Contains(msg, "can't get note") {
		return fmt.Errorf("apple notes: note %q not found: %w", p.noteTitle, err)
	}
	if strings.Contains(msg, "not allowed") || strings.Contains(msg, "Not authorized") {
		return fmt.Errorf("apple notes: automation permission denied: %w", err)
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("apple notes: osascript failed: %w", err)
	}

	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("apple notes: osascript not found (not macOS?): %w", err)
	}

	return fmt.Errorf("apple notes: %w", err)
}

// parseNoteBody splits plaintext note content into core.
func (p *AppleNotesProvider) parseNoteBody(body string) []*core.Task {
	if strings.TrimSpace(body) == "" {
		return nil
	}

	lines := strings.Split(body, "\n")
	var tasks []*core.Task
	lineIndex := 0
	now := time.Now().UTC()

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			lineIndex++
			continue
		}

		text, status := parseCheckboxLine(trimmed)
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(p.noteTitle+":"+strconv.Itoa(lineIndex))).String()

		tasks = append(tasks, &core.Task{
			ID:        id,
			Text:      text,
			Status:    status,
			Notes:     []core.TaskNote{},
			CreatedAt: now,
			UpdatedAt: now,
		})

		lineIndex++
	}

	return tasks
}

// parseCheckboxLine extracts task text and status from a line with optional checkbox prefix.
func parseCheckboxLine(line string) (string, core.TaskStatus) {
	// Try checkbox patterns: "- [ ] text", "- [x] text", "* [ ] text", "* [x] text"
	prefixes := []struct {
		prefix string
		status core.TaskStatus
	}{
		{"- [x] ", core.StatusComplete},
		{"- [X] ", core.StatusComplete},
		{"* [x] ", core.StatusComplete},
		{"* [X] ", core.StatusComplete},
		{"- [ ] ", core.StatusTodo},
		{"* [ ] ", core.StatusTodo},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(line, p.prefix) {
			return strings.TrimSpace(line[len(p.prefix):]), p.status
		}
	}

	return line, core.StatusTodo
}
