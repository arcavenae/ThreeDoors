package retrospector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseStoryRef(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want string
	}{
		{"standard format", "feat: add widget (Story 51.3)", "51.3"},
		{"uppercase", "feat: add widget (STORY 10.2)", "10.2"},
		{"mixed case", "fix: bug (story 4.5)", "4.5"},
		{"in body text", "This implements Story 22.8 for the dashboard", "22.8"},
		{"no reference", "feat: update README", ""},
		{"number only", "fix: issue #42", ""},
		{"empty string", "", ""},
		{"multiple refs returns first", "Story 1.2 and Story 3.4", "1.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseStoryRef(tt.text)
			if got != tt.want {
				t.Errorf("ParseStoryRef(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestCalculateACMatch(t *testing.T) {
	t.Parallel()

	// Create a temporary stories directory with a test story
	tmpDir := t.TempDir()
	storyContent := `# Story 10.1: Test Story

## Status: In Progress

## Tasks

### Task 1: Implement widget
- Create internal/widget/widget.go
- Update internal/widget/config.go

### Task 2: Add tests
- Create internal/widget/widget_test.go
`
	storyPath := filepath.Join(tmpDir, "10.1.story.md")
	if err := os.WriteFile(storyPath, []byte(storyContent), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tests := []struct {
		name     string
		storyRef string
		prFiles  []string
		want     ACMatch
	}{
		{
			name:     "no story ref",
			storyRef: "",
			prFiles:  []string{"internal/foo/bar.go"},
			want:     ACMatchNoStory,
		},
		{
			name:     "full match",
			storyRef: "10.1",
			prFiles:  []string{"internal/widget/widget.go", "internal/widget/config.go", "internal/widget/widget_test.go"},
			want:     ACMatchFull,
		},
		{
			name:     "partial match",
			storyRef: "10.1",
			prFiles:  []string{"internal/widget/widget.go"},
			want:     ACMatchPartial,
		},
		{
			name:     "no match",
			storyRef: "10.1",
			prFiles:  []string{"internal/other/stuff.go"},
			want:     ACMatchNone,
		},
		{
			name:     "missing story file",
			storyRef: "99.9",
			prFiles:  []string{"internal/foo/bar.go"},
			want:     ACMatchPartial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CalculateACMatch(tmpDir, tt.storyRef, tt.prFiles)
			if err != nil {
				t.Fatalf("CalculateACMatch() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CalculateACMatch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractTaskPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	storyContent := `# Story 5.1: Widget Feature

## Tasks

### Task 1: Core implementation
- Create internal/widget/handler.go
- Update cmd/threedoors/main.go

### Task 2: Tests
- Add internal/widget/handler_test.go

## Quality Gate

AC-Q1
`
	storyPath := filepath.Join(tmpDir, "5.1.story.md")
	if err := os.WriteFile(storyPath, []byte(storyContent), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	paths, err := extractTaskPaths(storyPath)
	if err != nil {
		t.Fatalf("extractTaskPaths() error = %v", err)
	}

	expected := []string{"internal/widget/handler.go", "cmd/threedoors/main.go", "internal/widget/handler_test.go"}
	if len(paths) != len(expected) {
		t.Fatalf("extractTaskPaths() returned %d paths, want %d: %v", len(paths), len(expected), paths)
	}

	for i, want := range expected {
		if paths[i] != want {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want)
		}
	}
}

func TestExtractTaskPaths_NoTasks(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	storyContent := `# Story 1.1: Simple Story

## Tasks

### Task 1: Write documentation
- Update the user guide
- Add examples

## Quality Gate
`
	storyPath := filepath.Join(tmpDir, "1.1.story.md")
	if err := os.WriteFile(storyPath, []byte(storyContent), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	paths, err := extractTaskPaths(storyPath)
	if err != nil {
		t.Fatalf("extractTaskPaths() error = %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("extractTaskPaths() returned %d paths, want 0: %v", len(paths), paths)
	}
}

func TestPathOverlaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		taskPath string
		prFile   string
		want     bool
	}{
		{"directory contains file", "internal/retrospector", "internal/retrospector/log.go", true},
		{"exact match", "internal/retrospector/log.go", "internal/retrospector/log.go", true},
		{"filename match", "log.go", "internal/retrospector/log.go", true},
		{"no overlap", "internal/widget", "internal/retrospector/log.go", false},
		{"partial dir no overlap", "internal/retro", "internal/retrospector/log.go", true},
		{"different filename", "config.go", "internal/retrospector/log.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := pathOverlaps(tt.taskPath, tt.prFile)
			if got != tt.want {
				t.Errorf("pathOverlaps(%q, %q) = %v, want %v", tt.taskPath, tt.prFile, got, tt.want)
			}
		})
	}
}

func TestIsInfraOrDocsPR(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		labels []ghLabel
		want   bool
	}{
		{"infrastructure label", []ghLabel{{Name: "infrastructure"}}, true},
		{"infra label", []ghLabel{{Name: "infra"}}, true},
		{"docs label", []ghLabel{{Name: "documentation"}}, true},
		{"dependencies label", []ghLabel{{Name: "dependencies"}}, true},
		{"chore label", []ghLabel{{Name: "chore"}}, true},
		{"feature label", []ghLabel{{Name: "feature"}}, false},
		{"no labels", nil, false},
		{"mixed labels", []ghLabel{{Name: "feature"}, {Name: "docs"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsInfraOrDocsPR(tt.labels)
			if got != tt.want {
				t.Errorf("IsInfraOrDocsPR(%v) = %v, want %v", tt.labels, got, tt.want)
			}
		})
	}
}
