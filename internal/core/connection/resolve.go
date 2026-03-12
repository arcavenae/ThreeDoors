package connection

import (
	"fmt"
	"log"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// ResolvedConnections holds the result of resolving config connections into
// a fully wired ConnectionManager with providers.
type ResolvedConnections struct {
	Manager   *ConnectionManager
	Bridge    *ProviderBridge
	Service   *ConnectionService
	Providers map[string]core.TaskProvider // connection ID → provider
}

// ResolveFromConfig reads the connections[] array from a ProviderConfig,
// creates a TaskProvider for each via the Registry, registers them with a
// ConnectionManager, and returns the fully wired result.
//
// Connections that fail to initialize are logged and skipped — the app
// continues with whichever connections succeed.
//
// If cfg.Connections is empty, returns nil (caller should use legacy path).
func ResolveFromConfig(cfg *core.ProviderConfig, reg *core.Registry, configPath string, eventLog *SyncEventLog) (*ResolvedConnections, error) {
	if len(cfg.Connections) == 0 {
		return nil, nil
	}

	manager := NewConnectionManager(nil)
	bridge := NewProviderBridge()
	providers := make(map[string]core.TaskProvider, len(cfg.Connections))

	var successCount int
	for _, cc := range cfg.Connections {
		if !reg.IsRegistered(cc.Provider) {
			log.Printf("Warning: provider %q not registered, skipping connection %q", cc.Provider, cc.Label)
			continue
		}

		// Build a ProviderConfig scoped to this connection's settings.
		connCfg := connectionProviderConfig(cfg, cc)

		provider, err := reg.InitProvider(cc.Provider, connCfg)
		if err != nil {
			log.Printf("Warning: connection %q (%s) failed to initialize: %v", cc.Label, cc.Provider, err)
			continue
		}

		// Register with ConnectionManager using the config-defined ID.
		conn, err := addConnectionWithID(manager, cc)
		if err != nil {
			log.Printf("Warning: connection %q (%s) failed to add to manager: %v", cc.Label, cc.Provider, err)
			continue
		}

		bridge.Register(conn.ID, provider)
		providers[conn.ID] = provider
		successCount++
	}

	if successCount == 0 {
		return nil, fmt.Errorf("no connections could be initialized from config")
	}

	svc, err := NewConnectionService(ServiceConfig{
		Manager:    manager,
		Creds:      NewEnvCredentialStore(),
		ConfigPath: configPath,
		EventLog:   eventLog,
		Checker:    bridge,
		Syncer:     bridge,
	})
	if err != nil {
		return nil, fmt.Errorf("create connection service: %w", err)
	}

	return &ResolvedConnections{
		Manager:   manager,
		Bridge:    bridge,
		Service:   svc,
		Providers: providers,
	}, nil
}

// addConnectionWithID creates a Connection with the ID from config rather than
// generating a new ULID. This ensures IDs are stable across restarts.
func addConnectionWithID(manager *ConnectionManager, cc core.ConnectionConfig) (*Connection, error) {
	if cc.ID == "" {
		return nil, fmt.Errorf("connection config missing ID")
	}

	s := make(map[string]string, len(cc.Settings))
	for k, v := range cc.Settings {
		s[k] = v
	}

	conn := &Connection{
		ID:           cc.ID,
		ProviderName: cc.Provider,
		Label:        cc.Label,
		State:        StateDisconnected,
		SyncMode:     "readonly",
		PollInterval: 5 * 60e9, // 5 minutes
		Settings:     s,
		CreatedAt:    time.Now().UTC(),
	}

	manager.mu.Lock()
	manager.connections[conn.ID] = conn
	manager.mu.Unlock()

	return conn, nil
}

// connectionProviderConfig creates a ProviderConfig scoped to a specific
// connection. It copies global settings and injects the connection's settings
// into the Providers list so existing adapter factories can read them.
func connectionProviderConfig(global *core.ProviderConfig, cc core.ConnectionConfig) *core.ProviderConfig {
	return &core.ProviderConfig{
		SchemaVersion: global.SchemaVersion,
		Provider:      cc.Provider,
		NoteTitle:     global.NoteTitle,
		Providers: []core.ProviderEntry{
			{
				Name:     cc.Provider,
				Settings: cc.Settings,
			},
		},
		LLM:   global.LLM,
		Theme: global.Theme,
	}
}
