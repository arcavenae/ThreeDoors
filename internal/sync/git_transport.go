package sync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arcaven/ThreeDoors/internal/device"
)

// GitSyncTransport implements SyncTransport using a local Git repository
// that syncs with a shared bare remote.
type GitSyncTransport struct {
	mu          sync.Mutex
	repoDir     string
	remoteURL   string
	deviceID    device.DeviceID
	deviceName  string
	executor    GitExecutor
	now         func() time.Time
	initialized bool
}

// GitSyncTransportConfig holds configuration for creating a GitSyncTransport.
type GitSyncTransportConfig struct {
	RepoDir    string
	RemoteURL  string
	DeviceID   device.DeviceID
	DeviceName string
	Executor   GitExecutor
	Now        func() time.Time
}

// NewGitSyncTransport creates a new Git-based sync transport.
func NewGitSyncTransport(cfg GitSyncTransportConfig) *GitSyncTransport {
	now := cfg.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &GitSyncTransport{
		repoDir:    cfg.RepoDir,
		remoteURL:  cfg.RemoteURL,
		deviceID:   cfg.DeviceID,
		deviceName: cfg.DeviceName,
		executor:   cfg.Executor,
		now:        now,
	}
}

// Init initializes the sync repository — clones from remote or creates a new repo.
func (g *GitSyncTransport) Init(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Try to clone from the remote
	_, err := g.executor.Run(ctx, ".", "clone", g.remoteURL, g.repoDir)
	if err != nil {
		// Clone failed — might be empty remote. Initialize locally.
		if mkErr := os.MkdirAll(g.repoDir, 0o700); mkErr != nil {
			return fmt.Errorf("create sync dir: %w", mkErr)
		}
		if _, initErr := g.executor.Run(ctx, g.repoDir, "init"); initErr != nil {
			return fmt.Errorf("git init: %w", initErr)
		}
		if _, remoteErr := g.executor.Run(ctx, g.repoDir, "remote", "add", "origin", g.remoteURL); remoteErr != nil {
			return fmt.Errorf("git remote add: %w", remoteErr)
		}
	}

	// Configure git user for commits
	_, _ = g.executor.Run(ctx, g.repoDir, "config", "user.name", "ThreeDoors Sync")
	_, _ = g.executor.Run(ctx, g.repoDir, "config", "user.email", "sync@threedoors.local")

	// Create .gitattributes with merge strategies if it doesn't exist
	gitattrsPath := filepath.Join(g.repoDir, ".gitattributes")
	if _, statErr := os.Stat(gitattrsPath); os.IsNotExist(statErr) {
		if err := g.writeGitattributes(); err != nil {
			return fmt.Errorf("write .gitattributes: %w", err)
		}
		// Commit .gitattributes so it doesn't conflict with pulls
		_, _ = g.executor.Run(ctx, g.repoDir, "add", ".gitattributes")
		_, _ = g.executor.Run(ctx, g.repoDir, "commit", "-m", "chore: add .gitattributes merge strategies")
	}

	g.initialized = true
	return nil
}

// Push stages, commits, and pushes local changes to the remote.
func (g *GitSyncTransport) Push(ctx context.Context, changeset Changeset) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.initialized {
		return ErrNotInitialized
	}

	// Apply file changes
	for _, f := range changeset.Files {
		path := filepath.Join(g.repoDir, f.Path)
		switch f.Op {
		case OpAdd, OpModify:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("mkdir for %s: %w", f.Path, err)
			}
			if err := os.WriteFile(path, f.Content, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", f.Path, err)
			}
		case OpDelete:
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("delete %s: %w", f.Path, err)
			}
		}
	}

	// Stage all changes
	if _, err := g.executor.Run(ctx, g.repoDir, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are changes to commit
	status, _ := g.executor.Run(ctx, g.repoDir, "status", "--porcelain")
	if strings.TrimSpace(status) == "" {
		return nil // nothing to commit
	}

	// Commit with device info
	shortID := g.deviceID.String()
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	msg := fmt.Sprintf("sync: %s (%s)", g.deviceName, shortID)
	if _, err := g.executor.Run(ctx, g.repoDir, "commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	// Pull with rebase before push (to integrate remote changes)
	if err := g.pullRebase(ctx); err != nil {
		// If rebase fails, abort and keep local version
		_, _ = g.executor.Run(ctx, g.repoDir, "rebase", "--abort")
		// Log the conflict but don't fail — local version wins
	}

	// Push to remote
	if _, err := g.executor.Run(ctx, g.repoDir, "push", "-u", "origin", "HEAD"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	return nil
}

// Pull fetches and rebases remote changes, returning the changeset.
func (g *GitSyncTransport) Pull(ctx context.Context, since time.Time) (Changeset, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.initialized {
		return Changeset{}, ErrNotInitialized
	}

	// Get the current HEAD before pull
	beforeHead, _ := g.executor.Run(ctx, g.repoDir, "rev-parse", "HEAD")

	// Fetch and rebase
	if err := g.pullRebase(ctx); err != nil {
		return Changeset{}, fmt.Errorf("pull: %w", err)
	}

	// Get HEAD after pull
	afterHead, _ := g.executor.Run(ctx, g.repoDir, "rev-parse", "HEAD")

	// Build changeset from diff
	cs := Changeset{
		DeviceID:  g.deviceID,
		Timestamp: g.now(),
	}

	if beforeHead != afterHead {
		// Get list of changed files
		diffOut, err := g.executor.Run(ctx, g.repoDir, "diff", "--name-status", beforeHead, afterHead)
		if err != nil {
			// If diff fails (e.g. beforeHead was empty), list all tracked files
			diffOut, _ = g.executor.Run(ctx, g.repoDir, "ls-files")
			for _, line := range strings.Split(diffOut, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				content, _ := os.ReadFile(filepath.Join(g.repoDir, line))
				cs.Files = append(cs.Files, SyncFile{
					Path:    line,
					Content: content,
					Op:      OpModify,
				})
			}
		} else {
			for _, line := range strings.Split(diffOut, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, "\t", 2)
				if len(parts) != 2 {
					continue
				}
				status := parts[0]
				name := parts[1]

				var op SyncOp
				switch status {
				case "A":
					op = OpAdd
				case "D":
					op = OpDelete
				default:
					op = OpModify
				}

				var content []byte
				if op != OpDelete {
					content, _ = os.ReadFile(filepath.Join(g.repoDir, name))
				}

				cs.Files = append(cs.Files, SyncFile{
					Path:    name,
					Content: content,
					Op:      op,
				})
			}
		}
	}

	return cs, nil
}

// pullRebase fetches from origin and rebases the current branch.
func (g *GitSyncTransport) pullRebase(ctx context.Context) error {
	// Fetch first
	_, fetchErr := g.executor.Run(ctx, g.repoDir, "fetch", "origin")
	if fetchErr != nil {
		return fmt.Errorf("git fetch: %w", fetchErr)
	}

	// Check if there's a remote branch to rebase onto
	_, err := g.executor.Run(ctx, g.repoDir, "rev-parse", "origin/HEAD")
	if err != nil {
		// Try origin/main or origin/master
		for _, branch := range []string{"origin/main", "origin/master"} {
			if _, branchErr := g.executor.Run(ctx, g.repoDir, "rev-parse", branch); branchErr == nil {
				_, rebaseErr := g.executor.Run(ctx, g.repoDir, "rebase", branch)
				return rebaseErr
			}
		}
		// No remote branch to rebase onto — that's fine for fresh repos
		return nil
	}

	_, rebaseErr := g.executor.Run(ctx, g.repoDir, "rebase", "origin/HEAD")
	return rebaseErr
}

// Status returns the current sync status.
func (g *GitSyncTransport) Status(ctx context.Context) (SyncStatus, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.initialized {
		return SyncStatus{State: "not_initialized"}, nil
	}

	status := SyncStatus{
		State:     "idle",
		RemoteURL: g.remoteURL,
	}

	// Count unpushed commits
	unpushed, err := g.executor.Run(ctx, g.repoDir, "rev-list", "--count", "@{upstream}..HEAD")
	if err == nil {
		count, parseErr := strconv.Atoi(strings.TrimSpace(unpushed))
		if parseErr == nil {
			status.UnpushedCount = count
		}
	}

	// Get last sync time from the latest commit
	logOut, err := g.executor.Run(ctx, g.repoDir, "log", "-1", "--format=%aI")
	if err == nil && logOut != "" {
		if t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(logOut)); parseErr == nil {
			status.LastSyncTime = t
		}
	}

	return status, nil
}

// writeGitattributes creates the .gitattributes file with merge strategies.
func (g *GitSyncTransport) writeGitattributes() error {
	content := `# ThreeDoors sync merge strategies
tasks.yaml merge=threedoors-task-merge
sessions.jsonl merge=union
`
	path := filepath.Join(g.repoDir, ".gitattributes")
	return os.WriteFile(path, []byte(content), 0o644)
}

// SyncStatus holds the current state of the sync repository.
type SyncStatus struct {
	State         string
	LastSyncTime  time.Time
	UnpushedCount int
	RemoteURL     string
	CircuitState  string
}
