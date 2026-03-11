package docaudit

import (
	"path/filepath"
	"testing"
)

func TestParseRoadmap(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "ROADMAP.md")

	writeFile(t, path, `# ROADMAP — ThreeDoors

### Epic 41: Charm Ecosystem Adoption & TUI Polish (P1) — 6/6 stories done

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 41.1 | Lipgloss Table Migration | Done (PR #350) | P1 | None |
| 41.2 | Bubble Tea Patterns | Done (PR #355) | P1 | 41.1 |

### Epic 51: SLAES — Self-Learning Agentic Engineering System (P1) — 0/10 stories done

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 51.1 | Retrospector Agent Definition | Not Started | P1 | None |
| 51.5 | Doc Consistency Audit | Not Started | P1 | 51.1 |
`)

	stories, epics, err := ParseRoadmap(path)
	if err != nil {
		t.Fatalf("ParseRoadmap() error: %v", err)
	}

	// Check stories.
	storyTests := []struct {
		id     string
		status string
		title  string
	}{
		{"41.1", "Done", "Lipgloss Table Migration"},
		{"41.2", "Done", "Bubble Tea Patterns"},
		{"51.1", "Not Started", "Retrospector Agent Definition"},
		{"51.5", "Not Started", "Doc Consistency Audit"},
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
			if entry.Title != tt.title {
				t.Errorf("title = %q, want %q", entry.Title, tt.title)
			}
		})
	}

	// Check epics.
	epicTests := []struct {
		id     string
		status string
		done   int
		total  int
	}{
		{"41", "Complete", 6, 6},
		{"51", "Not Started", 0, 10},
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
			if entry.StoriesDone != tt.done {
				t.Errorf("stories done = %d, want %d", entry.StoriesDone, tt.done)
			}
			if entry.StoryCount != tt.total {
				t.Errorf("story count = %d, want %d", entry.StoryCount, tt.total)
			}
		})
	}
}

func TestParseRoadmap_MissingFile(t *testing.T) {
	t.Parallel()
	_, _, err := ParseRoadmap("/nonexistent/ROADMAP.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
