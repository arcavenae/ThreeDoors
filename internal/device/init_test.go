package device

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetOrCreateDevice_CreatesNew(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	dev, err := GetOrCreateDevice(configDir)
	if err != nil {
		t.Fatalf("GetOrCreateDevice() unexpected error: %v", err)
	}

	if !dev.ID.IsValid() {
		t.Error("created device should have valid ID")
	}
	if dev.Name == "" {
		t.Error("created device should have non-empty name")
	}
	if dev.FirstSeen.IsZero() {
		t.Error("created device should have non-zero FirstSeen")
	}

	// Verify file was written
	devicePath := filepath.Join(configDir, "device.yaml")
	if _, err := os.Stat(devicePath); err != nil {
		t.Errorf("device.yaml should exist: %v", err)
	}

	// Verify registered in local registry
	devicesDir := filepath.Join(configDir, "devices")
	reg := NewLocalDeviceRegistry(devicesDir)
	devices, err := reg.List()
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("registry should have 1 device, got %d", len(devices))
	}
}

func TestGetOrCreateDevice_LoadsExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	// Create first
	dev1, err := GetOrCreateDevice(configDir)
	if err != nil {
		t.Fatalf("GetOrCreateDevice() first call unexpected error: %v", err)
	}

	// Load existing
	dev2, err := GetOrCreateDevice(configDir)
	if err != nil {
		t.Fatalf("GetOrCreateDevice() second call unexpected error: %v", err)
	}

	if dev1.ID != dev2.ID {
		t.Errorf("second call should return same ID: %s vs %s", dev1.ID, dev2.ID)
	}
}

func TestGetOrCreateDevice_UpgradePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	// Simulate existing config dir with tasks but no device.yaml
	tasksPath := filepath.Join(configDir, "tasks.yaml")
	if err := os.WriteFile(tasksPath, []byte("tasks: []\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}

	dev, err := GetOrCreateDevice(configDir)
	if err != nil {
		t.Fatalf("GetOrCreateDevice() unexpected error: %v", err)
	}

	if !dev.ID.IsValid() {
		t.Error("upgrade path should create valid device")
	}

	// Verify existing files not disturbed
	if _, err := os.Stat(tasksPath); err != nil {
		t.Errorf("existing tasks.yaml should not be disturbed: %v", err)
	}
}

func TestGetOrCreateDevice_CorruptedYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	// Write corrupted device.yaml
	devicePath := filepath.Join(configDir, "device.yaml")
	if err := os.WriteFile(devicePath, []byte("id: [broken yaml {{{"), 0o600); err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}

	dev, err := GetOrCreateDevice(configDir)
	if err != nil {
		t.Fatalf("GetOrCreateDevice() should recover from corrupted YAML: %v", err)
	}

	if !dev.ID.IsValid() {
		t.Error("recovered device should have valid ID")
	}

	// Verify backup was created
	backupPath := devicePath + ".bak"
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("corrupted device.yaml should be backed up: %v", err)
	}
}
