package tasks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImportResult holds the outcome of a task import operation.
type ImportResult struct {
	Tasks      []*Task
	SourcePath string
	Format     string // "text" or "markdown"
}

// ImportTasksFromFile reads tasks from a file at the given path.
// Supports plain text (one task per line) and Markdown (checkbox syntax).
func ImportTasksFromFile(path string) (*ImportResult, error) {
	absPath, err := expandPath(path)
	if err != nil {
		return nil, fmt.Errorf("expand path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", absPath, err)
	}

	format := detectFormat(absPath, string(data))
	var imported []*Task

	switch format {
	case "markdown":
		imported = parseMarkdownTasks(string(data))
	default:
		imported = parseTextTasks(string(data))
	}

	return &ImportResult{
		Tasks:      imported,
		SourcePath: absPath,
		Format:     format,
	}, nil
}

// detectFormat determines the file format based on extension and content.
func detectFormat(path string, content string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".md" || ext == ".markdown" {
		return "markdown"
	}

	// Check content for markdown checkbox patterns
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- [ ] ") || strings.HasPrefix(line, "- [x] ") ||
			strings.HasPrefix(line, "- [X] ") || strings.HasPrefix(line, "* [ ] ") ||
			strings.HasPrefix(line, "* [x] ") || strings.HasPrefix(line, "* [X] ") {
			return "markdown"
		}
	}

	return "text"
}

// parseTextTasks parses plain text with one task per line.
func parseTextTasks(content string) []*Task {
	var result []*Task
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		// Strip leading list markers (-, *, numbers)
		text = stripListMarker(text)
		if text == "" {
			continue
		}
		if len(text) > 500 {
			text = text[:500]
		}
		result = append(result, NewTask(text))
	}
	return result
}

// parseMarkdownTasks parses Markdown checkbox items.
func parseMarkdownTasks(content string) []*Task {
	var result []*Task
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		text, completed := parseCheckbox(line)
		if text == "" {
			continue
		}
		if len(text) > 500 {
			text = text[:500]
		}
		task := NewTask(text)
		if completed {
			now := task.CreatedAt
			task.Status = StatusComplete
			task.CompletedAt = &now
		}
		result = append(result, task)
	}
	return result
}

// parseCheckbox extracts task text from a Markdown checkbox line.
// Returns the text and whether the checkbox was checked.
// Returns empty string if the line is not a checkbox.
func parseCheckbox(line string) (string, bool) {
	prefixes := []struct {
		prefix    string
		completed bool
	}{
		{"- [x] ", true},
		{"- [X] ", true},
		{"* [x] ", true},
		{"* [X] ", true},
		{"- [ ] ", false},
		{"* [ ] ", false},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(line, p.prefix) {
			text := strings.TrimSpace(line[len(p.prefix):])
			return text, p.completed
		}
	}
	return "", false
}

// stripListMarker removes common list markers from a line.
func stripListMarker(line string) string {
	// Remove "- ", "* ", "1. ", "2) " etc.
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:])
	}

	// Numbered lists: "1. " or "1) "
	for i, c := range line {
		if c >= '0' && c <= '9' {
			continue
		}
		if (c == '.' || c == ')') && i > 0 && i < len(line)-1 && line[i+1] == ' ' {
			return strings.TrimSpace(line[i+2:])
		}
		break
	}

	return line
}

// expandPath expands ~ to home directory and cleans the path.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}
	return filepath.Clean(path), nil
}
