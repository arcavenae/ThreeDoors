package connection

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// Connection represents a configured instance of a data source.
type Connection struct {
	ID           string // unique instance ID (ULID)
	ProviderName string // adapter name: "jira", "todoist", "github"
	Label        string // user-friendly: "Work Jira", "Personal Todoist"
	State        ConnectionState
	LastSync     time.Time
	LastError    string
	SyncMode     string // "bidirectional", "readonly"
	PollInterval time.Duration
	Settings     map[string]string // provider-specific non-secret config
	TaskCount    int               // cached active task count
	CreatedAt    time.Time
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
	ConnectionID string
	From         ConnectionState
	To           ConnectionState
	Timestamp    time.Time
	Error        string // populated when transitioning to Error or AuthExpired
}
