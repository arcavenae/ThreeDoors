package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureConfigDir_CreatesWithRestrictivePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })

	configPath, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir: %v", err)
		return
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
		return
	}

	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("config dir permissions = %o, want 0700", perm)
	}
}

func TestEnsureConfigDir_MigratesPermissiveDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })

	// Create directory with permissive permissions (simulating old version)
	configPath := filepath.Join(tmpDir, configDir)
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		t.Fatalf("create permissive dir: %v", err)
	}

	// Verify it starts permissive
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
		return
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("setup: expected 0755, got %o", info.Mode().Perm())
	}

	// EnsureConfigDir should tighten permissions
	result, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir: %v", err)
		return
	}
	if result != configPath {
		t.Fatalf("EnsureConfigDir returned %q, want %q", result, configPath)
	}

	info, err = os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat after migration: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("after migration: permissions = %o, want 0700", perm)
	}
}

func TestEnsureConfigDir_NoChangeWhenAlreadyRestrictive(t *testing.T) {
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })

	// Create directory with correct permissions
	configPath := filepath.Join(tmpDir, configDir)
	if err := os.MkdirAll(configPath, 0o700); err != nil {
		t.Fatalf("create restrictive dir: %v", err)
	}

	_, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir: %v", err)
		return
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("permissions = %o, want 0700", perm)
	}
}
