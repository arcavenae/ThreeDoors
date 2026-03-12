package llm

import (
	"context"
	"testing"
)

func TestExecRunnerRunWithStdin(t *testing.T) {
	t.Parallel()

	runner := &ExecRunner{}

	// echo reads from stdin and we use printf to test basic execution
	stdout, stderr, err := runner.RunWithStdin(context.Background(), "", "echo", "hello")
	if err != nil {
		t.Fatalf("RunWithStdin() error = %v, stderr = %q", err, stderr)
	}
	if stdout != "hello\n" {
		t.Errorf("stdout = %q, want %q", stdout, "hello\n")
	}
}

func TestExecRunnerRunWithStdinPipes(t *testing.T) {
	t.Parallel()

	runner := &ExecRunner{}

	// cat reads from stdin and echoes to stdout
	stdout, stderr, err := runner.RunWithStdin(context.Background(), "piped input", "cat", "-")
	if err != nil {
		t.Fatalf("RunWithStdin() error = %v, stderr = %q", err, stderr)
	}
	if stdout != "piped input" {
		t.Errorf("stdout = %q, want %q", stdout, "piped input")
	}
}

func TestExecRunnerNonZeroExit(t *testing.T) {
	t.Parallel()

	runner := &ExecRunner{}

	_, _, err := runner.RunWithStdin(context.Background(), "", "false")
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
}

func TestExecRunnerCommandNotFound(t *testing.T) {
	t.Parallel()

	runner := &ExecRunner{}

	_, _, err := runner.RunWithStdin(context.Background(), "", "nonexistent-command-12345")
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}
