package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	valuesConfigFile = "config.yaml"
	maxValues        = 5
	maxValueLength   = 200
)

// ValuesConfig holds user-defined values and goals.
type ValuesConfig struct {
	Values []string `yaml:"values"`
}

// LoadValuesConfig reads values configuration from a YAML file.
// Returns an empty config if the file does not exist.
func LoadValuesConfig(path string) (*ValuesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ValuesConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read values config: %w", err)
	}

	if len(data) == 0 {
		return &ValuesConfig{}, nil
	}

	cfg := &ValuesConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse values config: %w", err)
	}

	return cfg, nil
}

// SaveValuesConfig persists values configuration to a YAML file using atomic write.
func SaveValuesConfig(path string, cfg *ValuesConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal values config: %w", err)
	}

	tmpPath := path + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	_ = f.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

// GetValuesConfigPath returns the path to config.yaml.
func GetValuesConfigPath() (string, error) {
	configPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, valuesConfigFile), nil
}

// HasValues returns true if the config has at least one value.
func (vc *ValuesConfig) HasValues() bool {
	return len(vc.Values) > 0
}

// AddValue appends a value if under the limit.
func (vc *ValuesConfig) AddValue(value string) error {
	if len(vc.Values) >= maxValues {
		return fmt.Errorf("cannot add more than %d values", maxValues)
	}
	if len(value) > maxValueLength {
		return fmt.Errorf("value exceeds maximum length of %d characters", maxValueLength)
	}
	if value == "" {
		return errors.New("value cannot be empty")
	}
	vc.Values = append(vc.Values, value)
	return nil
}

// RemoveValue removes a value by index.
func (vc *ValuesConfig) RemoveValue(index int) error {
	if index < 0 || index >= len(vc.Values) {
		return fmt.Errorf("index %d out of range", index)
	}
	vc.Values = append(vc.Values[:index], vc.Values[index+1:]...)
	return nil
}

// MoveUp moves a value up by one position.
func (vc *ValuesConfig) MoveUp(index int) error {
	if index <= 0 || index >= len(vc.Values) {
		return fmt.Errorf("cannot move index %d up", index)
	}
	vc.Values[index], vc.Values[index-1] = vc.Values[index-1], vc.Values[index]
	return nil
}

// MoveDown moves a value down by one position.
func (vc *ValuesConfig) MoveDown(index int) error {
	if index < 0 || index >= len(vc.Values)-1 {
		return fmt.Errorf("cannot move index %d down", index)
	}
	vc.Values[index], vc.Values[index+1] = vc.Values[index+1], vc.Values[index]
	return nil
}
