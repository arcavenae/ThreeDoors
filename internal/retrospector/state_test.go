package retrospector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadState_NewFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "state.json")

	state, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if state.LastProcessedPR != 0 {
		t.Errorf("LastProcessedPR = %d, want 0", state.LastProcessedPR)
	}
}

func TestState_SaveAndLoad(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "state.json")

	state, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	state.LastProcessedPR = 42
	if err := state.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState() after save error = %v", err)
	}
	if loaded.LastProcessedPR != 42 {
		t.Errorf("LastProcessedPR = %d, want 42", loaded.LastProcessedPR)
	}
}

func TestState_Save_AtomicWrite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "state.json")

	state, _ := LoadState(path)
	state.LastProcessedPR = 100
	if err := state.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify no .tmp file remains
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("Expected .tmp file to be cleaned up after atomic write")
	}
}

func TestState_FilePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "state.json")

	state, _ := LoadState(path)
	state.LastProcessedPR = 1
	if err := state.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("File permissions = %o, want 600", perm)
	}
}
