package tasks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrReadOnly indicates the provider does not support write operations.
var ErrReadOnly = errors.New("apple notes provider is read-only")

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
func (p *AppleNotesProvider) LoadTasks() ([]*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	script := fmt.Sprintf(`tell application "Notes" to get plaintext text of note "%s"`, p.noteTitle)
	output, err := p.executor(ctx, script)
	if err != nil {
		return nil, p.wrapError(err)
	}

	return p.parseNoteBody(output), nil
}

// SaveTask is not supported — Apple Notes is read-only in this story.
func (p *AppleNotesProvider) SaveTask(_ *Task) error {
	return ErrReadOnly
}

// SaveTasks is not supported — Apple Notes is read-only in this story.
func (p *AppleNotesProvider) SaveTasks(_ []*Task) error {
	return ErrReadOnly
}

// DeleteTask is not supported — Apple Notes is read-only in this story.
func (p *AppleNotesProvider) DeleteTask(_ string) error {
	return ErrReadOnly
}

// wrapError maps osascript errors to meaningful wrapped errors.
func (p *AppleNotesProvider) wrapError(err error) error {
	msg := err.Error()

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("apple notes: osascript timed out after 2s: %w", err)
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

// parseNoteBody splits plaintext note content into tasks.
func (p *AppleNotesProvider) parseNoteBody(body string) []*Task {
	if strings.TrimSpace(body) == "" {
		return nil
	}

	lines := strings.Split(body, "\n")
	var tasks []*Task
	lineIndex := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			lineIndex++
			continue
		}

		text, status := parseCheckboxLine(trimmed)
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(p.noteTitle+":"+strconv.Itoa(lineIndex))).String()
		now := time.Now().UTC()

		tasks = append(tasks, &Task{
			ID:        id,
			Text:      text,
			Status:    status,
			Notes:     []TaskNote{},
			CreatedAt: now,
			UpdatedAt: now,
		})

		lineIndex++
	}

	return tasks
}

// parseCheckboxLine extracts task text and status from a line with optional checkbox prefix.
func parseCheckboxLine(line string) (string, TaskStatus) {
	// Try checkbox patterns: "- [ ] text", "- [x] text", "* [ ] text", "* [x] text"
	prefixes := []struct {
		prefix string
		status TaskStatus
	}{
		{"- [x] ", StatusComplete},
		{"- [X] ", StatusComplete},
		{"* [x] ", StatusComplete},
		{"* [X] ", StatusComplete},
		{"- [ ] ", StatusTodo},
		{"* [ ] ", StatusTodo},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(line, p.prefix) {
			return strings.TrimSpace(line[len(p.prefix):]), p.status
		}
	}

	return line, StatusTodo
}
