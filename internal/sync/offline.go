package sync

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

const (
	// DefaultGCThreshold is the number of commits before triggering git gc.
	DefaultGCThreshold = 100
)

// OfflineManagerConfig holds configuration for creating an OfflineManager.
type OfflineManagerConfig struct {
	Executor    GitExecutor
	Breaker     *core.CircuitBreaker
	RepoDir     string
	RemoteURL   string
	GCThreshold int
	OnPush      func() // optional callback after successful push
}

// ExtendedSyncStatus holds the full sync status for CLI display (AC7).
type ExtendedSyncStatus struct {
	ConnectivityState string    `json:"connectivity_state"`
	UnpushedCount     int       `json:"unpushed_count"`
	OldestUnpushed    time.Time `json:"oldest_unpushed,omitempty"`
	LocalHEAD         string    `json:"local_head"`
	RemoteHEAD        string    `json:"remote_head"`
}

// OfflineManager manages offline queue and reconciliation.
// It separates local commits (always work) from pushes (require connectivity),
// probes connectivity via circuit breaker, and triggers git gc periodically.
type OfflineManager struct {
	executor       GitExecutor
	breaker        *core.CircuitBreaker
	repoDir        string
	remoteURL      string
	gcThreshold    int
	commitsSinceGC atomic.Int64
	onPush         func()
}

// NewOfflineManager creates an OfflineManager with the given configuration.
func NewOfflineManager(cfg OfflineManagerConfig) *OfflineManager {
	gcThreshold := cfg.GCThreshold
	if gcThreshold <= 0 {
		gcThreshold = DefaultGCThreshold
	}
	return &OfflineManager{
		executor:    cfg.Executor,
		breaker:     cfg.Breaker,
		repoDir:     cfg.RepoDir,
		remoteURL:   cfg.RemoteURL,
		gcThreshold: gcThreshold,
		onPush:      cfg.OnPush,
	}
}

// IsOnline returns true if the circuit breaker indicates connectivity.
func (om *OfflineManager) IsOnline() bool {
	return om.breaker.State() != core.CircuitOpen
}

// ConnectivityState returns a human-readable connectivity state string.
func (om *OfflineManager) ConnectivityState() string {
	switch om.breaker.State() {
	case core.CircuitOpen:
		return "offline"
	case core.CircuitHalfOpen:
		return "probing"
	default:
		return "online"
	}
}

// CommitLocal stages and commits all changes locally without pushing.
// This always works regardless of connectivity — Git's local commits ARE the offline queue.
func (om *OfflineManager) CommitLocal(ctx context.Context, msg string) error {
	if _, err := om.executor.Run(ctx, om.repoDir, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	status, _ := om.executor.Run(ctx, om.repoDir, "status", "--porcelain")
	if strings.TrimSpace(status) == "" {
		return nil // nothing to commit
	}

	if _, err := om.executor.Run(ctx, om.repoDir, "commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	count := om.commitsSinceGC.Add(1)
	if int(count) >= om.gcThreshold {
		om.runGC(ctx)
	}

	return nil
}

// CommitsSinceGC returns the number of commits since the last gc run.
func (om *OfflineManager) CommitsSinceGC() int {
	return int(om.commitsSinceGC.Load())
}

// PushQueued pushes all queued local commits to the remote.
// Returns an error if the circuit breaker is open (offline).
func (om *OfflineManager) PushQueued(ctx context.Context) error {
	err := om.breaker.Execute(func() error {
		_, pushErr := om.executor.Run(ctx, om.repoDir, "push", "-u", "origin", "HEAD")
		return pushErr
	})
	if err != nil {
		return fmt.Errorf("push queued: %w", err)
	}

	if om.onPush != nil {
		om.onPush()
	}
	return nil
}

// ProbeConnectivity checks if the remote is reachable using git ls-remote.
func (om *OfflineManager) ProbeConnectivity(ctx context.Context) bool {
	remote := om.remoteURL
	if remote == "" {
		remote = "origin"
	}
	_, err := om.executor.Run(ctx, om.repoDir, "ls-remote", "--exit-code", remote)
	return err == nil
}

// PullAndReconcile fetches remote changes and rebases local commits on top.
// If rebase conflicts occur, the rebase is aborted and an error is returned
// for manual resolution on the next cycle.
func (om *OfflineManager) PullAndReconcile(ctx context.Context) error {
	if _, err := om.executor.Run(ctx, om.repoDir, "fetch", "origin"); err != nil {
		return fmt.Errorf("git fetch: %w", err)
	}

	// Find remote branch to rebase onto
	remoteBranch := om.findRemoteBranch(ctx)
	if remoteBranch == "" {
		return nil // no remote branch — nothing to reconcile
	}

	if _, err := om.executor.Run(ctx, om.repoDir, "rebase", remoteBranch); err != nil {
		// Rebase conflict — abort and report
		_, _ = om.executor.Run(ctx, om.repoDir, "rebase", "--abort")
		return fmt.Errorf("rebase conflict (aborted): %w", ErrRebaseConflict)
	}

	return nil
}

// UnpushedCount returns the number of local commits not yet pushed to remote.
func (om *OfflineManager) UnpushedCount(ctx context.Context) int {
	output, err := om.executor.Run(ctx, om.repoDir, "rev-list", "--count", "@{upstream}..HEAD")
	if err != nil {
		return 0
	}
	count, parseErr := strconv.Atoi(strings.TrimSpace(output))
	if parseErr != nil {
		return 0
	}
	return count
}

// OldestUnpushedTimestamp returns the author date of the oldest unpushed commit.
func (om *OfflineManager) OldestUnpushedTimestamp(ctx context.Context) time.Time {
	output, err := om.executor.Run(ctx, om.repoDir, "log", "--reverse", "--format=%aI", "@{upstream}..HEAD")
	if err != nil {
		return time.Time{}
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return time.Time{}
	}
	t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(lines[0]))
	if parseErr != nil {
		return time.Time{}
	}
	return t
}

// HeadDivergence returns the local and remote HEAD commit hashes.
func (om *OfflineManager) HeadDivergence(ctx context.Context) (localHEAD, remoteHEAD string) {
	local, err := om.executor.Run(ctx, om.repoDir, "rev-parse", "HEAD")
	if err != nil {
		return "", ""
	}
	remote, err := om.executor.Run(ctx, om.repoDir, "rev-parse", "@{upstream}")
	if err != nil {
		return strings.TrimSpace(local), ""
	}
	return strings.TrimSpace(local), strings.TrimSpace(remote)
}

// ExtendedStatus returns a full status snapshot for CLI display (AC7).
func (om *OfflineManager) ExtendedStatus(ctx context.Context) ExtendedSyncStatus {
	return ExtendedSyncStatus{
		ConnectivityState: om.ConnectivityState(),
		UnpushedCount:     om.UnpushedCount(ctx),
		OldestUnpushed:    om.OldestUnpushedTimestamp(ctx),
		LocalHEAD:         func() string { h, _ := om.HeadDivergence(ctx); return h }(),
		RemoteHEAD:        func() string { _, h := om.HeadDivergence(ctx); return h }(),
	}
}

// runGC runs git gc and resets the commit counter.
func (om *OfflineManager) runGC(ctx context.Context) {
	_, _ = om.executor.Run(ctx, om.repoDir, "gc")
	om.commitsSinceGC.Store(0)
}

// findRemoteBranch returns the remote branch to rebase onto.
func (om *OfflineManager) findRemoteBranch(ctx context.Context) string {
	if _, err := om.executor.Run(ctx, om.repoDir, "rev-parse", "origin/HEAD"); err == nil {
		return "origin/HEAD"
	}
	for _, branch := range []string{"origin/main", "origin/master"} {
		if _, err := om.executor.Run(ctx, om.repoDir, "rev-parse", branch); err == nil {
			return branch
		}
	}
	return ""
}
