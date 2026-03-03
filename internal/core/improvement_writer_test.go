package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteImprovement(t *testing.T) {
	dir := t.TempDir()

	err := WriteImprovement(dir, "test-session-123", "Break tasks into smaller pieces")
	if err != nil {
		t.Fatalf("WriteImprovement failed: %v", err)
	}

	path := filepath.Join(dir, "improvements.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading improvements file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "(test-session-123)") {
		t.Errorf("expected session ID in output, got: %s", content)
	}
	if !strings.Contains(content, "Break tasks into smaller pieces") {
		t.Errorf("expected improvement text in output, got: %s", content)
	}
	if !strings.HasSuffix(content, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestWriteImprovement_AppendsMultiple(t *testing.T) {
	dir := t.TempDir()

	err := WriteImprovement(dir, "session-1", "First improvement")
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	err = WriteImprovement(dir, "session-2", "Second improvement")
	if err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	path := filepath.Join(dir, "improvements.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading improvements file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %s", len(lines), string(data))
	}
}

func TestWriteImprovement_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "improvements.txt")

	// File should not exist yet
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("file should not exist before write")
	}

	err := WriteImprovement(dir, "session-1", "test")
	if err != nil {
		t.Fatalf("WriteImprovement failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file should exist after write")
	}
}

func TestWriteImprovement_TimestampFormat(t *testing.T) {
	dir := t.TempDir()

	err := WriteImprovement(dir, "session-1", "test improvement")
	if err != nil {
		t.Fatalf("WriteImprovement failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "improvements.txt"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	content := string(data)
	// Should start with [YYYY-MM-DD HH:MM:SS]
	if !strings.HasPrefix(content, "[") {
		t.Errorf("expected timestamp prefix, got: %s", content)
	}
	if !strings.Contains(content, "] (session-1) test improvement") {
		t.Errorf("expected formatted line, got: %s", content)
	}
}
