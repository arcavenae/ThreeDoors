package device

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDevice_SaveAndLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "device.yaml")

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	now := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)
	dev := &Device{
		ID:        id,
		Name:      "test-laptop",
		FirstSeen: now,
		LastSync:  now,
	}

	if err := SaveDevice(dev, path); err != nil {
		t.Fatalf("SaveDevice() unexpected error: %v", err)
	}

	loaded, err := LoadDevice(path)
	if err != nil {
		t.Fatalf("LoadDevice() unexpected error: %v", err)
	}

	if loaded.ID != dev.ID {
		t.Errorf("loaded ID = %s, want %s", loaded.ID, dev.ID)
	}
	if loaded.Name != dev.Name {
		t.Errorf("loaded Name = %s, want %s", loaded.Name, dev.Name)
	}
	if !loaded.FirstSeen.Equal(dev.FirstSeen) {
		t.Errorf("loaded FirstSeen = %v, want %v", loaded.FirstSeen, dev.FirstSeen)
	}
	if !loaded.LastSync.Equal(dev.LastSync) {
		t.Errorf("loaded LastSync = %v, want %v", loaded.LastSync, dev.LastSync)
	}
}

func TestDevice_LoadCorrupted(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "corrupted_device.yaml")

	_, err := LoadDevice(path)
	if err == nil {
		t.Error("LoadDevice() should return error for corrupted YAML")
	}
}

func TestDevice_LoadPartial(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "partial_device.yaml")

	dev, err := LoadDevice(path)
	if err != nil {
		t.Fatalf("LoadDevice() unexpected error: %v", err)
	}

	if dev.ID.IsValid() {
		t.Error("device loaded from partial YAML should have invalid ID")
	}
	if dev.Name != "missing-id-device" {
		t.Errorf("device Name = %s, want missing-id-device", dev.Name)
	}
}

func TestDevice_LoadNonExistent(t *testing.T) {
	t.Parallel()

	_, err := LoadDevice("/nonexistent/path/device.yaml")
	if err == nil {
		t.Error("LoadDevice() should return error for non-existent file")
	}
}

func TestDevice_SaveAtomicWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "device.yaml")

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := &Device{
		ID:        id,
		Name:      "test-laptop",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := SaveDevice(dev, path); err != nil {
		t.Fatalf("SaveDevice() unexpected error: %v", err)
	}

	// Verify no .tmp files left behind
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() unexpected error: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("temporary file left behind: %s", entry.Name())
		}
	}
}

func TestDevice_DefaultName(t *testing.T) {
	t.Parallel()

	hostname, err := os.Hostname()
	if err != nil {
		t.Skipf("cannot get hostname: %v", err)
	}

	name := DefaultDeviceName()
	if name != hostname {
		t.Errorf("DefaultDeviceName() = %s, want %s", name, hostname)
	}
}
