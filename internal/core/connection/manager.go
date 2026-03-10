package connection

import (
	"fmt"
	"sort"
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
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.connections[id]; !ok {
		return fmt.Errorf("remove connection %s: %w", id, ErrConnectionNotFound)
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

// Count returns the number of connections.
func (m *ConnectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// ErrConnectionNotFound is returned when a connection ID doesn't exist.
var ErrConnectionNotFound = fmt.Errorf("connection not found")
