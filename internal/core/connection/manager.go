package connection

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// StateChangeCallback is called when a connection's state changes.
type StateChangeCallback func(event StateChangeEvent)

// ConnectionManager manages the lifecycle of all data source connections.
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	onChange    StateChangeCallback
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(onChange StateChangeCallback) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
		onChange:    onChange,
	}
}

// Add creates and stores a new connection with the given provider, label, and settings.
func (m *ConnectionManager) Add(providerName, label string, settings map[string]string) (*Connection, error) {
	conn, err := NewConnection(providerName, label, settings)
	if err != nil {
		return nil, fmt.Errorf("add connection: %w", err)
	}

	m.mu.Lock()
	m.connections[conn.ID] = conn
	m.mu.Unlock()

	return conn, nil
}

// Remove deletes a connection by ID.
func (m *ConnectionManager) Remove(id string) error {
	return m.Disconnect(id, false)
}

// Disconnect removes a connection by ID. If keepTasks is true, the caller
// should preserve synced tasks locally (strip source attribution). If false,
// synced tasks should also be removed. The keepTasks preference is returned
// via DisconnectResult so the caller can act on it.
func (m *ConnectionManager) Disconnect(id string, keepTasks bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.connections[id]; !ok {
		return fmt.Errorf("disconnect connection %s: %w", id, ErrConnectionNotFound)
	}
	delete(m.connections, id)
	return nil
}

// Get returns a connection by ID.
func (m *ConnectionManager) Get(id string) (*Connection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, ok := m.connections[id]
	if !ok {
		return nil, fmt.Errorf("get connection %s: %w", id, ErrConnectionNotFound)
	}
	return conn, nil
}

// List returns all connections sorted by label.
func (m *ConnectionManager) List() []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		result = append(result, conn)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Label < result[j].Label
	})
	return result
}

// Transition changes a connection's state, validating the transition first.
// On success, emits a StateChangeEvent via the callback.
func (m *ConnectionManager) Transition(id string, to ConnectionState) error {
	return m.transitionWithError(id, to, "")
}

// TransitionWithError changes a connection's state and records an error message.
// Use this when transitioning to Error or AuthExpired states.
func (m *ConnectionManager) TransitionWithError(id string, to ConnectionState, errMsg string) error {
	return m.transitionWithError(id, to, errMsg)
}

func (m *ConnectionManager) transitionWithError(id string, to ConnectionState, errMsg string) error {
	m.mu.Lock()

	conn, ok := m.connections[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("transition connection %s: %w", id, ErrConnectionNotFound)
	}

	from := conn.State
	if err := ValidateTransition(from, to); err != nil {
		m.mu.Unlock()
		return fmt.Errorf("transition connection %s: %w", id, err)
	}

	conn.State = to
	if errMsg != "" {
		conn.LastError = errMsg
	}
	if to == StateConnected && from == StateSyncing {
		conn.LastSync = time.Now().UTC()
	}

	m.mu.Unlock()

	if m.onChange != nil {
		m.onChange(StateChangeEvent{
			ConnectionID: id,
			From:         from,
			To:           to,
			Timestamp:    time.Now().UTC(),
			Error:        errMsg,
		})
	}

	return nil
}

// GetByLabel returns the first connection matching the given label (case-insensitive).
func (m *ConnectionManager) GetByLabel(label string) (*Connection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, conn := range m.connections {
		if strings.EqualFold(conn.Label, label) {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("get connection by label %q: %w", label, ErrConnectionNotFound)
}

// NeedsAttention returns connections in Error or AuthExpired state,
// sorted by priority: AuthExpired first, then Error, then by label.
func (m *ConnectionManager) NeedsAttention() []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Connection
	for _, conn := range m.connections {
		if conn.State == StateError || conn.State == StateAuthExpired {
			result = append(result, conn)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		// AuthExpired is higher priority than Error
		if result[i].State != result[j].State {
			return result[i].State == StateAuthExpired
		}
		return result[i].Label < result[j].Label
	})
	return result
}

// Count returns the number of connections.
func (m *ConnectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// ErrConnectionNotFound is returned when a connection ID doesn't exist.
var ErrConnectionNotFound = fmt.Errorf("connection not found")
