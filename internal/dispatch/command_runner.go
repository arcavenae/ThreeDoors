package dispatch

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// DefaultTimeout is the default timeout for CLI calls.
const DefaultTimeout = 30 * time.Second

// CommandRunner abstracts subprocess execution for testability.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ExecRunner implements CommandRunner using os/exec.
type ExecRunner struct{}

// Run executes the named command with the given arguments and returns its combined output.
func (r *ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec %s: %s: %w", name, stderr.String(), err)
	}

	return stdout.Bytes(), nil
}
