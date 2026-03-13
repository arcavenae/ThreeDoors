package device

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GetOrCreateDevice loads the existing device identity or creates a new one.
// On first run, it generates a device ID, persists it, and registers it locally.
func GetOrCreateDevice(configDir string) (*Device, error) {
	devicePath := filepath.Join(configDir, "device.yaml")

	dev, err := LoadDevice(devicePath)
	if err == nil {
		return dev, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		// File exists but is corrupted — back it up and regenerate
		backupPath := devicePath + ".bak"
		_ = os.Rename(devicePath, backupPath)
	}

	reader := NewPlatformMachineIDReader()
	id, err := GenerateDeviceID(reader, configDir)
	if err != nil {
		return nil, fmt.Errorf("generate device ID: %w", err)
	}

	now := time.Now().UTC()
	dev = &Device{
		ID:        id,
		Name:      DefaultDeviceName(),
		FirstSeen: now,
		LastSync:  now,
	}

	if err := SaveDevice(dev, devicePath); err != nil {
		return nil, fmt.Errorf("save device: %w", err)
	}

	// Register in local registry
	devicesDir := filepath.Join(configDir, "devices")
	reg := NewLocalDeviceRegistry(devicesDir)
	if regErr := reg.Register(*dev); regErr != nil {
		if !errors.Is(regErr, ErrDeviceAlreadyExists) {
			return nil, fmt.Errorf("register device: %w", regErr)
		}
	}

	return dev, nil
}
