package adapters_test

import (
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/adapters"
	"github.com/arcavenae/ThreeDoors/internal/adapters/textfile"
	"github.com/arcavenae/ThreeDoors/internal/core"
)

// TestFallbackProviderContract runs the full contract test suite against
// FallbackProvider with two TextFileProvider instances (primary and fallback).
func TestFallbackProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()

		// Each provider gets its own temp dir for isolation
		primaryDir := t.TempDir()
		core.SetHomeDir(primaryDir)

		primary := textfile.NewTextFileProvider()
		fallback := textfile.NewTextFileProvider()

		t.Cleanup(func() {
			core.SetHomeDir("")
		})

		return core.NewFallbackProvider(primary, fallback)
	}

	adapters.RunContractTests(t, factory)
}

// TestFallbackProviderContract_FallbackActivation verifies that when the
// primary provider fails on LoadTasks, the FallbackProvider switches to
// the fallback and reports IsFallback() == true.
func TestFallbackProviderContract_FallbackActivation(t *testing.T) {
	primaryDir := t.TempDir()
	fallbackDir := t.TempDir()

	// Set up the fallback provider with real data
	core.SetHomeDir(fallbackDir)
	fallbackProvider := textfile.NewTextFileProvider()
	seed := []*core.Task{core.NewTask("Fallback task")}
	if err := fallbackProvider.SaveTasks(seed); err != nil {
		t.Fatalf("seed fallback: %v", err)
	}

	// Primary points to an empty but valid dir — set home to primary dir
	core.SetHomeDir(primaryDir)
	primaryProvider := textfile.NewTextFileProvider()

	t.Cleanup(func() {
		core.SetHomeDir("")
	})

	fp := core.NewFallbackProvider(primaryProvider, fallbackProvider)

	// Primary should succeed (returns default or empty tasks)
	_, err := fp.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// FallbackProvider should NOT have activated fallback since primary succeeded
	if fp.IsFallback() {
		t.Error("IsFallback() = true, want false (primary succeeded)")
	}
}
