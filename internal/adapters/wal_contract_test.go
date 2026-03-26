package adapters_test

import (
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/adapters"
	"github.com/arcavenae/ThreeDoors/internal/adapters/textfile"
	"github.com/arcavenae/ThreeDoors/internal/core"
)

// TestWALProviderContract runs the full contract test suite against
// WALProvider wrapping a TextFileProvider. When the inner provider
// succeeds, WAL should be transparent — all operations pass through.
func TestWALProviderContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()

		dir := t.TempDir()
		core.SetHomeDir(dir)

		inner := textfile.NewTextFileProvider()
		walProvider := core.NewWALProvider(inner, dir)

		t.Cleanup(func() {
			core.SetHomeDir("")
		})

		return walProvider
	}

	adapters.RunContractTests(t, factory)
}

// TestWALProviderContract_PendingCountZero verifies that when the inner
// provider is healthy, no WAL entries accumulate.
func TestWALProviderContract_PendingCountZero(t *testing.T) {
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() {
		core.SetHomeDir("")
	})

	inner := textfile.NewTextFileProvider()
	walProvider := core.NewWALProvider(inner, dir)

	task := core.NewTask("WAL test task")
	if err := walProvider.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	if walProvider.PendingCount() != 0 {
		t.Errorf("PendingCount() = %d, want 0 (inner provider succeeded)", walProvider.PendingCount())
	}
}
