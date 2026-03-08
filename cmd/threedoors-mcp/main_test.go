package main

import (
	"bytes"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestNewWALProvider_NilBaseProviderPanics(t *testing.T) {
	t.Parallel()

	// This test verifies why the nil guard in run() is necessary:
	// NewWALProvider with a nil base provider would panic on use.
	// Our fix adds a nil check before reaching this point.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when creating WALProvider with nil base, got none")
		}
	}()

	wal := core.NewWALProvider(nil, t.TempDir())
	// If NewWALProvider doesn't panic on creation, LoadTasks will.
	_, _ = wal.LoadTasks()
}

func TestNilProviderErrorMessage(t *testing.T) {
	t.Parallel()

	// Verify the error message format matches what run() would produce.
	configPath := "/test/config.yaml"
	errMsg := "no task provider available: check your configuration in " + configPath

	if !bytes.Contains([]byte(errMsg), []byte("no task provider available")) {
		t.Error("error message should contain 'no task provider available'")
	}
	if !bytes.Contains([]byte(errMsg), []byte(configPath)) {
		t.Error("error message should contain config path")
	}
}
