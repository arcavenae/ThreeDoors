package core

import (
	"fmt"
	"io/fs"
	"log"
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
// If the directory exists with more permissive permissions, it tightens them.
// It validates the directory is not a symlink and is owned by the current user.
func EnsureConfigDir() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}

	// Validate before creating — reject symlinked paths
	if err := ValidateDir(configPath); err != nil {
		return "", fmt.Errorf("config directory validation failed: %w", err)
	}

	if err := os.MkdirAll(configPath, 0o700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	migrateDirectoryPermissions(configPath)

	// Re-validate after creation to check ownership
	if err := ValidateDir(configPath); err != nil {
		return "", fmt.Errorf("config directory validation failed: %w", err)
	}

	return configPath, nil
}

// migrateDirectoryPermissions tightens the config directory to 0o700
// if it has more permissive permissions (e.g. 0o755 from older versions).
func migrateDirectoryPermissions(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	perm := info.Mode().Perm()
	if perm&(fs.FileMode(0o077)) != 0 {
		if err := os.Chmod(path, 0o700); err != nil {
			log.Printf("Warning: failed to tighten directory permissions on %s: %v", path, err)
			return
		}
		log.Printf("Migrated directory permissions on %s from %o to 0700", path, perm)
	}
}

// GetTasksFilePath returns the path to tasks.yaml.
func GetTasksFilePath() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, "tasks.yaml"), nil
}
