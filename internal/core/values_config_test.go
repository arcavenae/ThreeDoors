package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValuesConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadValuesConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.HasValues() {
		t.Error("expected empty values for missing file")
	}
}

func TestLoadValuesConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadValuesConfig(path)
	if err != nil {
		t.Fatalf("expected no error for empty file, got: %v", err)
	}
	if cfg.HasValues() {
		t.Error("expected empty values for empty file")
	}
}

func TestLoadValuesConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "values:\n  - Health\n  - Family\n  - Growth\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadValuesConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(cfg.Values))
	}
	if cfg.Values[0] != "Health" {
		t.Errorf("expected first value 'Health', got '%s'", cfg.Values[0])
	}
}

func TestLoadValuesConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadValuesConfig(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSaveValuesConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &ValuesConfig{Values: []string{"Health", "Family"}}
	if err := SaveValuesConfig(path, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := LoadValuesConfig(path)
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if len(loaded.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(loaded.Values))
	}
	if loaded.Values[0] != "Health" || loaded.Values[1] != "Family" {
		t.Errorf("values mismatch: %v", loaded.Values)
	}
}

func TestValuesConfig_HasValues(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected bool
	}{
		{"empty", nil, false},
		{"with values", []string{"Health"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ValuesConfig{Values: tt.values}
			if cfg.HasValues() != tt.expected {
				t.Errorf("HasValues() = %v, want %v", cfg.HasValues(), tt.expected)
			}
		})
	}
}

func TestValuesConfig_AddValue(t *testing.T) {
	cfg := &ValuesConfig{}

	if err := cfg.AddValue("Health"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(cfg.Values))
	}

	// Test empty value
	if err := cfg.AddValue(""); err == nil {
		t.Error("expected error for empty value")
	}

	// Fill to max
	for i := len(cfg.Values); i < maxValues; i++ {
		if err := cfg.AddValue("Value"); err != nil {
			t.Fatalf("unexpected error adding value %d: %v", i, err)
		}
	}

	// Test max exceeded
	if err := cfg.AddValue("One too many"); err == nil {
		t.Error("expected error when exceeding max values")
	}
}

func TestValuesConfig_RemoveValue(t *testing.T) {
	cfg := &ValuesConfig{Values: []string{"A", "B", "C"}}

	if err := cfg.RemoveValue(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(cfg.Values))
	}
	if cfg.Values[0] != "A" || cfg.Values[1] != "C" {
		t.Errorf("unexpected values: %v", cfg.Values)
	}

	// Out of range
	if err := cfg.RemoveValue(5); err == nil {
		t.Error("expected error for out of range index")
	}
}

func TestValuesConfig_MoveUp(t *testing.T) {
	cfg := &ValuesConfig{Values: []string{"A", "B", "C"}}

	if err := cfg.MoveUp(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Values[0] != "B" || cfg.Values[1] != "A" {
		t.Errorf("unexpected order: %v", cfg.Values)
	}

	// Cannot move first element up
	if err := cfg.MoveUp(0); err == nil {
		t.Error("expected error for moving first element up")
	}
}

func TestValuesConfig_MoveDown(t *testing.T) {
	cfg := &ValuesConfig{Values: []string{"A", "B", "C"}}

	if err := cfg.MoveDown(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Values[0] != "B" || cfg.Values[1] != "A" {
		t.Errorf("unexpected order: %v", cfg.Values)
	}

	// Cannot move last element down
	if err := cfg.MoveDown(2); err == nil {
		t.Error("expected error for moving last element down")
	}
}
