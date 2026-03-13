package device

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DeviceRegistry manages device registration and discovery.
type DeviceRegistry interface {
	Register(device Device) error
	Get(id DeviceID) (Device, error)
	List() ([]Device, error)
	Update(device Device) error
	Remove(id DeviceID) error
}

// LocalDeviceRegistry stores device entries as YAML files in a local directory.
type LocalDeviceRegistry struct {
	dir string
}

// NewLocalDeviceRegistry creates a registry backed by the given directory.
func NewLocalDeviceRegistry(dir string) *LocalDeviceRegistry {
	return &LocalDeviceRegistry{dir: dir}
}

func (r *LocalDeviceRegistry) devicePath(id DeviceID) string {
	return filepath.Join(r.dir, id.String()+".yaml")
}

func (r *LocalDeviceRegistry) ensureDir() error {
	return os.MkdirAll(r.dir, 0o700)
}

// Register adds a new device to the registry.
func (r *LocalDeviceRegistry) Register(dev Device) error {
	if err := r.ensureDir(); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	path := r.devicePath(dev.ID)

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrDeviceAlreadyExists, dev.ID)
	}

	return r.writeDevice(dev, path)
}

// Get retrieves a device by ID.
func (r *LocalDeviceRegistry) Get(id DeviceID) (Device, error) {
	path := r.devicePath(id)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Device{}, fmt.Errorf("%w: %s", ErrDeviceNotFound, id)
		}
		return Device{}, fmt.Errorf("read device %s: %w", id, err)
	}

	var dev Device
	if err := yaml.Unmarshal(data, &dev); err != nil {
		return Device{}, fmt.Errorf("unmarshal device %s: %w", id, err)
	}

	return dev, nil
}

// List returns all registered devices.
func (r *LocalDeviceRegistry) List() ([]Device, error) {
	if err := r.ensureDir(); err != nil {
		return nil, fmt.Errorf("create registry dir: %w", err)
	}

	entries, err := os.ReadDir(r.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Device{}, nil
		}
		return nil, fmt.Errorf("read registry dir: %w", err)
	}

	devices := []Device{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(r.dir, entry.Name()))
		if err != nil {
			continue
		}

		var dev Device
		if err := yaml.Unmarshal(data, &dev); err != nil {
			continue
		}

		devices = append(devices, dev)
	}

	return devices, nil
}

// Update overwrites an existing device entry.
func (r *LocalDeviceRegistry) Update(dev Device) error {
	path := r.devicePath(dev.ID)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrDeviceNotFound, dev.ID)
	}

	return r.writeDevice(dev, path)
}

// Remove deletes a device from the registry.
func (r *LocalDeviceRegistry) Remove(id DeviceID) error {
	path := r.devicePath(id)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrDeviceNotFound, id)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove device %s: %w", id, err)
	}

	return nil
}

// writeDevice atomically writes device data to a YAML file.
func (r *LocalDeviceRegistry) writeDevice(dev Device, path string) error {
	data, err := yaml.Marshal(dev)
	if err != nil {
		return fmt.Errorf("marshal device: %w", err)
	}

	tmpPath := path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	writeErr := func() error {
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("write device data: %w", err)
		}
		if err := f.Sync(); err != nil {
			return fmt.Errorf("sync device data: %w", err)
		}
		return f.Close()
	}()

	if writeErr != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return writeErr
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename device file: %w", err)
	}

	return nil
}
