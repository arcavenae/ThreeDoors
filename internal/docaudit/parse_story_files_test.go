package docaudit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseStoryFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create story files with various status formats.
	writeFile(t, filepath.Join(dir, "1.1.story.md"), `# Story 1.1: Project Setup

## Status: Done (PR #2)

## Epic
`)

	writeFile(t, filepath.Join(dir, "2.3.story.md"), `# Story 2.3: Some Feature

**Status:** Not Started

## Epic
`)

	writeFile(t, filepath.Join(dir, "3.5.1.story.md"), `# Story 3.5.1: Bridging Story

Status: In Progress

## Epic
`)

	writeFile(t, filepath.Join(dir, "4.2.story.md"), `# Story 4.2: Pattern Recognition

- **Status:** In Review (PR #42)

## Epic
`)

	// Non-story file should be ignored.
	writeFile(t, filepath.Join(dir, "README.md"), `# Not a story file`)

	result, err := ParseStoryFiles(dir)
	if err != nil {
		t.Fatalf("ParseStoryFiles() error: %v", err)
	}

	tests := []struct {
		id     string
		status string
		title  string
	}{
		{"1.1", "Done", "Project Setup"},
		{"2.3", "Not Started", "Some Feature"},
		{"3.5.1", "In Progress", "Bridging Story"},
		{"4.2", "In Review", "Pattern Recognition"},
	}

	if len(result) != len(tests) {
		t.Fatalf("got %d entries, want %d", len(result), len(tests))
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			entry, ok := result[tt.id]
			if !ok {
				t.Fatalf("missing entry for story %s", tt.id)
			}
			if entry.Status != tt.status {
				t.Errorf("status = %q, want %q", entry.Status, tt.status)
			}
			if entry.Title != tt.title {
				t.Errorf("title = %q, want %q", entry.Title, tt.title)
			}
			if entry.Source != "story_file" {
				t.Errorf("source = %q, want %q", entry.Source, "story_file")
			}
		})
	}
}

func TestParseStoryFiles_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	result, err := ParseStoryFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %d entries, want 0", len(result))
	}
}

func TestParseStoryFiles_MissingDir(t *testing.T) {
	t.Parallel()
	_, err := ParseStoryFiles("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestNormalizeStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"Done (PR #123)", "Done"},
		{"done", "Done"},
		{"Not Started", "Not Started"},
		{"not started", "Not Started"},
		{"In Progress", "In Progress"},
		{"in progress", "In Progress"},
		{"In Review (PR #456)", "In Review"},
		{"Draft", "Draft"},
		{"Icebox", "Icebox"},
		{"  Done  **", "Done"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := normalizeStatus(tt.input)
			if got != tt.want {
				t.Errorf("normalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
