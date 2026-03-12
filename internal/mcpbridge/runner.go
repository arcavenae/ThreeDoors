package mcpbridge

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// CommandRunner executes shell commands and returns their output.
// This interface enables testing with mock command output.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ExecRunner executes real commands via os/exec.
type ExecRunner struct{}

// Run executes a command and returns its combined stdout. If the command
// fails, the error includes stderr output for diagnostic purposes.
func (r *ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command %s failed: %w: %s", name, err, stderr.String())
	}
	return stdout.Bytes(), nil
}
