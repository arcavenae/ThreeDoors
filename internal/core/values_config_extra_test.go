package core

import "testing"

func TestGetValuesConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })

	path, err := GetValuesConfigPath()
	if err != nil {
		t.Fatalf("GetValuesConfigPath: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}
