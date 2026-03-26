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

func TestIntegration_TwoDeviceSync(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)
	executor := gosync.NewExecGitExecutor(30 * time.Second)

	deviceA_ID, _ := device.NewDeviceID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	deviceB_ID, _ := device.NewDeviceID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Device A
	repoDirA := filepath.Join(t.TempDir(), "device-a")
	transportA := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDirA,
		RemoteURL:  bareDir,
		DeviceID:   deviceA_ID,
		DeviceName: "device-a",
		Executor:   executor,
	})

	// Device B
	repoDirB := filepath.Join(t.TempDir(), "device-b")
	transportB := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDirB,
		RemoteURL:  bareDir,
		DeviceID:   deviceB_ID,
		DeviceName: "device-b",
		Executor:   executor,
	})

	ctx := context.Background()

	// Initialize both devices
	if err := transportA.Init(ctx); err != nil {
		t.Fatalf("Device A Init() error: %v", err)
	}
	if err := transportB.Init(ctx); err != nil {
		t.Fatalf("Device B Init() error: %v", err)
	}

	// Device A pushes a task
	changeA := gosync.Changeset{
		DeviceID:  deviceA_ID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks.yaml", Content: []byte("- id: task-from-a\n  text: created on device A\n"), Op: gosync.OpAdd},
		},
	}
	if err := transportA.Push(ctx, changeA); err != nil {
		t.Fatalf("Device A Push() error: %v", err)
	}

	// Device B pulls — should get Device A's task
	csB, err := transportB.Pull(ctx, time.Time{})
	if err != nil {
		t.Fatalf("Device B Pull() error: %v", err)
	}

	// Verify Device B received the file
	found := false
	for _, f := range csB.Files {
		if f.Path == "tasks.yaml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Device B Pull() should include tasks.yaml from Device A")
	}

	contentB, err := os.ReadFile(filepath.Join(repoDirB, "tasks.yaml"))
	if err != nil {
		t.Fatalf("Device B should have tasks.yaml after pull: %v", err)
	}
	if string(contentB) != "- id: task-from-a\n  text: created on device A\n" {
		t.Errorf("Device B tasks.yaml content wrong: %q", contentB)
	}

	// Device B pushes its own change (different file)
	changeB := gosync.Changeset{
		DeviceID:  deviceB_ID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "sessions.jsonl", Content: []byte(`{"session":"b-1"}` + "\n"), Op: gosync.OpAdd},
		},
	}
	if err := transportB.Push(ctx, changeB); err != nil {
		t.Fatalf("Device B Push() error: %v", err)
	}

	// Device A pulls — should get Device B's session
	csA, err := transportA.Pull(ctx, time.Time{})
	if err != nil {
		t.Fatalf("Device A Pull() error: %v", err)
	}

	foundSession := false
	for _, f := range csA.Files {
		if f.Path == "sessions.jsonl" {
			foundSession = true
			break
		}
	}
	if !foundSession {
		t.Error("Device A Pull() should include sessions.jsonl from Device B")
	}
}

func TestIntegration_ConflictResolution(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)
	executor := gosync.NewExecGitExecutor(30 * time.Second)

	deviceA_ID, _ := device.NewDeviceID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	deviceB_ID, _ := device.NewDeviceID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	repoDirA := filepath.Join(t.TempDir(), "device-a")
	transportA := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDirA,
		RemoteURL:  bareDir,
		DeviceID:   deviceA_ID,
		DeviceName: "device-a",
		Executor:   executor,
	})

	repoDirB := filepath.Join(t.TempDir(), "device-b")
	transportB := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDirB,
		RemoteURL:  bareDir,
		DeviceID:   deviceB_ID,
		DeviceName: "device-b",
		Executor:   executor,
	})

	ctx := context.Background()

	if err := transportA.Init(ctx); err != nil {
		t.Fatalf("Device A Init() error: %v", err)
	}
	if err := transportB.Init(ctx); err != nil {
		t.Fatalf("Device B Init() error: %v", err)
	}

	// Both devices modify the same file
	changeA := gosync.Changeset{
		DeviceID:  deviceA_ID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks.yaml", Content: []byte("- id: task1\n  text: version-A\n"), Op: gosync.OpAdd},
		},
	}
	if err := transportA.Push(ctx, changeA); err != nil {
		t.Fatalf("Device A Push() error: %v", err)
	}

	// Device B pushes conflicting change (hasn't pulled A's change)
	changeB := gosync.Changeset{
		DeviceID:  deviceB_ID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks.yaml", Content: []byte("- id: task1\n  text: version-B\n"), Op: gosync.OpAdd},
		},
	}

	// This push should handle the conflict — either succeed with rebase or return conflict error
	err := transportB.Push(ctx, changeB)
	// The push should either succeed (rebase resolved) or fail with a conflict error
	if err != nil {
		// If it's a rebase conflict, the transport should have handled it gracefully
		t.Logf("Device B Push() returned error (expected for conflict): %v", err)
	}

	// Verify that the local repo is in a consistent state (not mid-rebase)
	statusB, err := transportB.Status(ctx)
	if err != nil {
		t.Fatalf("Device B Status() error after conflict: %v", err)
	}

	// Status should not be "error" — conflicts should be handled gracefully
	if statusB.State == "error" {
		t.Errorf("Device B should handle conflicts gracefully, got state: %q", statusB.State)
	}
}

func TestIntegration_GitattributesCreated(t *testing.T) {
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

	// Verify .gitattributes was created with merge strategies
	content, err := os.ReadFile(filepath.Join(repoDir, ".gitattributes"))
	if err != nil {
		t.Fatalf(".gitattributes should be created during Init(): %v", err)
	}

	s := string(content)
	if !contains(s, "tasks.yaml") {
		t.Error(".gitattributes should contain tasks.yaml merge strategy")
	}
	if !contains(s, "sessions.jsonl") {
		t.Error(".gitattributes should contain sessions.jsonl merge strategy")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
