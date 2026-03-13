package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/device"
	"github.com/spf13/cobra"
)

func TestSyncCmdRegistered(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "sync" {
			found = true
			break
		}
	}
	if !found {
		t.Error("sync command not registered on root")
	}
}

func TestSyncSubcommands(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var syncCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "sync" {
			syncCmd = cmd
			break
		}
	}
	if syncCmd == nil {
		t.Fatal("sync command not found")
	}

	expected := map[string]bool{"init": false, "push": false, "status": false}
	for _, sub := range syncCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("sync subcommand %q not registered", name)
		}
	}
}

func TestSyncInitCmd_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"sync", "init"})

	err := cmd.Execute()
	if err == nil {
		t.Error("sync init without args should return error")
	}
}

func TestSyncStatusCmd_NotInitialized(t *testing.T) {
	// Do not run in parallel — modifies global testHomeDir
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	setupDeviceYAML(t, dir)

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"sync", "status"})

	err := cmd.Execute()
	if err == nil {
		t.Error("sync status without init should return error")
	}
}

func TestSyncInitCmd_ValidatesURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"ssh url", "git@github.com:user/repo.git", false},
		{"https url", "https://github.com/user/repo.git", false},
		{"local path", "/tmp/repo.git", false},
		{"invalid scheme", "ftp://example.com/repo", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateRemoteURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRemoteURL(%q) err=%v, wantErr=%v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestSyncInitCmd_Integration(t *testing.T) {
	// Do not run in parallel — modifies global testHomeDir
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	setupDeviceYAML(t, dir)

	bareDir := filepath.Join(dir, "remote.git")
	gitCmd := exec.Command("git", "init", "--bare", bareDir)
	if out, err := gitCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"sync", "init", bareDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync init error: %v\noutput: %s", err, buf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "Sync initialized") {
		t.Errorf("expected 'Sync initialized' in output, got: %s", output)
	}

	// Verify sync config was persisted
	configDir := filepath.Join(dir, ".threedoors")
	cfgData, err := os.ReadFile(filepath.Join(configDir, "sync.json"))
	if err != nil {
		t.Fatalf("sync config not saved: %v", err)
	}
	var cfg syncConfig
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		t.Fatalf("parse sync config: %v", err)
	}
	if cfg.RemoteURL != bareDir {
		t.Errorf("sync config remote_url = %q, want %q", cfg.RemoteURL, bareDir)
	}
	if !cfg.Enabled {
		t.Error("sync config should be enabled")
	}
}

func TestSyncInitCmd_JSON(t *testing.T) {
	// Do not run in parallel — modifies global testHomeDir
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	setupDeviceYAML(t, dir)

	bareDir := filepath.Join(dir, "remote.git")
	gitCmd := exec.Command("git", "init", "--bare", bareDir)
	if out, err := gitCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"sync", "init", bareDir, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync init --json error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON output invalid: %v\noutput: %s", err, buf.String())
	}
	if result["status"] != "initialized" {
		t.Errorf("JSON status = %q, want 'initialized'", result["status"])
	}
}

// setupDeviceYAML creates a minimal device.yaml in the test config directory.
func setupDeviceYAML(t *testing.T, homeDir string) {
	t.Helper()
	configDir := filepath.Join(homeDir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	id, err := device.NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID: %v", err)
	}

	dev := &device.Device{
		ID:        id,
		Name:      "test-device",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	devPath := filepath.Join(configDir, "device.yaml")
	if err := device.SaveDevice(dev, devPath); err != nil {
		t.Fatalf("save device: %v", err)
	}
}
