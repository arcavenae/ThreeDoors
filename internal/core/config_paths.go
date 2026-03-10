package core

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDir = ".threedoors"
)

var testHomeDir string

// SetHomeDir sets the home directory for testing purposes.
func SetHomeDir(dir string) {
	testHomeDir = dir
}

// GetConfigDirPath returns the path to ~/.threedoors/.
func GetConfigDirPath() (string, error) {
	var homeDir string
	if testHomeDir != "" {
		homeDir = testHomeDir
	} else {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
	}
	return filepath.Join(homeDir, configDir), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist.
// Uses 0o700 permissions so other local users cannot read task data.
// If the directory already exists with more permissive mode, it is tightened.
func EnsureConfigDir() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(configPath, 0o700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := tightenDirPermissions(configPath); err != nil {
		return "", err
	}
	return configPath, nil
}

// tightenDirPermissions checks if a directory is more permissive than 0o700
// and tightens it. This handles upgrades from older versions that used 0o755.
func tightenDirPermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat config directory: %w", err)
	}
	mode := info.Mode().Perm()
	if mode&0o077 != 0 {
		if err := os.Chmod(path, 0o700); err != nil {
			return fmt.Errorf("tighten config directory permissions: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Tightened %s permissions from %04o to 0700\n", path, mode)
	}
	return nil
}

// GetTasksFilePath returns the path to tasks.yaml.
func GetTasksFilePath() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, "tasks.yaml"), nil
}
