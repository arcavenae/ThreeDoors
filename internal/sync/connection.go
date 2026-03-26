package sync

import (
	"context"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/core/connection"
)

const (
	// ProviderName is the connection provider name for git sync.
	ProviderName = "git-sync"
)

// GitSyncConnection adapts GitSyncTransport for the Connection Manager.
// It manages state transitions and integrates with the circuit breaker
// and sync status tracker.
type GitSyncConnection struct {
	transport *GitSyncTransport
	connMgr   *connection.ConnectionManager
	tracker   *core.SyncStatusTracker
	breaker   *core.CircuitBreaker
	connID    string
}

// GitSyncConnectionConfig holds configuration for creating a GitSyncConnection.
type GitSyncConnectionConfig struct {
	Transport *GitSyncTransport
	ConnMgr   *connection.ConnectionManager
	Tracker   *core.SyncStatusTracker
	Breaker   *core.CircuitBreaker
}

// NewGitSyncConnection creates a new connection adapter.
func NewGitSyncConnection(cfg GitSyncConnectionConfig) *GitSyncConnection {
	return &GitSyncConnection{
		transport: cfg.Transport,
		connMgr:   cfg.ConnMgr,
		tracker:   cfg.Tracker,
		breaker:   cfg.Breaker,
	}
}

// Register adds the git-sync connection to the Connection Manager and tracker.
func (c *GitSyncConnection) Register(remoteURL string) error {
	conn, err := c.connMgr.Add(ProviderName, "Git Sync", map[string]string{
		"remote_url": remoteURL,
	})
	if err != nil {
		return err
	}
	c.connID = conn.ID

	if c.tracker != nil {
		c.tracker.Register(ProviderName)
	}

	return nil
}

// Connect initializes the sync transport.
func (c *GitSyncConnection) Connect(ctx context.Context) error {
	if err := c.connMgr.Transition(c.connID, connection.StateConnecting); err != nil {
		return err
	}
	if c.tracker != nil {
		c.tracker.SetSyncing(ProviderName)
	}

	if err := c.transport.Init(ctx); err != nil {
		_ = c.connMgr.TransitionWithError(c.connID, connection.StateError, err.Error())
		if c.tracker != nil {
			c.tracker.SetError(ProviderName, err.Error())
		}
		return err
	}

	if err := c.connMgr.Transition(c.connID, connection.StateConnected); err != nil {
		return err
	}
	if c.tracker != nil {
		c.tracker.SetSynced(ProviderName)
	}
	return nil
}

// Sync pushes local changes through the circuit breaker.
func (c *GitSyncConnection) Sync(ctx context.Context, changeset Changeset) error {
	if err := c.connMgr.Transition(c.connID, connection.StateSyncing); err != nil {
		return err
	}
	if c.tracker != nil {
		c.tracker.SetSyncing(ProviderName)
	}

	err := c.breaker.Execute(func() error {
		return c.transport.Push(ctx, changeset)
	})
	if err != nil {
		_ = c.connMgr.TransitionWithError(c.connID, connection.StateError, err.Error())
		if c.tracker != nil {
			c.tracker.SetError(ProviderName, err.Error())
		}
		return err
	}

	if err := c.connMgr.Transition(c.connID, connection.StateConnected); err != nil {
		return err
	}
	if c.tracker != nil {
		c.tracker.SetSynced(ProviderName)
	}
	return nil
}

// Pull retrieves remote changes through the circuit breaker.
func (c *GitSyncConnection) Pull(ctx context.Context, since time.Time) (Changeset, error) {
	if err := c.connMgr.Transition(c.connID, connection.StateSyncing); err != nil {
		return Changeset{}, err
	}
	if c.tracker != nil {
		c.tracker.SetSyncing(ProviderName)
	}

	var cs Changeset
	err := c.breaker.Execute(func() error {
		var pullErr error
		cs, pullErr = c.transport.Pull(ctx, since)
		return pullErr
	})
	if err != nil {
		_ = c.connMgr.TransitionWithError(c.connID, connection.StateError, err.Error())
		if c.tracker != nil {
			c.tracker.SetError(ProviderName, err.Error())
		}
		return Changeset{}, err
	}

	if err := c.connMgr.Transition(c.connID, connection.StateConnected); err != nil {
		return cs, err
	}
	if c.tracker != nil {
		c.tracker.SetSynced(ProviderName)
	}
	return cs, nil
}

// ConnectionID returns the connection manager ID for this sync connection.
func (c *GitSyncConnection) ConnectionID() string {
	return c.connID
}
