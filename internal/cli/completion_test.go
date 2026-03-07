package cli

import (
	"bytes"
	"testing"
)

func TestCompletionBash(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"completion", "bash"})

	if err := root.Execute(); err != nil {
		t.Fatalf("completion bash: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("expected bash completion output, got empty")
	}
	if !bytes.Contains(buf.Bytes(), []byte("bash")) {
		t.Error("output does not appear to be bash completion script")
	}
}

func TestCompletionZsh(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"completion", "zsh"})

	if err := root.Execute(); err != nil {
		t.Fatalf("completion zsh: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("expected zsh completion output, got empty")
	}
}

func TestCompletionFish(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"completion", "fish"})

	if err := root.Execute(); err != nil {
		t.Fatalf("completion fish: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("expected fish completion output, got empty")
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"completion", "powershell"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid shell")
	}
}

func TestCompletionNoArgs(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"completion"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no shell specified")
	}
}
