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
func EnsureConfigDir() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return configPath, nil
}

// GetTasksFilePath returns the path to tasks.yaml.
func GetTasksFilePath() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, "tasks.yaml"), nil
}
