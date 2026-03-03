package llm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CommandRunner abstracts shell command execution for testability.
type CommandRunner interface {
	Run(ctx context.Context, dir string, name string, args ...string) (string, error)
}

// ExecCommandRunner runs commands via os/exec.
type ExecCommandRunner struct{}

func (e *ExecCommandRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("run %s %s: %w: %s", name, strings.Join(args, " "), err, string(out))
	}
	return string(out), nil
}

// GitOutputWriter writes decomposed story specs to a git repository.
type GitOutputWriter struct {
	repoPath     string
	branchPrefix string
	runner       CommandRunner
}

// NewGitOutputWriter creates a GitOutputWriter targeting the given repo path.
func NewGitOutputWriter(repoPath string, branchPrefix string, runner CommandRunner) *GitOutputWriter {
	if branchPrefix == "" {
		branchPrefix = "story/"
	}
	if runner == nil {
		runner = &ExecCommandRunner{}
	}
	return &GitOutputWriter{
		repoPath:     repoPath,
		branchPrefix: branchPrefix,
		runner:       runner,
	}
}

// WriteStories writes story specs as Markdown files to the target repo, creating a branch and commit.
func (w *GitOutputWriter) WriteStories(ctx context.Context, result *DecompositionResult) error {
	if w.repoPath == "" {
		return fmt.Errorf("git writer: repo path not configured")
	}
	if result == nil || len(result.Stories) == 0 {
		return fmt.Errorf("git writer: no stories to write")
	}

	// Verify the repo exists.
	if _, err := os.Stat(filepath.Join(w.repoPath, ".git")); err != nil {
		return fmt.Errorf("git writer: %s is not a git repository: %w", w.repoPath, err)
	}

	branchName := w.branchPrefix + sanitizeBranchName(result.Stories[0].StoryID)

	// Create and checkout branch.
	if _, err := w.runner.Run(ctx, w.repoPath, "git", "checkout", "-b", branchName); err != nil {
		return fmt.Errorf("git writer create branch: %w", err)
	}

	// Write story files.
	storiesDir := filepath.Join(w.repoPath, "docs", "stories")
	if err := os.MkdirAll(storiesDir, 0o755); err != nil {
		return fmt.Errorf("git writer create stories dir: %w", err)
	}

	var filePaths []string
	for _, spec := range result.Stories {
		filename := fmt.Sprintf("%s.story.md", spec.StoryID)
		filePath := filepath.Join(storiesDir, filename)
		content := spec.ToMarkdown()

		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("git writer write %s: %w", filename, err)
		}
		filePaths = append(filePaths, filePath)
	}

	// Stage and commit.
	addArgs := append([]string{"add"}, filePaths...)
	if _, err := w.runner.Run(ctx, w.repoPath, "git", addArgs...); err != nil {
		return fmt.Errorf("git writer stage files: %w", err)
	}

	commitMsg := fmt.Sprintf("feat: add decomposed stories from task: %s", truncate(result.SourceTask, 60))
	if _, err := w.runner.Run(ctx, w.repoPath, "git", "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git writer commit: %w", err)
	}

	return nil
}

func sanitizeBranchName(s string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-", ".", "-")
	return strings.ToLower(replacer.Replace(s))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
