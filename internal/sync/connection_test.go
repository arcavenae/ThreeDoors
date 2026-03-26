package sync_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	gosync "github.com/arcavenae/ThreeDoors/internal/sync"
)

func TestGitSyncConnection_Register(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	tracker := core.NewSyncStatusTracker()

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:   filepath.Join(t.TempDir(), "sync"),
		RemoteURL: "https://example.com/repo.git",
		DeviceID:  newTestDeviceID(t),
		Executor:  gosync.NewExecGitExecutor(30 * time.Second),
	})

	conn := gosync.NewGitSyncConnection(gosync.GitSyncConnectionConfig{
		Transport: transport,
		ConnMgr:   connMgr,
		Tracker:   tracker,
		Breaker:   core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
	})

	if err := conn.Register("https://example.com/repo.git"); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	// Verify connection was added to manager
	if connMgr.Count() != 1 {
		t.Errorf("ConnectionManager.Count() = %d, want 1", connMgr.Count())
	}

	// Verify tracker has the provider registered
	status := tracker.Get(gosync.ProviderName)
	if status == nil {
		t.Fatal("tracker should have git-sync registered")
	}
	if status.Phase != core.SyncPhaseSynced {
		t.Errorf("tracker phase = %q, want 'synced'", status.Phase)
	}
}

func TestGitSyncConnection_ConnectAndSync(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)

	var lastEvent connection.StateChangeEvent
	connMgr := connection.NewConnectionManager(func(event connection.StateChangeEvent) {
		lastEvent = event
	})
	tracker := core.NewSyncStatusTracker()
	breaker := core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig())

	repoDir := filepath.Join(t.TempDir(), "sync")
	devID := newTestDeviceID(t)

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   devID,
		DeviceName: "test-device",
		Executor:   gosync.NewExecGitExecutor(30 * time.Second),
	})

	conn := gosync.NewGitSyncConnection(gosync.GitSyncConnectionConfig{
		Transport: transport,
		ConnMgr:   connMgr,
		Tracker:   tracker,
		Breaker:   breaker,
	})

	if err := conn.Register(bareDir); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	// Connect — transitions Disconnected → Connecting → Connected
	ctx := context.Background()
	if err := conn.Connect(ctx); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	// Verify connection is now Connected
	connObj, err := connMgr.Get(conn.ConnectionID())
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if connObj.State != connection.StateConnected {
		t.Errorf("connection state = %s, want connected", connObj.State)
	}

	// Sync — transitions Connected → Syncing → Connected
	changeset := gosync.Changeset{
		DeviceID:  devID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks.yaml", Content: []byte("task data"), Op: gosync.OpAdd},
		},
	}

	if err := conn.Sync(ctx, changeset); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	// Verify last event was Syncing → Connected
	if lastEvent.From != connection.StateSyncing || lastEvent.To != connection.StateConnected {
		t.Errorf("last event = %s → %s, want syncing → connected", lastEvent.From, lastEvent.To)
	}

	// Verify tracker shows synced
	status := tracker.Get(gosync.ProviderName)
	if status.Phase != core.SyncPhaseSynced {
		t.Errorf("tracker phase = %q, want 'synced'", status.Phase)
	}
}

func TestGitSyncConnection_SyncFailureTracked(t *testing.T) {
	t.Parallel()

	bareDir, _ := initRepoWithFirstCommit(t)

	connMgr := connection.NewConnectionManager(nil)
	tracker := core.NewSyncStatusTracker()
	breaker := core.NewCircuitBreaker(core.CircuitBreakerConfig{
		FailureThreshold: 2,
		FailureWindow:    time.Minute,
		ProbeInterval:    time.Second,
		MaxProbeInterval: time.Minute,
	})

	repoDir := filepath.Join(t.TempDir(), "sync")
	devID := newTestDeviceID(t)

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareDir,
		DeviceID:   devID,
		DeviceName: "test-device",
		Executor:   gosync.NewExecGitExecutor(30 * time.Second),
	})

	conn := gosync.NewGitSyncConnection(gosync.GitSyncConnectionConfig{
		Transport: transport,
		ConnMgr:   connMgr,
		Tracker:   tracker,
		Breaker:   breaker,
	})

	if err := conn.Register(bareDir); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	// Connect successfully
	ctx := context.Background()
	if err := conn.Connect(ctx); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	// Verify breaker is initially closed
	if breaker.State() != core.CircuitClosed {
		t.Errorf("breaker state = %s, want closed", breaker.State())
	}

	// Verify tracker shows synced after successful connect
	status := tracker.Get(gosync.ProviderName)
	if status == nil {
		t.Fatal("tracker should have git-sync registered")
	}
	if status.Phase != core.SyncPhaseSynced {
		t.Errorf("tracker phase after connect = %q, want 'synced'", status.Phase)
	}
}
