package connection

import (
	"context"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestConnAwareScheduler_SkipsPausedConnections(t *testing.T) {
	t.Parallel()

	manager := NewConnectionManager(nil)
	bridge := NewProviderBridge()

	// Create a connection and transition to Connected → Paused
	conn, err := manager.Add("textfile", "Test", nil)
	if err != nil {
		t.Fatalf("add connection: %v", err)
	}
	if err := manager.Transition(conn.ID, StateConnecting); err != nil {
		t.Fatalf("transition connecting: %v", err)
	}
	if err := manager.Transition(conn.ID, StateConnected); err != nil {
		t.Fatalf("transition connected: %v", err)
	}
	if err := manager.Transition(conn.ID, StatePaused); err != nil {
		t.Fatalf("transition paused: %v", err)
	}

	provider := &stubTaskProvider{
		name:  "textfile",
		tasks: []*core.Task{{ID: "t1", Text: "task"}},
	}
	bridge.Register(conn.ID, provider)

	scheduler := NewConnAwareScheduler(manager, bridge)
	scheduler.AddConnection(conn.ID, provider, core.ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 100 * time.Millisecond,
		Jitter:      0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results := scheduler.Start(ctx)

	// Drain results — should get none since connection is paused
	var count int
	for range results {
		count++
	}

	if count != 0 {
		t.Errorf("got %d results for paused connection, want 0", count)
	}
	if provider.loadCalls != 0 {
		t.Errorf("LoadTasks called %d times for paused connection, want 0", provider.loadCalls)
	}
}

func TestConnAwareScheduler_PollsConnectedConnection(t *testing.T) {
	t.Parallel()

	manager := NewConnectionManager(nil)
	bridge := NewProviderBridge()

	conn, err := manager.Add("textfile", "Active", nil)
	if err != nil {
		t.Fatalf("add connection: %v", err)
	}
	if err := manager.Transition(conn.ID, StateConnecting); err != nil {
		t.Fatalf("transition connecting: %v", err)
	}
	if err := manager.Transition(conn.ID, StateConnected); err != nil {
		t.Fatalf("transition connected: %v", err)
	}

	provider := &stubTaskProvider{
		name:  "textfile",
		tasks: []*core.Task{{ID: "t1", Text: "task"}},
	}
	bridge.Register(conn.ID, provider)

	scheduler := NewConnAwareScheduler(manager, bridge)
	scheduler.AddConnection(conn.ID, provider, core.ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 100 * time.Millisecond,
		Jitter:      0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results := scheduler.Start(ctx)

	var count int
	for r := range results {
		count++
		if r.ConnectionID != conn.ID {
			t.Errorf("got connection ID %q, want %q", r.ConnectionID, conn.ID)
		}
		if r.Err != nil {
			t.Errorf("unexpected error: %v", r.Err)
		}
	}

	if count == 0 {
		t.Error("expected at least one poll result for connected connection")
	}
	if provider.loadCalls == 0 {
		t.Error("expected LoadTasks to be called at least once")
	}
}

func TestConnAwareScheduler_StopIdempotent(t *testing.T) {
	t.Parallel()

	manager := NewConnectionManager(nil)
	bridge := NewProviderBridge()
	scheduler := NewConnAwareScheduler(manager, bridge)

	ctx := context.Background()
	_ = scheduler.Start(ctx)

	// Should not panic
	scheduler.Stop()
	scheduler.Stop()
}

func TestConnAwareScheduler_SkipsDisconnectedConnections(t *testing.T) {
	t.Parallel()

	manager := NewConnectionManager(nil)
	bridge := NewProviderBridge()

	conn, err := manager.Add("textfile", "Disconnected", nil)
	if err != nil {
		t.Fatalf("add connection: %v", err)
	}
	// Connection starts in Disconnected state — don't transition it

	provider := &stubTaskProvider{
		name:  "textfile",
		tasks: []*core.Task{{ID: "t1", Text: "task"}},
	}
	bridge.Register(conn.ID, provider)

	scheduler := NewConnAwareScheduler(manager, bridge)
	scheduler.AddConnection(conn.ID, provider, core.ProviderLoopConfig{
		MinInterval: 50 * time.Millisecond,
		MaxInterval: 100 * time.Millisecond,
		Jitter:      0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results := scheduler.Start(ctx)

	var count int
	for range results {
		count++
	}

	if count != 0 {
		t.Errorf("got %d results for disconnected connection, want 0", count)
	}
}
