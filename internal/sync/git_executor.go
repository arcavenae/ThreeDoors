package sync

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitExecutor runs git commands in a repository directory.
type GitExecutor interface {
	// Run executes a git command and returns its stdout output.
	Run(ctx context.Context, repoDir string, args ...string) (string, error)
}

// GitExecError wraps a git command failure with stderr output.
type GitExecError struct {
	Args   []string
	Stderr string
	Err    error
}

func (e *GitExecError) Error() string {
	return fmt.Sprintf("git %s: %v: %s", strings.Join(e.Args, " "), e.Err, strings.TrimSpace(e.Stderr))
}

func (e *GitExecError) Unwrap() error {
	return e.Err
}

// ExecGitExecutor implements GitExecutor using os/exec.
type ExecGitExecutor struct {
	timeout time.Duration
}

// NewExecGitExecutor creates a GitExecutor that calls the git binary via os/exec.
func NewExecGitExecutor(timeout time.Duration) *ExecGitExecutor {
	return &ExecGitExecutor{timeout: timeout}
}

// Run executes a git command in the given repo directory.
func (e *ExecGitExecutor) Run(ctx context.Context, repoDir string, args ...string) (string, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGitNotFound, err)
	}

	if e.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, gitPath, args...)
	cmd.Dir = repoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &GitExecError{
			Args:   args,
			Stderr: stderr.String(),
			Err:    err,
		}
	}

	return strings.TrimRight(stdout.String(), "\n"), nil
}
