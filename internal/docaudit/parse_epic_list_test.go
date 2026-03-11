package docaudit

import (
	"path/filepath"
	"testing"
)

func TestParseEpicList(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "epic-list.md")

	writeFile(t, path, `# Epic List

## Phase 1: Technical Demo & Validation COMPLETE

**Epic 1: Three Doors Technical Demo** COMPLETE
- **Goal:** Build the demo
- **Status:** COMPLETE -- All stories done

**Epic 2: Foundation & Apple Notes Integration** COMPLETE
- **Goal:** Replace text file backend

---

## Phase 2

**Epic 16: iPhone Mobile App (SwiftUI)** ICEBOX
- **Goal:** Mobile version

**Epic 51: SLAES — Self-Learning Agentic Engineering System**
- **Goal:** Continuous improvement
- **Status:** Not Started (0/10 stories)
`)

	result, err := ParseEpicList(path)
	if err != nil {
		t.Fatalf("ParseEpicList() error: %v", err)
	}

	tests := []struct {
		id     string
		status string
		title  string
	}{
		{"1", "Complete", "Three Doors Technical Demo"},
		{"2", "Complete", "Foundation & Apple Notes Integration"},
		{"16", "Icebox", "iPhone Mobile App (SwiftUI)"},
		{"51", "Not Started", "SLAES — Self-Learning Agentic Engineering System"},
	}

	for _, tt := range tests {
		t.Run("Epic_"+tt.id, func(t *testing.T) {
			t.Parallel()
			entry, ok := result[tt.id]
			if !ok {
				t.Fatalf("missing entry for epic %s", tt.id)
			}
			if entry.Status != tt.status {
				t.Errorf("status = %q, want %q", entry.Status, tt.status)
			}
			if entry.Title != tt.title {
				t.Errorf("title = %q, want %q", entry.Title, tt.title)
			}
		})
	}
}

func TestParseEpicList_MissingFile(t *testing.T) {
	t.Parallel()
	_, err := ParseEpicList("/nonexistent/epic-list.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestNormalizeEpicStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"COMPLETE", "Complete"},
		{"COMPLETE -- All stories done", "Complete"},
		{"ICEBOX", "Icebox"},
		{"Not Started (0/10 stories)", "Not Started"},
		{"IN PROGRESS", "In Progress"},
		{"", "Not Started"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := normalizeEpicStatus(tt.input)
			if got != tt.want {
				t.Errorf("normalizeEpicStatus(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
