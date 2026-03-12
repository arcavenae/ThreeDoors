package connection

import (
	"fmt"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// ConnectionService orchestrates CRUD operations across ConnectionManager,
// CredentialStore, and config persistence. It is the single entry point for
// callers who need to add, remove, pause, resume, test, or force-sync a
// connection.
type ConnectionService struct {
	manager    *ConnectionManager
	creds      CredentialStore
	configPath string
	eventLog   *SyncEventLog
	checker    HealthChecker
	syncer     Syncer
}

// ServiceConfig holds the dependencies for creating a ConnectionService.
type ServiceConfig struct {
	Manager    *ConnectionManager
	Creds      CredentialStore
	ConfigPath string
	EventLog   *SyncEventLog
	Checker    HealthChecker
	Syncer     Syncer
}

// NewConnectionService creates a ConnectionService from the given dependencies.
func NewConnectionService(cfg ServiceConfig) (*ConnectionService, error) {
	if cfg.Manager == nil {
		return nil, fmt.Errorf("create connection service: manager must not be nil")
	}
	if cfg.Creds == nil {
		return nil, fmt.Errorf("create connection service: credential store must not be nil")
	}
	if cfg.ConfigPath == "" {
		return nil, fmt.Errorf("create connection service: config path must not be empty")
	}
	return &ConnectionService{
		manager:    cfg.Manager,
		creds:      cfg.Creds,
		configPath: cfg.ConfigPath,
		eventLog:   cfg.EventLog,
		checker:    cfg.Checker,
		syncer:     cfg.Syncer,
	}, nil
}

// EventLog returns the service's SyncEventLog, or nil if none was configured.
func (s *ConnectionService) EventLog() *SyncEventLog {
	return s.eventLog
}

// Add creates a new connection, stores its credential, and persists config.
// credential may be empty if the provider uses env-var-based auth.
func (s *ConnectionService) Add(providerName, label string, settings map[string]string, credential string) (*Connection, error) {
	conn, err := s.manager.Add(providerName, label, settings)
	if err != nil {
		return nil, fmt.Errorf("service add: %w", err)
	}

	if credential != "" {
		credKey := ConnCredentialKey(conn)
		if err := s.creds.Set(credKey, credential); err != nil {
			// Roll back: remove the connection from the manager.
			_ = s.manager.Remove(conn.ID)
			return nil, fmt.Errorf("service add credential for %s: %w", conn.ID, err)
		}
	}

	if err := s.persistConfig(); err != nil {
		// Roll back: remove credential and connection.
		_ = s.creds.Delete(ConnCredentialKey(conn))
		_ = s.manager.Remove(conn.ID)
		return nil, fmt.Errorf("service add persist config: %w", err)
	}

	return conn, nil
}

// Remove deletes a connection. When keepTasks is true, local tasks from this
// connection are retained; otherwise they would be cleaned up by the caller.
// Credentials are always deleted.
func (s *ConnectionService) Remove(id string, keepTasks bool) error {
	conn, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("service remove: %w", err)
	}

	credKey := ConnCredentialKey(conn)

	if err := s.manager.Remove(id); err != nil {
		return fmt.Errorf("service remove: %w", err)
	}

	// Best-effort credential deletion — log errors but don't fail the operation.
	_ = s.creds.Delete(credKey)

	if err := s.persistConfig(); err != nil {
		return fmt.Errorf("service remove persist config: %w", err)
	}

	return nil
}

// Pause transitions a connected connection to the Paused state, stopping sync polling.
func (s *ConnectionService) Pause(id string) error {
	if err := s.manager.Transition(id, StatePaused); err != nil {
		return fmt.Errorf("service pause: %w", err)
	}

	if s.eventLog != nil {
		_ = s.eventLog.LogStateChange(id, StateConnected, StatePaused, "")
	}

	return nil
}

// Resume transitions a paused connection back to Connected, resuming sync.
func (s *ConnectionService) Resume(id string) error {
	if err := s.manager.Transition(id, StateConnected); err != nil {
		return fmt.Errorf("service resume: %w", err)
	}

	if s.eventLog != nil {
		_ = s.eventLog.LogStateChange(id, StatePaused, StateConnected, "")
	}

	return nil
}

// TestConnection performs a lightweight health check and returns the result.
// Returns an error if no HealthChecker is configured.
func (s *ConnectionService) TestConnection(id string) (HealthCheckResult, error) {
	if s.checker == nil {
		return HealthCheckResult{}, fmt.Errorf("service test connection: no health checker configured")
	}

	conn, err := s.manager.Get(id)
	if err != nil {
		return HealthCheckResult{}, fmt.Errorf("service test connection: %w", err)
	}

	credKey := ConnCredentialKey(conn)
	cred, err := s.creds.Get(credKey)
	if err != nil {
		// Credential may not be required for all providers; pass empty.
		cred = ""
	}

	result, err := s.checker.CheckHealth(conn, cred)
	if err != nil {
		return HealthCheckResult{}, fmt.Errorf("service test connection %s: %w", id, err)
	}

	return result, nil
}

// ForceSync triggers an immediate sync cycle for a connection outside its
// normal polling interval. Returns an error if no Syncer is configured.
func (s *ConnectionService) ForceSync(id string) error {
	if s.syncer == nil {
		return fmt.Errorf("service force sync: no syncer configured")
	}

	conn, err := s.manager.Get(id)
	if err != nil {
		return fmt.Errorf("service force sync: %w", err)
	}

	// Must be in Connected state to sync.
	if conn.State != StateConnected {
		return fmt.Errorf("service force sync %s: connection must be in connected state, got %s", id, conn.State)
	}

	if err := s.manager.Transition(id, StateSyncing); err != nil {
		return fmt.Errorf("service force sync transition: %w", err)
	}

	if s.eventLog != nil {
		_ = s.eventLog.Append(SyncEvent{
			ConnectionID: id,
			Type:         EventSyncStart,
			Summary:      "Force sync triggered",
		})
	}

	credKey := ConnCredentialKey(conn)
	cred, err := s.creds.Get(credKey)
	if err != nil {
		cred = ""
	}

	if err := s.syncer.Sync(conn, cred); err != nil {
		_ = s.manager.TransitionWithError(id, StateError, err.Error())
		if s.eventLog != nil {
			_ = s.eventLog.LogSyncError(id, err)
		}
		return fmt.Errorf("service force sync %s: %w", id, err)
	}

	if err := s.manager.Transition(id, StateConnected); err != nil {
		return fmt.Errorf("service force sync complete transition: %w", err)
	}

	return nil
}

// persistConfig loads the current config, updates connections from the manager's
// state, and saves it atomically.
func (s *ConnectionService) persistConfig() error {
	cfg, err := core.LoadProviderConfig(s.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Rebuild the connections list from the manager's current state.
	conns := s.manager.List()
	cfg.Connections = make([]core.ConnectionConfig, 0, len(conns))
	for _, c := range conns {
		cfg.Connections = append(cfg.Connections, core.ConnectionConfig{
			ID:       c.ID,
			Provider: c.ProviderName,
			Label:    c.Label,
			Settings: c.Settings,
		})
	}

	if err := core.SaveProviderConfig(s.configPath, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}
