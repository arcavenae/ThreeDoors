package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/device"
)

func TestDevicesCmd_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"devices"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No devices") && !strings.Contains(output, "NAME") {
		t.Errorf("unexpected output for empty device list: %s", output)
	}
}

func TestDevicesCmd_ListWithDevices(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	devicesDir := filepath.Join(dir, ".threedoors", "devices")
	if err := os.MkdirAll(devicesDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	reg := device.NewLocalDeviceRegistry(devicesDir)

	id, err := device.NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := device.Device{
		ID:        id,
		Name:      "test-laptop",
		FirstSeen: time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC),
		LastSync:  time.Date(2026, 3, 13, 14, 30, 0, 0, time.UTC),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"devices"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-laptop") {
		t.Errorf("output should contain device name, got: %s", output)
	}
	if !strings.Contains(output, "550e8400") {
		t.Errorf("output should contain device ID prefix, got: %s", output)
	}
}

func TestDevicesCmd_ListJSON(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"devices", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(output), "[") {
		t.Errorf("JSON output should start with [, got: %s", output)
	}
}

func TestDevicesRenameCmd(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	devicesDir := filepath.Join(dir, ".threedoors", "devices")
	if err := os.MkdirAll(devicesDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	reg := device.NewLocalDeviceRegistry(devicesDir)

	id, err := device.NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := device.Device{
		ID:        id,
		Name:      "old-name",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"devices", "rename", "550e8400-e29b-41d4-a716-446655440000", "new-name"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	got, err := reg.Get(id)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if got.Name != "new-name" {
		t.Errorf("Name = %s, want new-name", got.Name)
	}
}

func TestDevicesRenameCmd_NotFound(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"devices", "rename", "550e8400-e29b-41d4-a716-446655440000", "new-name"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should return error for non-existent device")
	}
}
