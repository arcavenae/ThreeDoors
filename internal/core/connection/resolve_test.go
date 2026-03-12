package connection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestResolveFromConfig_EmptyConnections(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion: 3,
		Provider:      "textfile",
	}
	reg := core.NewRegistry()

	result, err := ResolveFromConfig(cfg, reg, "/tmp/config.yaml", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty connections")
	}
}

func TestResolveFromConfig_UnregisteredProvider(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion: 3,
		Connections: []core.ConnectionConfig{
			{ID: "conn-1", Provider: "nosuch", Label: "Bad"},
		},
	}
	reg := core.NewRegistry()

	_, err := ResolveFromConfig(cfg, reg, "/tmp/config.yaml", nil)
	if err == nil {
		t.Fatal("expected error for unregistered provider")
	}
}

func TestResolveFromConfig_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write minimal config so persistConfig can read it
	if err := os.WriteFile(configPath, []byte("schema_version: 3\nprovider: textfile\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	reg := core.NewRegistry()
	_ = reg.Register("textfile", func(_ *core.ProviderConfig) (core.TaskProvider, error) {
		return &stubTaskProvider{name: "textfile"}, nil
	})

	cfg := &core.ProviderConfig{
		SchemaVersion: 3,
		Connections: []core.ConnectionConfig{
			{ID: "conn-1", Provider: "textfile", Label: "My Tasks"},
		},
	}

	eventLog := NewSyncEventLog(tmpDir)
	resolved, err := ResolveFromConfig(cfg, reg, configPath, eventLog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved == nil {
		t.Fatal("expected non-nil result")
	}
	if resolved.Manager == nil {
		t.Error("Manager should not be nil")
	}
	if resolved.Bridge == nil {
		t.Error("Bridge should not be nil")
	}
	if resolved.Service == nil {
		t.Error("Service should not be nil")
	}
	if len(resolved.Providers) != 1 {
		t.Errorf("got %d providers, want 1", len(resolved.Providers))
	}

	// Verify provider is registered in bridge
	p := resolved.Bridge.Provider("conn-1")
	if p == nil {
		t.Error("expected provider in bridge for conn-1")
	}

	// Verify connection is in manager
	conn, err := resolved.Manager.Get("conn-1")
	if err != nil {
		t.Fatalf("get connection: %v", err)
	}
	if conn.ProviderName != "textfile" {
		t.Errorf("got provider %q, want %q", conn.ProviderName, "textfile")
	}
	if conn.Label != "My Tasks" {
		t.Errorf("got label %q, want %q", conn.Label, "My Tasks")
	}
	if conn.State != StateDisconnected {
		t.Errorf("got state %v, want %v", conn.State, StateDisconnected)
	}
}

func TestResolveFromConfig_MultipleConnections(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("schema_version: 3\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	reg := core.NewRegistry()
	_ = reg.Register("textfile", func(_ *core.ProviderConfig) (core.TaskProvider, error) {
		return &stubTaskProvider{name: "textfile"}, nil
	})
	_ = reg.Register("jira", func(_ *core.ProviderConfig) (core.TaskProvider, error) {
		return &stubTaskProvider{name: "jira"}, nil
	})

	cfg := &core.ProviderConfig{
		SchemaVersion: 3,
		Connections: []core.ConnectionConfig{
			{ID: "c1", Provider: "textfile", Label: "Local"},
			{ID: "c2", Provider: "jira", Label: "Work Jira"},
		},
	}

	resolved, err := ResolveFromConfig(cfg, reg, configPath, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resolved.Providers) != 2 {
		t.Errorf("got %d providers, want 2", len(resolved.Providers))
	}
	if resolved.Manager.Count() != 2 {
		t.Errorf("got %d connections, want 2", resolved.Manager.Count())
	}
}

func TestResolveFromConfig_PartialFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("schema_version: 3\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	reg := core.NewRegistry()
	_ = reg.Register("textfile", func(_ *core.ProviderConfig) (core.TaskProvider, error) {
		return &stubTaskProvider{name: "textfile"}, nil
	})

	cfg := &core.ProviderConfig{
		SchemaVersion: 3,
		Connections: []core.ConnectionConfig{
			{ID: "c1", Provider: "textfile", Label: "Local"},
			{ID: "c2", Provider: "nosuch", Label: "Bad"},
		},
	}

	resolved, err := ResolveFromConfig(cfg, reg, configPath, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only textfile should succeed
	if len(resolved.Providers) != 1 {
		t.Errorf("got %d providers, want 1", len(resolved.Providers))
	}
}

func TestConnectionProviderConfig(t *testing.T) {
	t.Parallel()

	global := &core.ProviderConfig{
		SchemaVersion: 3,
		NoteTitle:     "Test Notes",
		Theme:         "classic",
	}
	cc := core.ConnectionConfig{
		ID:       "conn-1",
		Provider: "jira",
		Label:    "Work",
		Settings: map[string]string{"url": "https://test.atlassian.net"},
	}

	result := connectionProviderConfig(global, cc)

	if result.Provider != "jira" {
		t.Errorf("got provider %q, want %q", result.Provider, "jira")
	}
	if result.NoteTitle != "Test Notes" {
		t.Errorf("got note title %q, want %q", result.NoteTitle, "Test Notes")
	}
	if len(result.Providers) != 1 {
		t.Fatalf("got %d provider entries, want 1", len(result.Providers))
	}
	if result.Providers[0].Settings["url"] != "https://test.atlassian.net" {
		t.Error("settings not propagated")
	}
}

func TestAddConnectionWithID_EmptyID(t *testing.T) {
	t.Parallel()

	manager := NewConnectionManager(nil)
	cc := core.ConnectionConfig{Provider: "textfile", Label: "Test"}

	_, err := addConnectionWithID(manager, cc)
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}
