package sync

import (
	"context"
	"time"

	"github.com/arcaven/ThreeDoors/internal/device"
)

// SyncOp represents the type of file change in a sync operation.
type SyncOp int

const (
	// OpAdd indicates a new file was added.
	OpAdd SyncOp = iota
	// OpModify indicates an existing file was modified.
	OpModify
	// OpDelete indicates a file was deleted.
	OpDelete
)

// String returns the human-readable name of the sync operation.
func (o SyncOp) String() string {
	switch o {
	case OpAdd:
		return "add"
	case OpModify:
		return "modify"
	case OpDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// SyncFile represents a single file change in a sync changeset.
type SyncFile struct {
	Path    string
	Content []byte
	Op      SyncOp
}

// Changeset contains the set of file changes for a sync operation.
type Changeset struct {
	DeviceID  device.DeviceID
	Timestamp time.Time
	Files     []SyncFile
}

// SyncTransport defines the interface for syncing data between devices.
type SyncTransport interface {
	// Push sends local changes to the remote.
	Push(ctx context.Context, changeset Changeset) error
	// Pull retrieves remote changes since the given time.
	Pull(ctx context.Context, since time.Time) (Changeset, error)
}
