package docaudit

import (
	"path/filepath"
	"testing"
)

func TestParseEpicsAndStories(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "epics-and-stories.md")

	writeFile(t, path, `# ThreeDoors - Epic Breakdown

## Epic 1: Three Doors Technical Demo (P0)

**Status:** COMPLETE

### Stories

#### Story 1.1: Project Setup ✅

Create the Go module and basic project structure.

#### Story 1.2: Display Three Doors

Build the TUI view.

## Epic 51: SLAES — Self-Learning Agentic Engineering System (P1)

**Status:** Not Started (0/10 stories)

### Stories

#### Story 51.1: Retrospector Agent Definition (Responsibility+WHY Format)

Create agents/retrospector.md.

#### Story 51.5: Doc Consistency Audit (Periodic Cross-Check)

Periodically cross-check docs.
`)

	stories, epics, err := ParseEpicsAndStories(path)
	if err != nil {
		t.Fatalf("ParseEpicsAndStories() error: %v", err)
	}

	// Check stories.
	storyTests := []struct {
		id     string
		title  string
		status string
	}{
		{"1.1", "Project Setup ✅", "Done"},
		{"1.2", "Display Three Doors", "Referenced"},
		{"51.1", "Retrospector Agent Definition (Responsibility+WHY Format)", "Referenced"},
		{"51.5", "Doc Consistency Audit (Periodic Cross-Check)", "Referenced"},
	}

	for _, tt := range storyTests {
		t.Run("Story_"+tt.id, func(t *testing.T) {
			t.Parallel()
			entry, ok := stories[tt.id]
			if !ok {
				t.Fatalf("missing story entry for %s", tt.id)
			}
			if entry.Status != tt.status {
				t.Errorf("status = %q, want %q", entry.Status, tt.status)
			}
		})
	}

	// Check epics.
	epicTests := []struct {
		id     string
		status string
	}{
		{"1", "Complete"},
		{"51", "Not Started"},
	}

	for _, tt := range epicTests {
		t.Run("Epic_"+tt.id, func(t *testing.T) {
			t.Parallel()
			entry, ok := epics[tt.id]
			if !ok {
				t.Fatalf("missing epic entry for %s", tt.id)
			}
			if entry.Status != tt.status {
				t.Errorf("status = %q, want %q", entry.Status, tt.status)
			}
		})
	}
}

func TestParseEpicsAndStories_MissingFile(t *testing.T) {
	t.Parallel()
	_, _, err := ParseEpicsAndStories("/nonexistent/epics-and-stories.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestStripPriority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"Three Doors Technical Demo (P0)", "Three Doors Technical Demo"},
		{"SLAES (P1)", "SLAES"},
		{"No Priority", "No Priority"},
		{"Something (P2)", "Something"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := stripPriority(tt.input)
			if got != tt.want {
				t.Errorf("stripPriority(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
