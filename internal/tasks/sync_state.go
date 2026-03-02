package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const syncStateFile = "sync_state.yaml"

// SyncState tracks the last-known synced state for three-way sync comparison.
type SyncState struct {
	LastSyncTime  time.Time               `yaml:"last_sync_time"`
	TaskSnapshots map[string]TaskSnapshot `yaml:"task_snapshots"`
}

// TaskSnapshot records the state of a task at last sync time.
type TaskSnapshot struct {
	ID        string     `yaml:"id"`
	Text      string     `yaml:"text"`
	Status    TaskStatus `yaml:"status"`
	UpdatedAt time.Time  `yaml:"updated_at"`
	Dirty     bool       `yaml:"dirty"` // true if modified locally since last sync
}

// SaveSyncState persists the sync state to ~/.threedoors/sync_state.yaml using atomic write.
func SaveSyncState(state SyncState) error {
	configPath, err := EnsureConfigDir()
	if err != nil {
		return fmt.Errorf("sync state: %w", err)
	}

	statePath := filepath.Join(configPath, syncStateFile)
	tmpPath := statePath + ".tmp"

	data, err := yaml.Marshal(&state)
	if err != nil {
		return fmt.Errorf("sync state marshal: %w", err)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("sync state create temp: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync state write: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync state sync: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, statePath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync state rename: %w", err)
	}
	return nil
}

// LoadSyncState reads the sync state from ~/.threedoors/sync_state.yaml.
// Returns an empty SyncState if the file doesn't exist.
// Returns an empty SyncState with logged warning if the file is corrupt.
func LoadSyncState() (SyncState, error) {
	empty := SyncState{TaskSnapshots: make(map[string]TaskSnapshot)}

	configPath, err := GetConfigDirPath()
	if err != nil {
		return empty, nil
	}

	statePath := filepath.Join(configPath, syncStateFile)
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return empty, nil
		}
		return empty, fmt.Errorf("sync state read: %w", err)
	}

	var state SyncState
	if err := yaml.Unmarshal(data, &state); err != nil {
		// Corrupt file - return empty state
		return empty, fmt.Errorf("sync state parse: %w", err)
	}

	if state.TaskSnapshots == nil {
		state.TaskSnapshots = make(map[string]TaskSnapshot)
	}

	return state, nil
}
