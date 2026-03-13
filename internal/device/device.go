package device

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Device holds identity metadata for this ThreeDoors installation.
type Device struct {
	ID        DeviceID  `yaml:"id"`
	Name      string    `yaml:"name"`
	FirstSeen time.Time `yaml:"first_seen"`
	LastSync  time.Time `yaml:"last_sync"`
}

// DefaultDeviceName returns the hostname as the default device name.
func DefaultDeviceName() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

// SaveDevice writes device metadata to the given path using atomic write.
func SaveDevice(dev *Device, path string) error {
	data, err := yaml.Marshal(dev)
	if err != nil {
		return fmt.Errorf("marshal device: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create device dir: %w", err)
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

// LoadDevice reads device metadata from the given path.
func LoadDevice(path string) (*Device, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read device file: %w", err)
	}

	var dev Device
	if err := yaml.Unmarshal(data, &dev); err != nil {
		return nil, fmt.Errorf("unmarshal device: %w", err)
	}

	return &dev, nil
}
