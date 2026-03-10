package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureConfigDir_CreatesWithRestrictivePermissions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })

	configPath, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("config dir permissions = %04o, want 0700", perm)
	}
}

func TestEnsureConfigDir_TightensPermissiveDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })

	// Pre-create directory with old permissive permissions
	configPath := filepath.Join(tmpDir, configDir)
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Verify it's permissive before the call
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("pre-condition: expected 0755, got %04o", info.Mode().Perm())
	}

	// EnsureConfigDir should tighten it
	result, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error: %v", err)
	}
	if result != configPath {
		t.Fatalf("unexpected path: got %s, want %s", result, configPath)
	}

	info, err = os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat after: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("after tightening: permissions = %04o, want 0700", perm)
	}
}

func TestEnsureConfigDir_NoChangeWhenAlreadyRestrictive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })

	// Pre-create directory with correct permissions
	configPath := filepath.Join(tmpDir, configDir)
	if err := os.MkdirAll(configPath, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	_, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("permissions = %04o, want 0700", perm)
	}
}
