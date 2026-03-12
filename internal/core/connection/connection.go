package connection

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// Connection represents a configured instance of a data source.
type Connection struct {
	ID           string            `json:"id"`
	ProviderName string            `json:"provider"`
	Label        string            `json:"label"`
	State        ConnectionState   `json:"state"`
	LastSync     time.Time         `json:"last_sync"`
	LastError    string            `json:"last_error,omitempty"`
	SyncMode     string            `json:"sync_mode"`
	PollInterval time.Duration     `json:"poll_interval"`
	Settings     map[string]string `json:"settings,omitempty"`
	TaskCount    int               `json:"task_count"`
	CreatedAt    time.Time         `json:"created_at"`
}

// NewConnection creates a Connection with a ULID ID and state Disconnected.
func NewConnection(providerName, label string, settings map[string]string) (*Connection, error) {
	if providerName == "" {
		return nil, fmt.Errorf("create connection: provider name must not be empty")
	}
	if label == "" {
		return nil, fmt.Errorf("create connection: label must not be empty")
	}

	id, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("create connection: generate ID: %w", err)
	}

	s := make(map[string]string, len(settings))
	for k, v := range settings {
		s[k] = v
	}

	now := time.Now().UTC()
	return &Connection{
		ID:           id.String(),
		ProviderName: providerName,
		Label:        label,
		State:        StateDisconnected,
		SyncMode:     "readonly",
		PollInterval: 5 * time.Minute,
		Settings:     s,
		CreatedAt:    now,
	}, nil
}

// StateChangeEvent is emitted when a connection's state changes.
type StateChangeEvent struct {
	ConnectionID string          `json:"connection_id"`
	From         ConnectionState `json:"from"`
	To           ConnectionState `json:"to"`
	Timestamp    time.Time       `json:"timestamp"`
	Error        string          `json:"error,omitempty"`
}
