package sync_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/device"
	gosync "github.com/arcavenae/ThreeDoors/internal/sync"
)

func newTestDeviceID(t *testing.T) device.DeviceID {
	t.Helper()
	id, err := device.NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}
	return id
}

func TestGitSyncTransport_Init_ClonesFromRemote(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)
	repoDir := filepath.Join(t.TempDir(), "sync")

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   newTestDeviceID(t),
		DeviceName: "test-device",
		Executor:   gosync.NewExecGitExecutor(30 * time.Second),
	})

	if err := transport.Init(context.Background()); err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	// Verify the repo was cloned
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		t.Errorf("Init() should create a .git directory, got error: %v", err)
	}

	// Verify README.md was cloned
	content, err := os.ReadFile(filepath.Join(repoDir, "README.md"))
	if err != nil {
		t.Errorf("Init() should clone README.md: %v", err)
	}
	if string(content) != "# sync test repo" {
		t.Errorf("Init() cloned wrong content: %q", content)
	}
}

func TestGitSyncTransport_Init_CreatesNewRepo(t *testing.T) {
	t.Parallel()

	// Create an empty bare repo (no commits)
	bareDir := createBareRepo(t)
	repoDir := filepath.Join(t.TempDir(), "sync")

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   newTestDeviceID(t),
		DeviceName: "test-device",
		Executor:   gosync.NewExecGitExecutor(30 * time.Second),
	})

	if err := transport.Init(context.Background()); err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	// Verify the repo was created
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		t.Errorf("Init() should create a .git directory: %v", err)
	}
}

func TestGitSyncTransport_Push_CommitsAndPushes(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)
	repoDir := filepath.Join(t.TempDir(), "sync")

	executor := gosync.NewExecGitExecutor(30 * time.Second)
	devID := newTestDeviceID(t)

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   devID,
		DeviceName: "test-device",
		Executor:   executor,
	})

	if err := transport.Init(context.Background()); err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	changeset := gosync.Changeset{
		DeviceID:  devID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks.yaml", Content: []byte("- id: task1\n  text: test task\n"), Op: gosync.OpAdd},
		},
	}

	if err := transport.Push(context.Background(), changeset); err != nil {
		t.Fatalf("Push() unexpected error: %v", err)
	}

	// Verify the file was written to the repo
	content, err := os.ReadFile(filepath.Join(repoDir, "tasks.yaml"))
	if err != nil {
		t.Fatalf("Push() should create tasks.yaml: %v", err)
	}
	if string(content) != "- id: task1\n  text: test task\n" {
		t.Errorf("Push() wrote wrong content: %q", content)
	}

	// Verify it was pushed to remote — clone into another dir and check
	verifyDir := cloneRepo(t, bareDir)
	verifyContent, err := os.ReadFile(filepath.Join(verifyDir, "tasks.yaml"))
	if err != nil {
		t.Fatalf("Push() should push to remote, but tasks.yaml not found in fresh clone: %v", err)
	}
	if string(verifyContent) != "- id: task1\n  text: test task\n" {
		t.Errorf("Push() pushed wrong content to remote: %q", verifyContent)
	}
}

func TestGitSyncTransport_Pull_RetrievesRemoteChanges(t *testing.T) {
	t.Parallel()

	bareDir, pushWorkDir := initRepoWithFirstCommit(t)

	// Create the transport's repo by cloning
	repoDir := filepath.Join(t.TempDir(), "sync")

	executor := gosync.NewExecGitExecutor(30 * time.Second)
	devID := newTestDeviceID(t)

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   devID,
		DeviceName: "test-device",
		Executor:   executor,
	})

	if err := transport.Init(context.Background()); err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	// Push a file from another "device"
	commitFile(t, pushWorkDir, "tasks.yaml", "- id: remote-task\n  text: from other device\n")
	pushRepo(t, pushWorkDir)

	// Pull should retrieve the changes
	changeset, err := transport.Pull(context.Background(), time.Time{})
	if err != nil {
		t.Fatalf("Pull() unexpected error: %v", err)
	}

	// Verify the pulled changeset contains the file
	if len(changeset.Files) == 0 {
		t.Fatal("Pull() should return files in changeset")
	}

	found := false
	for _, f := range changeset.Files {
		if f.Path == "tasks.yaml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Pull() changeset should contain tasks.yaml")
	}

	// Verify the file exists in the local repo
	content, err := os.ReadFile(filepath.Join(repoDir, "tasks.yaml"))
	if err != nil {
		t.Fatalf("Pull() should update local repo: %v", err)
	}
	if string(content) != "- id: remote-task\n  text: from other device\n" {
		t.Errorf("Pull() wrote wrong content locally: %q", content)
	}
}

func TestGitSyncTransport_Status_ReportsIdle(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)
	repoDir := filepath.Join(t.TempDir(), "sync")

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   newTestDeviceID(t),
		DeviceName: "test-device",
		Executor:   gosync.NewExecGitExecutor(30 * time.Second),
	})

	if err := transport.Init(context.Background()); err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	status, err := transport.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() unexpected error: %v", err)
	}

	if status.State != "idle" {
		t.Errorf("Status().State = %q, want 'idle'", status.State)
	}
	if status.RemoteURL != bareDir {
		t.Errorf("Status().RemoteURL = %q, want %q", status.RemoteURL, bareDir)
	}
	// Init creates a .gitattributes commit that gets pushed during first push,
	// but after just Init+clone from a repo that already has commits, the
	// .gitattributes commit is unpushed. For a freshly cloned repo with
	// an additional local .gitattributes commit, count may be 0 or 1.
	if status.UnpushedCount > 1 {
		t.Errorf("Status().UnpushedCount = %d, want 0 or 1", status.UnpushedCount)
	}
}

func TestGitSyncTransport_Status_ReportsUnpushed(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)
	repoDir := filepath.Join(t.TempDir(), "sync")

	executor := gosync.NewExecGitExecutor(30 * time.Second)
	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   newTestDeviceID(t),
		DeviceName: "test-device",
		Executor:   executor,
	})

	if err := transport.Init(context.Background()); err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	// Get baseline unpushed count (may include .gitattributes commit from Init)
	baselineStatus, err := transport.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() baseline unexpected error: %v", err)
	}
	baseline := baselineStatus.UnpushedCount

	// Create a local commit without pushing
	commitFile(t, repoDir, "local-change.txt", "local only")

	status, err := transport.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() unexpected error: %v", err)
	}

	if status.UnpushedCount != baseline+1 {
		t.Errorf("Status().UnpushedCount = %d, want %d (baseline %d + 1)", status.UnpushedCount, baseline+1, baseline)
	}
}

func TestGitSyncTransport_Push_DeletedFile(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)
	repoDir := filepath.Join(t.TempDir(), "sync")

	executor := gosync.NewExecGitExecutor(30 * time.Second)
	devID := newTestDeviceID(t)

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   devID,
		DeviceName: "test-device",
		Executor:   executor,
	})

	if err := transport.Init(context.Background()); err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	// First push a file
	changeset := gosync.Changeset{
		DeviceID:  devID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks.yaml", Content: []byte("task data"), Op: gosync.OpAdd},
		},
	}
	if err := transport.Push(context.Background(), changeset); err != nil {
		t.Fatalf("Push(add) unexpected error: %v", err)
	}

	// Then delete it
	delChangeset := gosync.Changeset{
		DeviceID:  devID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks.yaml", Op: gosync.OpDelete},
		},
	}
	if err := transport.Push(context.Background(), delChangeset); err != nil {
		t.Fatalf("Push(delete) unexpected error: %v", err)
	}

	// Verify file no longer exists
	if _, err := os.Stat(filepath.Join(repoDir, "tasks.yaml")); !os.IsNotExist(err) {
		t.Error("Push(delete) should remove the file")
	}
}

func TestGitSyncTransport_NotInitialized(t *testing.T) {
	t.Parallel()

	repoDir := filepath.Join(t.TempDir(), "sync-not-init")

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  "https://example.com/repo.git",
		DeviceID:   newTestDeviceID(t),
		DeviceName: "test-device",
		Executor:   gosync.NewExecGitExecutor(30 * time.Second),
	})

	// Push without init should fail
	err := transport.Push(context.Background(), gosync.Changeset{})
	if err == nil {
		t.Error("Push() without Init() should return error")
	}

	// Pull without init should fail
	_, err = transport.Pull(context.Background(), time.Time{})
	if err == nil {
		t.Error("Pull() without Init() should return error")
	}
}
