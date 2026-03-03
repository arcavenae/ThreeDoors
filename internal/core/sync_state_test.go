package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncState_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	// Ensure config dir exists
	configPath := filepath.Join(tempDir, configDir)
	_ = os.MkdirAll(configPath, 0o755)

	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusInProgress, laterTime)

	original := newTestSyncState(taskA, taskB)
	original.LastSyncTime = time.Date(2026, 3, 1, 15, 0, 0, 0, time.UTC)

	// Save
	err := SaveSyncState(original)
	if err != nil {
		t.Fatalf("SaveSyncState() failed: %v", err)
	}

	// Load
	loaded, err := LoadSyncState()
	if err != nil {
		t.Fatalf("LoadSyncState() failed: %v", err)
	}

	// Verify LastSyncTime
	if !loaded.LastSyncTime.Equal(original.LastSyncTime) {
		t.Errorf("LastSyncTime = %v, want %v", loaded.LastSyncTime, original.LastSyncTime)
	}

	// Verify task snapshots
	if len(loaded.TaskSnapshots) != 2 {
		t.Fatalf("TaskSnapshots count = %d, want 2", len(loaded.TaskSnapshots))
	}

	snapA, ok := loaded.TaskSnapshots["aaa"]
	if !ok {
		t.Fatal("TaskSnapshot for 'aaa' not found")
	}
	if snapA.Text != "Task A" {
		t.Errorf("Snapshot A text = %q, want %q", snapA.Text, "Task A")
	}
	if snapA.Status != StatusTodo {
		t.Errorf("Snapshot A status = %q, want %q", snapA.Status, StatusTodo)
	}
	if !snapA.UpdatedAt.Equal(baseTime) {
		t.Errorf("Snapshot A UpdatedAt = %v, want %v", snapA.UpdatedAt, baseTime)
	}

	snapB, ok := loaded.TaskSnapshots["bbb"]
	if !ok {
		t.Fatal("TaskSnapshot for 'bbb' not found")
	}
	if snapB.Status != StatusInProgress {
		t.Errorf("Snapshot B status = %q, want %q", snapB.Status, StatusInProgress)
	}
}

func TestSyncState_SaveAndLoadWithDirtyFlags(t *testing.T) {
	tempDir := t.TempDir()
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	configPath := filepath.Join(tempDir, configDir)
	_ = os.MkdirAll(configPath, 0o755)

	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	original := newDirtySyncState([]string{"aaa"}, taskA, taskB)

	err := SaveSyncState(original)
	if err != nil {
		t.Fatalf("SaveSyncState() failed: %v", err)
	}

	loaded, err := LoadSyncState()
	if err != nil {
		t.Fatalf("LoadSyncState() failed: %v", err)
	}

	snapA := loaded.TaskSnapshots["aaa"]
	if !snapA.Dirty {
		t.Error("Snapshot A should be dirty")
	}

	snapB := loaded.TaskSnapshots["bbb"]
	if snapB.Dirty {
		t.Error("Snapshot B should not be dirty")
	}
}

func TestLoadSyncState_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	configPath := filepath.Join(tempDir, configDir)
	_ = os.MkdirAll(configPath, 0o755)

	state, err := LoadSyncState()
	if err != nil {
		t.Fatalf("LoadSyncState() should not return error for missing file, got: %v", err)
	}

	if state.TaskSnapshots == nil {
		t.Error("TaskSnapshots should be initialized (not nil)")
	}
	if len(state.TaskSnapshots) != 0 {
		t.Errorf("TaskSnapshots should be empty, got %d entries", len(state.TaskSnapshots))
	}
	if !state.LastSyncTime.IsZero() {
		t.Errorf("LastSyncTime should be zero for new state, got %v", state.LastSyncTime)
	}
}

func TestLoadSyncState_CorruptYAML(t *testing.T) {
	tempDir := t.TempDir()
	SetHomeDir(tempDir)
	defer SetHomeDir("")

	configPath := filepath.Join(tempDir, configDir)
	_ = os.MkdirAll(configPath, 0o755)

	// Write corrupt YAML
	syncPath := filepath.Join(configPath, "sync_state.yaml")
	_ = os.WriteFile(syncPath, []byte("{{{{not yaml!"), 0o644)

	state, err := LoadSyncState()
	// Should return empty state, not crash. Error is acceptable but state must be usable.
	if err != nil {
		// Error is OK for corrupt file, but state should still be usable
		if state.TaskSnapshots == nil {
			t.Error("TaskSnapshots should be initialized even on error")
		}
		return
	}

	if len(state.TaskSnapshots) != 0 {
		t.Errorf("TaskSnapshots should be empty for corrupt file, got %d", len(state.TaskSnapshots))
	}
}
