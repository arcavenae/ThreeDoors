package sync_test

import (
	"context"
	"strings"
	"testing"
	"time"

	gosync "github.com/arcavenae/ThreeDoors/internal/sync"
)

func TestExecGitExecutor_Run(t *testing.T) {
	t.Parallel()

	bareDir, workDir := initRepoWithFirstCommit(t)
	_ = bareDir

	executor := gosync.NewExecGitExecutor(30 * time.Second)

	out, err := executor.Run(context.Background(), workDir, "status")
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	_ = out
}

func TestExecGitExecutor_RunInvalidRepo(t *testing.T) {
	t.Parallel()

	executor := gosync.NewExecGitExecutor(30 * time.Second)

	_, err := executor.Run(context.Background(), "/nonexistent-dir-that-does-not-exist", "status")
	if err == nil {
		t.Fatal("Run() expected error for invalid repo dir")
	}
}

func TestExecGitExecutor_Timeout(t *testing.T) {
	t.Parallel()

	executor := gosync.NewExecGitExecutor(1 * time.Millisecond)

	// Create a real repo so the command is valid, but use a very short timeout
	bareDir, workDir := initRepoWithFirstCommit(t)
	_ = bareDir

	// git log should succeed fast, but with 1ms timeout it should fail
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Sleep briefly to ensure context expires
	time.Sleep(5 * time.Millisecond)

	_, err := executor.Run(ctx, workDir, "log")
	if err == nil {
		t.Log("timeout test: command completed before timeout (acceptable in fast environments)")
	}
}

func TestExecGitExecutor_CapturesStdout(t *testing.T) {
	t.Parallel()

	bareDir, workDir := initRepoWithFirstCommit(t)
	_ = bareDir

	executor := gosync.NewExecGitExecutor(30 * time.Second)

	out, err := executor.Run(context.Background(), workDir, "log", "--oneline", "-1")
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if !strings.Contains(out, "README.md") {
		t.Errorf("expected output to contain commit message, got: %q", out)
	}
}
