package sync

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// mockGitExecutor records commands and returns preconfigured results.
type mockGitExecutor struct {
	calls   []gitCall
	results map[string]gitResult
}

type gitCall struct {
	dir  string
	args []string
}

type gitResult struct {
	output string
	err    error
}

func newMockGitExecutor() *mockGitExecutor {
	return &mockGitExecutor{
		results: make(map[string]gitResult),
	}
}

func (m *mockGitExecutor) Run(_ context.Context, dir string, args ...string) (string, error) {
	m.calls = append(m.calls, gitCall{dir: dir, args: args})
	key := strings.Join(args, " ")
	if r, ok := m.results[key]; ok {
		return r.output, r.err
	}
	return "", nil
}

func (m *mockGitExecutor) setResult(args string, output string, err error) {
	m.results[args] = gitResult{output: output, err: err}
}

func (m *mockGitExecutor) callCount(prefix string) int {
	count := 0
	for _, c := range m.calls {
		if len(c.args) > 0 && strings.HasPrefix(strings.Join(c.args, " "), prefix) {
			count++
		}
	}
	return count
}

func TestOfflineManager_IsOnline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cbState    core.CircuitState
		wantOnline bool
	}{
		{"closed circuit is online", core.CircuitClosed, true},
		{"open circuit is offline", core.CircuitOpen, false},
		{"half-open circuit is probing", core.CircuitHalfOpen, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cb := core.NewCircuitBreaker(core.CircuitBreakerConfig{
				FailureThreshold: 5,
				FailureWindow:    time.Minute,
				ProbeInterval:    time.Second,
				MaxProbeInterval: time.Minute,
			})
			if tt.cbState == core.CircuitOpen {
				// Trip the breaker by recording enough failures
				for i := 0; i < 5; i++ {
					_ = cb.Execute(func() error { return errors.New("fail") })
				}
			}
			exec := newMockGitExecutor()
			om := NewOfflineManager(OfflineManagerConfig{
				Executor: exec,
				Breaker:  cb,
				RepoDir:  t.TempDir(),
			})
			got := om.IsOnline()
			if got != tt.wantOnline {
				t.Errorf("IsOnline() = %v, want %v (cb state: %v)", got, tt.wantOnline, cb.State())
			}
		})
	}
}

func TestOfflineManager_ConnectivityState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cbState core.CircuitState
		want    string
	}{
		{"closed is online", core.CircuitClosed, "online"},
		{"open is offline", core.CircuitOpen, "offline"},
		{"half-open is probing", core.CircuitHalfOpen, "probing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cb := core.NewCircuitBreaker(core.CircuitBreakerConfig{
				FailureThreshold: 1,
				FailureWindow:    time.Minute,
				ProbeInterval:    time.Millisecond,
				MaxProbeInterval: time.Minute,
			})
			switch tt.cbState {
			case core.CircuitOpen:
				_ = cb.Execute(func() error { return errors.New("fail") })
			case core.CircuitHalfOpen:
				_ = cb.Execute(func() error { return errors.New("fail") })
				time.Sleep(2 * time.Millisecond) // let probe interval elapse
			}

			exec := newMockGitExecutor()
			om := NewOfflineManager(OfflineManagerConfig{
				Executor: exec,
				Breaker:  cb,
				RepoDir:  t.TempDir(),
			})
			got := om.ConnectivityState()
			if got != tt.want {
				t.Errorf("ConnectivityState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOfflineManager_CommitLocal(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("status --porcelain", "M tasks.yaml\n", nil)

	om := NewOfflineManager(OfflineManagerConfig{
		Executor: exec,
		Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:  "/fake/repo",
	})

	err := om.CommitLocal(context.Background(), "test commit")
	if err != nil {
		t.Fatalf("CommitLocal() error = %v", err)
	}

	// Should have called: add -A, status --porcelain, commit
	if exec.callCount("add -A") != 1 {
		t.Error("expected git add -A call")
	}
	if exec.callCount("commit") != 1 {
		t.Error("expected git commit call")
	}
}

func TestOfflineManager_CommitLocal_NothingToCommit(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("status --porcelain", "", nil)

	om := NewOfflineManager(OfflineManagerConfig{
		Executor: exec,
		Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:  "/fake/repo",
	})

	err := om.CommitLocal(context.Background(), "test commit")
	if err != nil {
		t.Fatalf("CommitLocal() error = %v", err)
	}

	// Should NOT have called commit
	if exec.callCount("commit") != 0 {
		t.Error("should not commit when nothing to stage")
	}
}

func TestOfflineManager_CommitLocal_IncrementsCounter(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("status --porcelain", "M tasks.yaml\n", nil)

	om := NewOfflineManager(OfflineManagerConfig{
		Executor: exec,
		Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:  "/fake/repo",
	})

	for i := 0; i < 3; i++ {
		if err := om.CommitLocal(context.Background(), "commit"); err != nil {
			t.Fatalf("CommitLocal() iteration %d error = %v", i, err)
		}
	}

	if om.CommitsSinceGC() != 3 {
		t.Errorf("CommitsSinceGC() = %d, want 3", om.CommitsSinceGC())
	}
}

func TestOfflineManager_GCTriggersAfter100Commits(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("status --porcelain", "M tasks.yaml\n", nil)

	om := NewOfflineManager(OfflineManagerConfig{
		Executor:    exec,
		Breaker:     core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:     "/fake/repo",
		GCThreshold: 100,
	})

	for i := 0; i < 100; i++ {
		if err := om.CommitLocal(context.Background(), fmt.Sprintf("commit %d", i)); err != nil {
			t.Fatalf("CommitLocal() iteration %d error = %v", i, err)
		}
	}

	// Should have triggered gc
	if exec.callCount("gc") != 1 {
		t.Errorf("expected 1 gc call, got %d", exec.callCount("gc"))
	}

	// Counter should reset after gc
	if om.CommitsSinceGC() != 0 {
		t.Errorf("CommitsSinceGC() = %d, want 0 after gc", om.CommitsSinceGC())
	}
}

func TestOfflineManager_PushQueued(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	om := NewOfflineManager(OfflineManagerConfig{
		Executor: exec,
		Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:  "/fake/repo",
	})

	err := om.PushQueued(context.Background())
	if err != nil {
		t.Fatalf("PushQueued() error = %v", err)
	}

	if exec.callCount("push") != 1 {
		t.Error("expected git push call")
	}
}

func TestOfflineManager_PushQueued_OfflineFails(t *testing.T) {
	t.Parallel()

	cb := core.NewCircuitBreaker(core.CircuitBreakerConfig{
		FailureThreshold: 1,
		FailureWindow:    time.Minute,
		ProbeInterval:    time.Hour,
		MaxProbeInterval: time.Hour,
	})
	// Trip the breaker
	_ = cb.Execute(func() error { return errors.New("fail") })

	exec := newMockGitExecutor()
	om := NewOfflineManager(OfflineManagerConfig{
		Executor: exec,
		Breaker:  cb,
		RepoDir:  "/fake/repo",
	})

	err := om.PushQueued(context.Background())
	if err == nil {
		t.Fatal("PushQueued() should fail when offline")
	}

	if exec.callCount("push") != 0 {
		t.Error("should not attempt push when circuit is open")
	}
}

func TestOfflineManager_ProbeConnectivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		lsRemoteErr error
		wantOnline  bool
	}{
		{"reachable remote", nil, true},
		{"unreachable remote", errors.New("connection refused"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := newMockGitExecutor()
			exec.setResult("ls-remote --exit-code origin", "", tt.lsRemoteErr)

			om := NewOfflineManager(OfflineManagerConfig{
				Executor:  exec,
				Breaker:   core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
				RepoDir:   "/fake/repo",
				RemoteURL: "origin",
			})

			online := om.ProbeConnectivity(context.Background())
			if online != tt.wantOnline {
				t.Errorf("ProbeConnectivity() = %v, want %v", online, tt.wantOnline)
			}
		})
	}
}

func TestOfflineManager_UnpushedCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output string
		err    error
		want   int
	}{
		{"three unpushed", "3\n", nil, 3},
		{"zero unpushed", "0\n", nil, 0},
		{"error returns zero", "", errors.New("no upstream"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := newMockGitExecutor()
			exec.setResult("rev-list --count @{upstream}..HEAD", tt.output, tt.err)

			om := NewOfflineManager(OfflineManagerConfig{
				Executor: exec,
				Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
				RepoDir:  "/fake/repo",
			})

			got := om.UnpushedCount(context.Background())
			if got != tt.want {
				t.Errorf("UnpushedCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestOfflineManager_OldestUnpushedTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		output   string
		err      error
		wantZero bool
	}{
		{"has unpushed", "2026-03-13T10:00:00+00:00\n", nil, false},
		{"no unpushed", "", nil, true},
		{"error", "", errors.New("no upstream"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := newMockGitExecutor()
			exec.setResult("log --reverse --format=%aI @{upstream}..HEAD", tt.output, tt.err)

			om := NewOfflineManager(OfflineManagerConfig{
				Executor: exec,
				Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
				RepoDir:  "/fake/repo",
			})

			got := om.OldestUnpushedTimestamp(context.Background())
			if tt.wantZero && !got.IsZero() {
				t.Errorf("OldestUnpushedTimestamp() = %v, want zero", got)
			}
			if !tt.wantZero && got.IsZero() {
				t.Error("OldestUnpushedTimestamp() = zero, want non-zero")
			}
		})
	}
}

func TestOfflineManager_HeadDivergence(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("rev-parse HEAD", "abc123\n", nil)
	exec.setResult("rev-parse @{upstream}", "def456\n", nil)

	om := NewOfflineManager(OfflineManagerConfig{
		Executor: exec,
		Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:  "/fake/repo",
	})

	local, remote := om.HeadDivergence(context.Background())
	if local != "abc123" {
		t.Errorf("local HEAD = %q, want %q", local, "abc123")
	}
	if remote != "def456" {
		t.Errorf("remote HEAD = %q, want %q", remote, "def456")
	}
}

func TestOfflineManager_PullAndReconcile(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	// pullRebase succeeds
	exec.setResult("fetch origin", "", nil)
	exec.setResult("rev-parse origin/HEAD", "abc123", nil)
	exec.setResult("rebase origin/HEAD", "", nil)

	om := NewOfflineManager(OfflineManagerConfig{
		Executor:  exec,
		Breaker:   core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:   "/fake/repo",
		RemoteURL: "origin",
	})

	err := om.PullAndReconcile(context.Background())
	if err != nil {
		t.Fatalf("PullAndReconcile() error = %v", err)
	}

	if exec.callCount("fetch") != 1 {
		t.Error("expected git fetch call")
	}
}

func TestOfflineManager_PullAndReconcile_RebaseConflict(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("fetch origin", "", nil)
	exec.setResult("rev-parse origin/HEAD", "abc123", nil)
	exec.setResult("rebase origin/HEAD", "", errors.New("conflict"))

	om := NewOfflineManager(OfflineManagerConfig{
		Executor:  exec,
		Breaker:   core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:   "/fake/repo",
		RemoteURL: "origin",
	})

	err := om.PullAndReconcile(context.Background())
	if err == nil {
		t.Fatal("PullAndReconcile() should return error on rebase conflict")
	}

	// Should have called rebase --abort
	if exec.callCount("rebase --abort") != 1 {
		t.Error("expected rebase --abort after conflict")
	}
}

func TestOfflineManager_SyncCycle(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("status --porcelain", "M tasks.yaml\n", nil)
	exec.setResult("fetch origin", "", nil)
	exec.setResult("rev-parse origin/HEAD", "", errors.New("no remote"))

	var pushCalled atomic.Int32

	om := NewOfflineManager(OfflineManagerConfig{
		Executor:  exec,
		Breaker:   core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:   "/fake/repo",
		RemoteURL: "origin",
		OnPush: func() {
			pushCalled.Add(1)
		},
	})

	// Commit locally (always works)
	if err := om.CommitLocal(context.Background(), "sync change"); err != nil {
		t.Fatalf("CommitLocal() error = %v", err)
	}

	// Push (should work since CB is closed)
	if err := om.PushQueued(context.Background()); err != nil {
		t.Fatalf("PushQueued() error = %v", err)
	}

	if exec.callCount("push") != 1 {
		t.Error("expected push call in sync cycle")
	}
}

func TestOfflineManager_ExtendedStatus(t *testing.T) {
	t.Parallel()

	exec := newMockGitExecutor()
	exec.setResult("rev-list --count @{upstream}..HEAD", "5\n", nil)
	exec.setResult("log --reverse --format=%aI @{upstream}..HEAD", "2026-03-13T10:00:00+00:00\n2026-03-13T11:00:00+00:00\n", nil)
	exec.setResult("rev-parse HEAD", "abc123\n", nil)
	exec.setResult("rev-parse @{upstream}", "def456\n", nil)

	om := NewOfflineManager(OfflineManagerConfig{
		Executor: exec,
		Breaker:  core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig()),
		RepoDir:  "/fake/repo",
	})

	status := om.ExtendedStatus(context.Background())

	if status.ConnectivityState != "online" {
		t.Errorf("ConnectivityState = %q, want %q", status.ConnectivityState, "online")
	}
	if status.UnpushedCount != 5 {
		t.Errorf("UnpushedCount = %d, want 5", status.UnpushedCount)
	}
	if status.OldestUnpushed.IsZero() {
		t.Error("OldestUnpushed should not be zero")
	}
	if status.LocalHEAD != "abc123" {
		t.Errorf("LocalHEAD = %q, want %q", status.LocalHEAD, "abc123")
	}
	if status.RemoteHEAD != "def456" {
		t.Errorf("RemoteHEAD = %q, want %q", status.RemoteHEAD, "def456")
	}
}
