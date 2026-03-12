package llm

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// CLIRunner abstracts subprocess execution with stdin support for testability.
type CLIRunner interface {
	RunWithStdin(ctx context.Context, stdin string, name string, args ...string) (stdout string, stderr string, err error)
}

// ExecRunner implements CLIRunner using os/exec.
type ExecRunner struct{}

// RunWithStdin executes a command, optionally piping stdin, and returns stdout/stderr separately.
func (r *ExecRunner) RunWithStdin(ctx context.Context, stdin string, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	err := cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}
