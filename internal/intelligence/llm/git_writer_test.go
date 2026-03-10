package llm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockRunner records commands for verification.
type mockRunner struct {
	commands []string
	err      error
}

func (m *mockRunner) Run(_ context.Context, dir string, name string, args ...string) (string, error) {
	cmd := name + " " + strings.Join(args, " ")
	m.commands = append(m.commands, cmd)
	if m.err != nil {
		return "", m.err
	}
	return "", nil
}

func TestGitOutputWriter_WriteStories(t *testing.T) {
	t.Parallel()

	t.Run("writes stories and creates git commit", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		// Create a fake .git directory to pass the repo check.
		if err := os.MkdirAll(filepath.Join(tmpDir, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}

		runner := &mockRunner{}
		writer := NewGitOutputWriter(tmpDir, "story/", runner)

		result := &DecompositionResult{
			SourceTask: "Build authentication",
			Stories: []StorySpec{
				{
					StoryID:   "14.1",
					Title:     "Auth Spike",
					UserStory: "As a dev, I want auth",
					ACs:       []string{"AC1"},
					Tasks:     []string{"T1"},
				},
			},
			Backend:     "test",
			GeneratedAt: time.Now().UTC(),
		}

		err := writer.WriteStories(context.Background(), result)
		if err != nil {
			t.Fatalf("WriteStories() error = %v", err)
		}

		// Verify git commands were executed.
		if len(runner.commands) != 3 {
			t.Fatalf("expected 3 git commands, got %d: %v", len(runner.commands), runner.commands)
		}

		if !strings.Contains(runner.commands[0], "checkout -b story/14-1") {
			t.Errorf("expected checkout command, got %q", runner.commands[0])
		}
		if !strings.Contains(runner.commands[1], "add") {
			t.Errorf("expected add command, got %q", runner.commands[1])
		}
		if !strings.Contains(runner.commands[2], "commit") {
			t.Errorf("expected commit command, got %q", runner.commands[2])
		}

		// Verify the story file was written.
		storyPath := filepath.Join(tmpDir, "docs", "stories", "14.1.story.md")
		content, err := os.ReadFile(storyPath)
		if err != nil {
			t.Fatalf("story file not written: %v", err)
			return
		}
		if !strings.Contains(string(content), "# Story 14.1: Auth Spike") {
			t.Error("story file missing expected content")
		}
	})

	t.Run("errors with empty repo path", func(t *testing.T) {
		t.Parallel()
		writer := NewGitOutputWriter("", "story/", &mockRunner{})
		err := writer.WriteStories(context.Background(), &DecompositionResult{
			Stories: []StorySpec{{StoryID: "1.1"}},
		})
		if err == nil {
			t.Error("expected error for empty repo path")
		}
	})

	t.Run("errors with nil result", func(t *testing.T) {
		t.Parallel()
		writer := NewGitOutputWriter("/tmp/repo", "story/", &mockRunner{})
		err := writer.WriteStories(context.Background(), nil)
		if err == nil {
			t.Error("expected error for nil result")
		}
	})

	t.Run("errors with no stories", func(t *testing.T) {
		t.Parallel()
		writer := NewGitOutputWriter("/tmp/repo", "story/", &mockRunner{})
		err := writer.WriteStories(context.Background(), &DecompositionResult{})
		if err == nil {
			t.Error("expected error for empty stories")
		}
	})

	t.Run("errors when not a git repo", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		writer := NewGitOutputWriter(tmpDir, "story/", &mockRunner{})
		err := writer.WriteStories(context.Background(), &DecompositionResult{
			Stories: []StorySpec{{StoryID: "1.1"}},
		})
		if err == nil {
			t.Error("expected error for non-git directory")
		}
	})
}

func TestSanitizeBranchName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"14.1", "14-1"},
		{"Feature Name", "feature-name"},
		{"a/b/c", "a-b-c"},
		{"UPPER", "upper"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := sanitizeBranchName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeBranchName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
